# Production Deployment Guide

Полное руководство по развертыванию Marimo ERP в production окружении.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Kubernetes Setup](#kubernetes-setup)
- [Secrets Configuration](#secrets-configuration)
- [Deployment Process](#deployment-process)
- [Blue-Green Deployment](#blue-green-deployment)
- [CI/CD Pipeline](#cicd-pipeline)
- [Monitoring](#monitoring)
- [Scaling](#scaling)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### Required Tools

```bash
# Kubernetes CLI
kubectl version --client

# Docker
docker --version

# Helm (optional)
helm version
```

### Kubernetes Cluster

- Kubernetes 1.25+
- Minimum 3 worker nodes
- Storage class for persistent volumes
- Ingress controller (nginx recommended)
- Cert-manager for TLS certificates

### Resource Requirements

**Minimum per environment:**
- CPU: 8 cores
- Memory: 16 GB RAM
- Storage: 100 GB SSD

**Recommended production:**
- CPU: 16 cores
- Memory: 32 GB RAM
- Storage: 500 GB SSD

## Kubernetes Setup

### 1. Create Namespace

```bash
kubectl create namespace marimo-erp
kubectl label namespace marimo-erp environment=production
```

### 2. Setup Secrets

⚠️ **IMPORTANT**: Never commit secrets to git!

```bash
# Create secrets from file
kubectl create secret generic marimo-secrets \
  --from-literal=DB_USER=postgres \
  --from-literal=DB_PASSWORD=$(openssl rand -base64 32) \
  --from-literal=JWT_SECRET=$(openssl rand -base64 64) \
  --from-literal=REDIS_PASSWORD=$(openssl rand -base64 32) \
  --from-literal=RABBITMQ_USER=admin \
  --from-literal=RABBITMQ_PASSWORD=$(openssl rand -base64 32) \
  --from-literal=SMTP_USER=your-email@gmail.com \
  --from-literal=SMTP_PASSWORD=your-app-password \
  --from-literal=MINIO_ACCESS_KEY=$(openssl rand -base64 20) \
  --from-literal=MINIO_SECRET_KEY=$(openssl rand -base64 40) \
  -n marimo-erp

# Verify secrets
kubectl get secrets -n marimo-erp
```

### 3. Apply ConfigMap

```bash
# Review and customize configuration
vim k8s/configmap.yml

# Apply
kubectl apply -f k8s/configmap.yml
```

### 4. Setup Storage

```bash
# Create PersistentVolumeClaims
kubectl apply -f k8s/postgres-statefulset.yml
kubectl apply -f k8s/consul-statefulset.yml
kubectl apply -f k8s/rabbitmq-statefulset.yml
kubectl apply -f k8s/backup-cronjob.yml
```

## Deployment Process

### Option 1: Manual Deployment

```bash
# Deploy infrastructure services
kubectl apply -f k8s/postgres-statefulset.yml
kubectl apply -f k8s/redis-deployment.yml
kubectl apply -f k8s/consul-statefulset.yml
kubectl apply -f k8s/rabbitmq-statefulset.yml

# Wait for infrastructure to be ready
kubectl wait --for=condition=ready pod -l app=postgres -n marimo-erp --timeout=300s
kubectl wait --for=condition=ready pod -l app=redis -n marimo-erp --timeout=300s
kubectl wait --for=condition=ready pod -l app=consul -n marimo-erp --timeout=300s
kubectl wait --for=condition=ready pod -l app=rabbitmq -n marimo-erp --timeout=300s

# Deploy application services
kubectl apply -f k8s/api-gateway-deployment.yml
kubectl apply -f k8s/auth-service-deployment.yml

# Wait for applications to be ready
kubectl wait --for=condition=available deployment/api-gateway -n marimo-erp --timeout=300s
kubectl wait --for=condition=available deployment/auth-service -n marimo-erp --timeout=300s

# Apply autoscaling
kubectl apply -f k8s/hpa.yml

# Setup ingress
kubectl apply -f k8s/ingress.yml

# Setup automated backups
kubectl apply -f k8s/backup-cronjob.yml
```

### Option 2: Automated Deployment (Disaster Recovery Script)

```bash
./scripts/disaster-recovery.sh marimo-erp [backup-file]
```

## Blue-Green Deployment

Blue-Green deployment позволяет обновлять приложение без даунтайма.

### Automated Script

```bash
# Deploy new version
./scripts/blue-green-deploy.sh api-gateway ghcr.io/org/api-gateway:v1.2.3 marimo-erp

# The script will:
# 1. Deploy green version
# 2. Wait for readiness
# 3. Run health checks
# 4. Ask for confirmation
# 5. Switch traffic
# 6. Monitor for issues
# 7. Cleanup old version
```

### Manual Process

See detailed guide in [k8s/blue-green/README.md](../k8s/blue-green/README.md)

## CI/CD Pipeline

### GitHub Actions Workflows

**CI Pipeline** (`.github/workflows/ci.yml`):
- Runs on every push and pull request
- Backend tests with Go
- Frontend tests with TypeScript
- Security scanning
- Docker build test
- Integration tests

**CD Pipeline** (`.github/workflows/cd.yml`):
- Triggered on main branch push or tags
- Builds and pushes Docker images
- Deploys to staging automatically
- Deploys to production on tags (with approval)
- Blue-green deployment for production
- Automated backups after deployment

### Required GitHub Secrets

```
KUBE_CONFIG_STAGING   # Base64 encoded kubeconfig for staging
KUBE_CONFIG_PROD      # Base64 encoded kubeconfig for production
SLACK_WEBHOOK         # Slack webhook for notifications (optional)
```

### Manual Trigger

```bash
# Trigger deployment via GitHub CLI
gh workflow run cd.yml -f environment=production
```

## Monitoring

### Health Checks

```bash
# Check pod health
kubectl get pods -n marimo-erp

# Check service endpoints
kubectl get endpoints -n marimo-erp

# Test health endpoints
kubectl port-forward -n marimo-erp svc/api-gateway-service 8080:8080
curl http://localhost:8080/health
curl http://localhost:8080/ready
```

### Logs

```bash
# View logs
kubectl logs -n marimo-erp -l app=api-gateway --tail=100 -f

# View logs from all replicas
kubectl logs -n marimo-erp deployment/api-gateway --all-containers --tail=100

# View previous pod logs (after crash)
kubectl logs -n marimo-erp <pod-name> --previous
```

### Metrics

```bash
# Check resource usage
kubectl top pods -n marimo-erp
kubectl top nodes

# Check HPA status
kubectl get hpa -n marimo-erp
kubectl describe hpa api-gateway-hpa -n marimo-erp
```

## Scaling

### Horizontal Pod Autoscaling (HPA)

HPA automatically scales based on CPU and memory usage:

```yaml
# Configured in k8s/hpa.yml
- Min replicas: 3 (api-gateway), 2 (auth-service)
- Max replicas: 10 (api-gateway), 8 (auth-service)
- Target CPU: 70%
- Target Memory: 80%
```

### Manual Scaling

```bash
# Scale deployment manually
kubectl scale deployment api-gateway --replicas=5 -n marimo-erp

# Disable HPA temporarily
kubectl delete hpa api-gateway-hpa -n marimo-erp
```

### Vertical Scaling

Update resource limits in deployment manifests:

```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "250m"
  limits:
    memory: "1Gi"
    cpu: "1000m"
```

## Backup and Restore

### Automated Backups

Backups run daily at 2 AM UTC via CronJob:

```bash
# Check backup status
kubectl get cronjob postgres-backup -n marimo-erp
kubectl get jobs -n marimo-erp

# Manual backup
kubectl create job --from=cronjob/postgres-backup manual-backup-$(date +%s) -n marimo-erp

# View backup logs
kubectl logs -n marimo-erp job/postgres-backup-<job-id>
```

### Manual Backup

```bash
./scripts/backup.sh marimo-erp /backups
```

### Restore

```bash
./scripts/restore.sh /backups/marimo_backup_20240115_120000.dump.gz marimo-erp
```

## Troubleshooting

### Pods not starting

```bash
# Describe pod to see events
kubectl describe pod <pod-name> -n marimo-erp

# Check logs
kubectl logs <pod-name> -n marimo-erp

# Common issues:
# - ImagePullBackOff: Check image name and registry credentials
# - CrashLoopBackOff: Check application logs
# - Pending: Check resource availability
```

### Service not accessible

```bash
# Check service
kubectl get svc -n marimo-erp
kubectl describe svc api-gateway-service -n marimo-erp

# Check endpoints
kubectl get endpoints api-gateway-service -n marimo-erp

# Port forward for testing
kubectl port-forward -n marimo-erp svc/api-gateway-service 8080:8080
```

### Database connection issues

```bash
# Check postgres pod
kubectl get pod -l app=postgres -n marimo-erp

# Test connection from application pod
kubectl exec -it <app-pod> -n marimo-erp -- sh
nc -zv postgres-service 5432

# Check secrets
kubectl get secret marimo-secrets -n marimo-erp -o yaml
```

### High resource usage

```bash
# Check resource usage
kubectl top pods -n marimo-erp --sort-by=memory
kubectl top pods -n marimo-erp --sort-by=cpu

# Check HPA
kubectl describe hpa -n marimo-erp

# Restart high-usage pods
kubectl rollout restart deployment/api-gateway -n marimo-erp
```

### Rollback deployment

```bash
# View rollout history
kubectl rollout history deployment/api-gateway -n marimo-erp

# Rollback to previous version
kubectl rollout undo deployment/api-gateway -n marimo-erp

# Rollback to specific revision
kubectl rollout undo deployment/api-gateway --to-revision=2 -n marimo-erp

# For blue-green, switch back to blue
kubectl patch service api-gateway-service -n marimo-erp \
  -p '{"spec":{"selector":{"color":"blue"}}}'
```

## Security Best Practices

1. **Secrets Management**
   - Use external secret manager (AWS Secrets Manager, Vault)
   - Rotate secrets regularly
   - Never commit secrets to git

2. **Network Policies**
   - Implement network policies to restrict pod communication
   - Use service mesh (Istio, Linkerd) for mTLS

3. **RBAC**
   - Use role-based access control
   - Principle of least privilege

4. **Image Security**
   - Scan images for vulnerabilities
   - Use minimal base images
   - Sign images

5. **TLS/SSL**
   - Use cert-manager for automatic certificate renewal
   - Enforce HTTPS everywhere

## Performance Optimization

1. **Resource Limits**
   - Set appropriate CPU/memory requests and limits
   - Use HPA for automatic scaling

2. **Caching**
   - Redis for application caching
   - CDN for static assets

3. **Database**
   - Connection pooling
   - Query optimization
   - Read replicas for scaling reads

4. **Monitoring**
   - Use Prometheus + Grafana
   - Set up alerts for critical metrics
   - Track response times and error rates

## Additional Resources

- [Blue-Green Deployment Guide](../k8s/blue-green/README.md)
- [Disaster Recovery Plan](./DISASTER_RECOVERY.md)
- [Architecture Documentation](./ARCHITECTURE.md)
- [Feature Documentation](./FEATURES.md)
