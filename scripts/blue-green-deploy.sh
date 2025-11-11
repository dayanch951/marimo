#!/bin/bash

# Blue-Green deployment script for Kubernetes
# Usage: ./blue-green-deploy.sh <service-name> <new-image> <namespace>

set -e

SERVICE_NAME=$1
NEW_IMAGE=$2
NAMESPACE=${3:-marimo-erp}

if [ -z "$SERVICE_NAME" ] || [ -z "$NEW_IMAGE" ]; then
    echo "Usage: $0 <service-name> <new-image> [namespace]"
    echo "Example: $0 api-gateway ghcr.io/org/api-gateway:v1.2.3 marimo-erp"
    exit 1
fi

echo "=========================================="
echo "Blue-Green Deployment"
echo "Service: $SERVICE_NAME"
echo "Image: $NEW_IMAGE"
echo "Namespace: $NAMESPACE"
echo "=========================================="

# Check current color
CURRENT_COLOR=$(kubectl get service ${SERVICE_NAME}-service -n $NAMESPACE -o jsonpath='{.spec.selector.color}' 2>/dev/null || echo "blue")
NEW_COLOR="green"

if [ "$CURRENT_COLOR" = "green" ]; then
    NEW_COLOR="blue"
fi

echo "Current color: $CURRENT_COLOR"
echo "New color: $NEW_COLOR"

# Step 1: Deploy new version with new color
echo ""
echo "Step 1: Deploying $NEW_COLOR version..."
cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${SERVICE_NAME}-${NEW_COLOR}
  namespace: $NAMESPACE
  labels:
    app: $SERVICE_NAME
    color: $NEW_COLOR
spec:
  replicas: 3
  selector:
    matchLabels:
      app: $SERVICE_NAME
      color: $NEW_COLOR
  template:
    metadata:
      labels:
        app: $SERVICE_NAME
        color: $NEW_COLOR
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      containers:
        - name: $SERVICE_NAME
          image: $NEW_IMAGE
          imagePullPolicy: Always
          ports:
            - containerPort: 8080
              name: http
          envFrom:
            - configMapRef:
                name: marimo-config
            - secretRef:
                name: marimo-secrets
          resources:
            requests:
              memory: "128Mi"
              cpu: "100m"
            limits:
              memory: "512Mi"
              cpu: "500m"
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 30
            periodSeconds: 10
            timeoutSeconds: 5
            failureThreshold: 3
          readinessProbe:
            httpGet:
              path: /ready
              port: 8080
            initialDelaySeconds: 10
            periodSeconds: 5
            timeoutSeconds: 3
            failureThreshold: 3
EOF

# Step 2: Wait for new version to be ready
echo ""
echo "Step 2: Waiting for $NEW_COLOR deployment to be ready..."
kubectl wait --for=condition=available --timeout=300s \
    deployment/${SERVICE_NAME}-${NEW_COLOR} -n $NAMESPACE

# Step 3: Test new version
echo ""
echo "Step 3: Testing $NEW_COLOR version..."

# Port forward for testing
kubectl port-forward -n $NAMESPACE \
    deployment/${SERVICE_NAME}-${NEW_COLOR} 9090:8080 &
PORT_FORWARD_PID=$!
sleep 5

# Health check
if curl -f -s http://localhost:9090/health > /dev/null; then
    echo "✓ Health check passed"
else
    echo "✗ Health check failed"
    kill $PORT_FORWARD_PID 2>/dev/null || true
    exit 1
fi

kill $PORT_FORWARD_PID 2>/dev/null || true

# Step 4: Switch traffic
echo ""
echo "Step 4: Switching traffic to $NEW_COLOR..."
read -p "Continue with traffic switch? (yes/no): " CONFIRM
if [ "$CONFIRM" != "yes" ]; then
    echo "Deployment cancelled. Cleaning up $NEW_COLOR deployment..."
    kubectl delete deployment ${SERVICE_NAME}-${NEW_COLOR} -n $NAMESPACE
    exit 0
fi

kubectl patch service ${SERVICE_NAME}-service -n $NAMESPACE \
    -p "{\"spec\":{\"selector\":{\"app\":\"${SERVICE_NAME}\",\"color\":\"${NEW_COLOR}\"}}}"

echo "✓ Traffic switched to $NEW_COLOR"

# Step 5: Monitor new version
echo ""
echo "Step 5: Monitoring $NEW_COLOR version for 60 seconds..."
sleep 60

# Check if pods are healthy
READY_PODS=$(kubectl get deployment ${SERVICE_NAME}-${NEW_COLOR} -n $NAMESPACE -o jsonpath='{.status.readyReplicas}')
DESIRED_PODS=$(kubectl get deployment ${SERVICE_NAME}-${NEW_COLOR} -n $NAMESPACE -o jsonpath='{.spec.replicas}')

if [ "$READY_PODS" != "$DESIRED_PODS" ]; then
    echo "✗ Not all pods are ready ($READY_PODS/$DESIRED_PODS)"
    echo "Rolling back to $CURRENT_COLOR..."
    kubectl patch service ${SERVICE_NAME}-service -n $NAMESPACE \
        -p "{\"spec\":{\"selector\":{\"app\":\"${SERVICE_NAME}\",\"color\":\"${CURRENT_COLOR}\"}}}"
    exit 1
fi

echo "✓ All pods are healthy"

# Step 6: Cleanup old version
echo ""
echo "Step 6: Cleaning up $CURRENT_COLOR deployment..."
read -p "Delete old $CURRENT_COLOR deployment? (yes/no): " CONFIRM_DELETE
if [ "$CONFIRM_DELETE" = "yes" ]; then
    kubectl delete deployment ${SERVICE_NAME}-${CURRENT_COLOR} -n $NAMESPACE || true
    echo "✓ Old deployment deleted"
else
    echo "Keeping old deployment for manual cleanup"
fi

echo ""
echo "=========================================="
echo "✓ Deployment completed successfully!"
echo "Active version: $NEW_COLOR"
echo "=========================================="
echo ""
echo "To rollback, run:"
echo "kubectl patch service ${SERVICE_NAME}-service -n $NAMESPACE -p '{\"spec\":{\"selector\":{\"color\":\"${CURRENT_COLOR}\"}}}'"
