# ADR-007: Kubernetes for Container Orchestration

## Status
Accepted

## Date
2024-02-01

## Context

Our microservices-based ERP system requires a robust container orchestration platform to:
- Deploy and manage multiple containerized services
- Ensure high availability and fault tolerance
- Enable auto-scaling based on load
- Provide service discovery and load balancing
- Support rolling updates and rollbacks
- Manage configuration and secrets
- Monitor and maintain health of services

We evaluated several orchestration platforms:

| Platform | Pros | Cons |
|----------|------|------|
| **Kubernetes** | Industry standard, feature-rich, large ecosystem | Complex, steep learning curve |
| **Docker Swarm** | Simple, built into Docker | Limited features, smaller ecosystem |
| **AWS ECS** | Managed, AWS-native | Vendor lock-in, limited portability |
| **Nomad** | Simple, flexible | Smaller ecosystem, less mature |

## Decision

We will use **Kubernetes** as our container orchestration platform.

### Implementation Strategy

1. **Cluster Architecture**
   ```
   Production Cluster:
   - 3 Master nodes (HA control plane)
   - 5+ Worker nodes (application workloads)
   - Separate namespace per environment

   Namespaces:
   - marimo-erp (production)
   - marimo-erp-staging
   - marimo-erp-dev
   ```

2. **Resource Organization**
   ```
   k8s/
   ├── namespace.yml
   ├── configmap.yml
   ├── secrets.example.yml
   ├── postgres-statefulset.yml
   ├── redis-statefulset.yml
   ├── consul-statefulset.yml
   ├── rabbitmq-statefulset.yml
   ├── api-gateway-deployment.yml
   ├── auth-service-deployment.yml
   ├── hpa.yml
   ├── ingress.yml
   └── blue-green/
       ├── api-gateway-blue.yml
       └── api-gateway-green.yml
   ```

3. **Deployment Patterns**

   **StatefulSets** (for stateful apps):
   - PostgreSQL
   - Redis
   - Consul
   - RabbitMQ

   **Deployments** (for stateless apps):
   - API Gateway
   - Auth Service
   - Users Service
   - Other microservices

   **Jobs** (for one-time tasks):
   - Database migrations
   - Data imports

   **CronJobs** (for scheduled tasks):
   - Backups
   - Report generation
   - Cleanup jobs

4. **Resource Limits**
   ```yaml
   resources:
     requests:
       memory: "256Mi"
       cpu: "250m"
     limits:
       memory: "1Gi"
       cpu: "1000m"
   ```

5. **Health Checks**
   ```yaml
   livenessProbe:
     httpGet:
       path: /health
       port: 8080
     initialDelaySeconds: 30
     periodSeconds: 10

   readinessProbe:
     httpGet:
       path: /ready
       port: 8080
     initialDelaySeconds: 5
     periodSeconds: 5
   ```

### Key Features We Use

1. **Auto-scaling (HPA)**
   ```yaml
   apiVersion: autoscaling/v2
   kind: HorizontalPodAutoscaler
   metadata:
     name: api-gateway-hpa
   spec:
     minReplicas: 3
     maxReplicas: 10
     metrics:
       - type: Resource
         resource:
           name: cpu
           target:
             type: Utilization
             averageUtilization: 70
   ```

2. **Service Discovery**
   - Services automatically discoverable via DNS
   - Example: `postgres-service.marimo-erp.svc.cluster.local`

3. **Load Balancing**
   - Built-in service load balancing
   - NGINX Ingress for external traffic

4. **Configuration Management**
   - ConfigMaps for non-sensitive config
   - Secrets for sensitive data
   - Environment variables injected into pods

5. **Storage**
   - PersistentVolumeClaims for databases
   - Dynamic provisioning
   - Snapshot support for backups

## Consequences

### Positive

- **Industry Standard**: Large community, extensive documentation
- **Vendor Neutral**: Run anywhere (cloud, on-prem, hybrid)
- **Rich Ecosystem**: Helm, Istio, Prometheus, etc.
- **Auto-scaling**: HPA for compute, VPA for resources
- **Self-healing**: Automatic restart of failed containers
- **Rolling Updates**: Zero-downtime deployments
- **Service Mesh Ready**: Can add Istio/Linkerd later
- **Declarative**: Infrastructure as Code
- **Extensible**: Custom Resource Definitions (CRDs)

### Negative

- **Complexity**: Steep learning curve for team
- **Operational Overhead**: Cluster management, upgrades
- **Resource Usage**: Master nodes require resources
- **Debugging**: More complex than traditional deployments
- **Network Complexity**: Pod networking, service mesh
- **Cost**: Managed Kubernetes (EKS, GKE, AKS) can be expensive

