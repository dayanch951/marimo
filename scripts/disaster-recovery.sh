#!/bin/bash

# Disaster Recovery Script
# This script helps recover the system after a disaster

set -e

NAMESPACE=${1:-marimo-erp}
BACKUP_FILE=$2

echo "=========================================="
echo "Disaster Recovery Procedure"
echo "Namespace: $NAMESPACE"
echo "=========================================="

# Check if kubectl is configured
if ! kubectl cluster-info > /dev/null 2>&1; then
    echo "✗ Error: kubectl is not configured or cluster is not accessible"
    exit 1
fi

echo "✓ Kubernetes cluster is accessible"

# Step 1: Create namespace if it doesn't exist
echo ""
echo "Step 1: Creating namespace..."
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -
echo "✓ Namespace ready"

# Step 2: Apply secrets (must be done manually for security)
echo ""
echo "Step 2: Applying secrets..."
echo "⚠️  You must manually create secrets before proceeding:"
echo "kubectl create secret generic marimo-secrets \\"
echo "  --from-literal=DB_USER=postgres \\"
echo "  --from-literal=DB_PASSWORD=your-password \\"
echo "  --from-literal=JWT_SECRET=your-jwt-secret \\"
echo "  --from-literal=REDIS_PASSWORD=your-redis-password \\"
echo "  --from-literal=RABBITMQ_USER=admin \\"
echo "  --from-literal=RABBITMQ_PASSWORD=your-rabbitmq-password \\"
echo "  -n $NAMESPACE"
echo ""
read -p "Have you created the secrets? (yes/no): " SECRETS_READY

if [ "$SECRETS_READY" != "yes" ]; then
    echo "Please create secrets first, then run this script again."
    exit 0
fi

# Step 3: Apply ConfigMap
echo ""
echo "Step 3: Applying ConfigMap..."
kubectl apply -f k8s/configmap.yml -n $NAMESPACE
echo "✓ ConfigMap applied"

# Step 4: Deploy infrastructure services
echo ""
echo "Step 4: Deploying infrastructure services..."
kubectl apply -f k8s/postgres-statefulset.yml -n $NAMESPACE
kubectl apply -f k8s/redis-deployment.yml -n $NAMESPACE
kubectl apply -f k8s/consul-statefulset.yml -n $NAMESPACE
kubectl apply -f k8s/rabbitmq-statefulset.yml -n $NAMESPACE

echo "Waiting for infrastructure services to be ready..."
kubectl wait --for=condition=ready pod -l app=postgres -n $NAMESPACE --timeout=300s
kubectl wait --for=condition=ready pod -l app=redis -n $NAMESPACE --timeout=300s
kubectl wait --for=condition=ready pod -l app=consul -n $NAMESPACE --timeout=300s
kubectl wait --for=condition=ready pod -l app=rabbitmq -n $NAMESPACE --timeout=300s

echo "✓ Infrastructure services ready"

# Step 5: Restore database from backup (if provided)
if [ -n "$BACKUP_FILE" ]; then
    echo ""
    echo "Step 5: Restoring database from backup..."
    if [ -f "$BACKUP_FILE" ]; then
        ./scripts/restore.sh "$BACKUP_FILE" "$NAMESPACE"
        echo "✓ Database restored"
    else
        echo "✗ Backup file not found: $BACKUP_FILE"
        echo "Skipping database restore. You'll need to restore manually."
    fi
else
    echo ""
    echo "Step 5: Skipping database restore (no backup file provided)"
    echo "To restore later, run: ./scripts/restore.sh <backup-file> $NAMESPACE"
fi

# Step 6: Deploy application services
echo ""
echo "Step 6: Deploying application services..."
kubectl apply -f k8s/api-gateway-deployment.yml -n $NAMESPACE
kubectl apply -f k8s/auth-service-deployment.yml -n $NAMESPACE

echo "Waiting for application services to be ready..."
kubectl wait --for=condition=available deployment/api-gateway -n $NAMESPACE --timeout=300s
kubectl wait --for=condition=available deployment/auth-service -n $NAMESPACE --timeout=300s

echo "✓ Application services ready"

# Step 7: Apply autoscaling
echo ""
echo "Step 7: Applying autoscaling..."
kubectl apply -f k8s/hpa.yml -n $NAMESPACE
echo "✓ Autoscaling configured"

# Step 8: Apply ingress
echo ""
echo "Step 8: Applying ingress..."
kubectl apply -f k8s/ingress.yml -n $NAMESPACE
echo "✓ Ingress configured"

# Step 9: Setup backup cronjob
echo ""
echo "Step 9: Setting up automated backups..."
kubectl apply -f k8s/backup-cronjob.yml -n $NAMESPACE
echo "✓ Backup cronjob configured"

# Step 10: Verify deployment
echo ""
echo "Step 10: Verifying deployment..."

# Check all pods
echo ""
echo "Pod status:"
kubectl get pods -n $NAMESPACE

# Check services
echo ""
echo "Service status:"
kubectl get services -n $NAMESPACE

# Check ingress
echo ""
echo "Ingress status:"
kubectl get ingress -n $NAMESPACE

# Step 11: Run health checks
echo ""
echo "Step 11: Running health checks..."
sleep 10

API_GATEWAY_POD=$(kubectl get pods -n $NAMESPACE -l app=api-gateway -o jsonpath='{.items[0].metadata.name}')
if [ -n "$API_GATEWAY_POD" ]; then
    kubectl exec -n $NAMESPACE $API_GATEWAY_POD -- wget -q -O- http://localhost:8080/health > /dev/null 2>&1
    if [ $? -eq 0 ]; then
        echo "✓ API Gateway health check passed"
    else
        echo "⚠️  API Gateway health check failed"
    fi
fi

AUTH_SERVICE_POD=$(kubectl get pods -n $NAMESPACE -l app=auth-service -o jsonpath='{.items[0].metadata.name}')
if [ -n "$AUTH_SERVICE_POD" ]; then
    kubectl exec -n $NAMESPACE $AUTH_SERVICE_POD -- wget -q -O- http://localhost:8081/health > /dev/null 2>&1
    if [ $? -eq 0 ]; then
        echo "✓ Auth Service health check passed"
    else
        echo "⚠️  Auth Service health check failed"
    fi
fi

echo ""
echo "=========================================="
echo "✓ Disaster Recovery Completed!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "1. Verify all services are working correctly"
echo "2. Check application logs: kubectl logs -n $NAMESPACE -l app=api-gateway"
echo "3. Update DNS if necessary"
echo "4. Run full system tests"
echo "5. Monitor metrics and alerts"
echo ""
echo "To access services:"
echo "kubectl port-forward -n $NAMESPACE svc/api-gateway-service 8080:8080"
echo ""