### Mitigation Strategies

1. **Managed Kubernetes**
   - Use EKS (AWS), GKE (Google), or AKS (Azure)
   - Reduces operational burden
   - Automatic master node management
   - Regular security updates

2. **Training**
   - Team training on Kubernetes fundamentals
   - Certification: CKA/CKAD
   - Internal documentation and runbooks

3. **Monitoring**
   - Prometheus for metrics
   - Grafana for visualization
   - Alert Manager for notifications
   - Jaeger for distributed tracing

4. **Simplified Deployments**
   - Helm charts for complex applications
   - GitOps with ArgoCD/Flux
   - Standardized deployment scripts

## Blue-Green Deployment

```yaml
# Service (switchable)
apiVersion: v1
kind: Service
metadata:
  name: api-gateway-service
spec:
  selector:
    app: api-gateway
    color: blue  # Switch to 'green' for deployment
```

Deployment process:
1. Deploy green version
2. Test green version
3. Switch service selector to green
4. Monitor for issues
5. Rollback if needed (switch back to blue)
6. Remove blue deployment when stable

## Disaster Recovery

1. **Cluster Backup**
   - etcd snapshots (control plane data)
   - Velero for cluster state and volumes
   - Daily backups to S3

2. **Recovery Steps**
   ```bash
   # Restore cluster from backup
   kubectl apply -f k8s/namespace.yml
   kubectl apply -f k8s/secrets.yml
   velero restore create --from-backup daily-backup-20240201
   ```

## Observability

1. **Metrics**
   - Prometheus metrics from all pods
   - Grafana dashboards
   - Custom metrics via Prometheus client

2. **Logging**
   - Centralized logging with EFK stack
   - (Elasticsearch, Fluentd, Kibana)
   - Log retention: 30 days

3. **Tracing**
   - Jaeger for distributed tracing
   - Trace all cross-service requests

## Security

1. **RBAC** (Role-Based Access Control)
   - Separate service accounts per service
   - Least privilege principle
   - Audit logging enabled

2. **Network Policies**
   ```yaml
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: api-gateway-policy
   spec:
     podSelector:
       matchLabels:
         app: api-gateway
     ingress:
       - from:
           - podSelector:
               matchLabels:
                 app: frontend
   ```

3. **Pod Security**
   - Run as non-root user
   - Read-only root filesystem
   - Drop all capabilities

4. **Secrets Management**
   - Kubernetes Secrets (base64)
   - Consider Vault for production
   - Sealed Secrets for GitOps

## Cost Optimization

1. **Resource Requests**
   - Right-size pod resources
   - Use VPA for recommendations

2. **Auto-scaling**
   - Scale down during off-hours
   - Cluster autoscaler for nodes

3. **Spot Instances**
   - Use spot/preemptible instances for dev/staging
   - Mix of on-demand and spot for production

## Alternatives Considered

### Docker Swarm
**Rejected** because:
- Limited features compared to K8s
- Smaller ecosystem
- Less industry adoption
- Future uncertain

### AWS ECS
**Rejected** because:
- Vendor lock-in
- Limited portability
- Feature set smaller than K8s
- Team wants cloud-agnostic solution

### Manual VMs
**Rejected** because:
- No auto-scaling
- Manual orchestration
- Poor resource utilization
- Complex deployments

## Migration Path

### Phase 1: Initial Deployment (Completed)
- ✅ Create Kubernetes manifests
- ✅ Deploy to development cluster
- ✅ Implement CI/CD pipeline

### Phase 2: Production Readiness (In Progress)
- ✅ Blue-green deployment
- ✅ Monitoring and alerting
- ✅ Backup and disaster recovery
- ⏳ Load testing
- ⏳ Security hardening

### Phase 3: Advanced Features (Future)
- Service mesh (Istio)
- GitOps (ArgoCD)
- Advanced networking (Calico)
- Policy management (OPA)

## Related Decisions

- [ADR-001](./001-microservices-architecture.md) - Microservices Architecture
- [ADR-010](./010-blue-green-deployment.md) - Blue-Green Deployment
- [ADR-009](./009-prometheus-monitoring.md) - Prometheus Monitoring

## References

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Kubernetes Patterns](https://www.redhat.com/en/resources/oreilly-kubernetes-patterns-cloud-native-apps)
- [Production Best Practices](https://kubernetes.io/docs/setup/best-practices/)
- [CNCF Landscape](https://landscape.cncf.io/)
