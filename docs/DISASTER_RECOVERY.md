# Disaster Recovery Plan

Comprehensive disaster recovery plan for Marimo ERP system.

## Table of Contents

- [Overview](#overview)
- [Recovery Objectives](#recovery-objectives)
- [Backup Strategy](#backup-strategy)
- [Recovery Procedures](#recovery-procedures)
- [Failure Scenarios](#failure-scenarios)
- [Testing](#testing)
- [Responsibilities](#responsibilities)

## Overview

This disaster recovery (DR) plan outlines procedures to recover from various failure scenarios and minimize downtime and data loss.

### Scope

- Database recovery
- Application service recovery
- Infrastructure recovery
- Complete datacenter failover

### Key Definitions

- **RTO (Recovery Time Objective)**: Maximum acceptable downtime
- **RPO (Recovery Point Objective)**: Maximum acceptable data loss
- **MTTR (Mean Time To Recovery)**: Average time to recover from failure

## Recovery Objectives

| Component | RTO | RPO | Priority |
|-----------|-----|-----|----------|
| Database | 1 hour | 1 hour | Critical |
| Auth Service | 15 minutes | 0 | Critical |
| API Gateway | 15 minutes | 0 | Critical |
| Redis Cache | 10 minutes | Acceptable loss | High |
| RabbitMQ | 30 minutes | Acceptable loss | Medium |
| Consul | 30 minutes | Acceptable loss | Medium |

## Backup Strategy

### Database Backups

**Automated Daily Backups:**
- Schedule: Daily at 2:00 AM UTC
- Retention: 30 days local, 90 days in S3
- Format: PostgreSQL custom format (compressed)
- Location:
  - Primary: Kubernetes PVC
  - Secondary: AWS S3 / Google Cloud Storage
  - Tertiary: Off-site storage

**Backup Verification:**
```bash
# Verify backup integrity
cd /backups
sha256sum -c marimo_backup_20240115_120000.sha256

# Test restore to temporary database
./scripts/restore.sh /backups/marimo_backup_20240115_120000.dump.gz test-namespace
```

### Application State

**StatefulSet Data:**
- Consul: Configuration and service registry
- RabbitMQ: Message queues and definitions
- Retention: 7 days

**Configuration Backups:**
- Kubernetes manifests in git
- ConfigMaps and Secrets (encrypted)
- Environment configurations

### Point-in-Time Recovery

For critical data, enable PostgreSQL WAL archiving:

```yaml
# postgres-statefulset.yml
args:
  - -c
  - wal_level=replica
  - -c
  - archive_mode=on
  - -c
  - archive_command='cp %p /archive/%f'
```

## Recovery Procedures

### 1. Database Recovery

#### Scenario: Database corruption or data loss

**Recovery Steps:**

```bash
# 1. List available backups
ls -lh /backups/marimo_backup_*.dump.gz

# 2. Verify backup integrity
sha256sum -c /backups/marimo_backup_20240115_120000.sha256

# 3. Stop application services
kubectl scale deployment --all --replicas=0 -n marimo-erp --selector='app!=postgres'

# 4. Restore database
./scripts/restore.sh /backups/marimo_backup_20240115_120000.dump.gz marimo-erp

# 5. Verify restoration
kubectl exec -n marimo-erp postgres-0 -- psql -U postgres -d marimo_erp -c "SELECT COUNT(*) FROM users;"

# 6. Restart application services
kubectl scale deployment api-gateway --replicas=3 -n marimo-erp
kubectl scale deployment auth-service --replicas=2 -n marimo-erp

# 7. Verify system health
curl https://marimo-erp.com/health
```

**Estimated RTO:** 1 hour
**Estimated RPO:** Last backup (typically 1-24 hours)

### 2. Service Recovery

#### Scenario: Single service failure

**Recovery Steps:**

```bash
# 1. Identify failed service
kubectl get pods -n marimo-erp
kubectl describe pod <failed-pod> -n marimo-erp

# 2. Check logs
kubectl logs <failed-pod> -n marimo-erp --previous

# 3. Restart deployment
kubectl rollout restart deployment/<service-name> -n marimo-erp

# 4. Monitor rollout
kubectl rollout status deployment/<service-name> -n marimo-erp

# 5. Verify health
kubectl port-forward svc/<service-name>-service 8080:8080 -n marimo-erp
curl http://localhost:8080/health
```

**Estimated RTO:** 5-15 minutes
**Estimated RPO:** 0 (no data loss)

### 3. Complete Cluster Failure

#### Scenario: Entire Kubernetes cluster unavailable

**Recovery Steps:**

```bash
# 1. Provision new Kubernetes cluster
# (Follow your cloud provider's instructions)

# 2. Install required components
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/cloud/deploy.yaml
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# 3. Run disaster recovery script
./scripts/disaster-recovery.sh marimo-erp /backups/marimo_backup_latest.dump.gz

# 4. Verify all services
kubectl get all -n marimo-erp

# 5. Update DNS
# Point domain to new cluster's ingress IP

# 6. Test application
curl https://marimo-erp.com/health

# 7. Monitor metrics
kubectl top pods -n marimo-erp
```

**Estimated RTO:** 2-4 hours
**Estimated RPO:** Last backup (typically 1-24 hours)

### 4. Datacenter Failover

#### Scenario: Primary datacenter unavailable

**Prerequisites:**
- Multi-region Kubernetes setup
- Database replication to secondary region
- DNS failover capability

**Recovery Steps:**

```bash
# 1. Confirm primary datacenter is down
kubectl --context=primary cluster-info

# 2. Promote secondary database to primary
# (Specific to your database replication setup)

# 3. Update DNS to point to secondary region
# (Use your DNS provider's API or console)

# 4. Verify secondary cluster
kubectl --context=secondary get pods -n marimo-erp

# 5. Scale up services if needed
kubectl --context=secondary scale deployment api-gateway --replicas=3 -n marimo-erp
kubectl --context=secondary scale deployment auth-service --replicas=2 -n marimo-erp

# 6. Monitor application
kubectl --context=secondary logs -n marimo-erp -l app=api-gateway -f
```

**Estimated RTO:** 30-60 minutes
**Estimated RPO:** Depends on replication lag (typically < 1 minute)

## Failure Scenarios

### Scenario Matrix

| Scenario | Likelihood | Impact | RTO | Recovery Procedure |
|----------|-----------|--------|-----|-------------------|
| Single pod failure | High | Low | 5 min | Automatic (K8s restart) |
| Database corruption | Low | Critical | 1 hour | Database Recovery |
| Service deployment failure | Medium | Medium | 15 min | Rollback deployment |
| Node failure | Medium | Low-Medium | 10 min | Automatic (K8s reschedule) |
| Cluster failure | Low | Critical | 2-4 hours | Complete Cluster Failure |
| Datacenter outage | Very Low | Critical | 30-60 min | Datacenter Failover |
| Data breach | Low | Critical | Varies | Security Incident Response |
| Accidental deletion | Medium | Varies | 1-2 hours | Database Recovery |

### Recovery Decision Tree

```
Failure Detected
    │
    ├─ Single Pod Failed?
    │   └─> Wait for automatic restart (K8s self-healing)
    │
    ├─ Multiple Pods Failed?
    │   └─> Check node health → Restart deployment
    │
    ├─ Database Issue?
    │   ├─> Connection issue → Check network/credentials
    │   └─> Data corruption → Restore from backup
    │
    ├─ Cluster Unreachable?
    │   └─> Execute Complete Cluster Recovery
    │
    └─ Datacenter Down?
        └─> Execute Datacenter Failover
```

## Testing

### DR Testing Schedule

| Test Type | Frequency | Scope |
|-----------|-----------|-------|
| Backup verification | Daily | Automated |
| Single service recovery | Monthly | Selected service |
| Database restore | Quarterly | Full restore to test environment |
| Complete cluster recovery | Annually | Full DR procedure |
| Datacenter failover | Annually | Multi-region setup |

### Test Procedures

#### Monthly Service Recovery Test

```bash
# 1. Select random service
SERVICE=$(kubectl get deployments -n marimo-erp -o name | shuf -n 1)

# 2. Simulate failure
kubectl delete $SERVICE -n marimo-erp

# 3. Time recovery
START_TIME=$(date +%s)
kubectl apply -f k8s/${SERVICE##*/}.yml
kubectl wait --for=condition=available $SERVICE -n marimo-erp --timeout=300s
END_TIME=$(date +%s)

# 4. Calculate RTO
echo "Recovery time: $((END_TIME - START_TIME)) seconds"

# 5. Verify functionality
# Run smoke tests
```

#### Quarterly Database Restore Test

```bash
# 1. Create test namespace
kubectl create namespace marimo-erp-test

# 2. Deploy infrastructure
kubectl apply -f k8s/postgres-statefulset.yml -n marimo-erp-test

# 3. Restore latest backup
./scripts/restore.sh /backups/marimo_backup_latest.dump.gz marimo-erp-test

# 4. Verify data integrity
kubectl exec -n marimo-erp-test postgres-0 -- psql -U postgres -d marimo_erp -c "SELECT COUNT(*) FROM users;"
kubectl exec -n marimo-erp-test postgres-0 -- psql -U postgres -d marimo_erp -c "SELECT * FROM users LIMIT 5;"

# 5. Cleanup
kubectl delete namespace marimo-erp-test
```

## Responsibilities

### On-Call Rotation

| Role | Primary Contact | Backup Contact |
|------|----------------|----------------|
| Incident Commander | DevOps Lead | Engineering Manager |
| Database Administrator | DBA | Senior Backend Engineer |
| Infrastructure Engineer | DevOps Engineer | Cloud Architect |
| Application Support | Backend Lead | Senior Developer |
| Communications | Product Manager | Engineering Manager |

### Escalation Path

```
Level 1: On-call Engineer (0-15 minutes)
    ↓ (if unresolved)
Level 2: Team Lead (15-30 minutes)
    ↓ (if unresolved)
Level 3: Engineering Manager (30-60 minutes)
    ↓ (if critical)
Level 4: CTO (60+ minutes)
```

### Contact Information

Update this section with actual contact information:

```yaml
oncall_primary:
  name: "DevOps Engineer"
  phone: "+1-XXX-XXX-XXXX"
  email: "oncall@company.com"
  slack: "@oncall"

oncall_backup:
  name: "Senior Engineer"
  phone: "+1-XXX-XXX-XXXX"
  email: "backup@company.com"
  slack: "@backup"

escalation:
  name: "Engineering Manager"
  phone: "+1-XXX-XXX-XXXX"
  email: "manager@company.com"
  slack: "@manager"
```

## Communication Plan

### Incident Declaration

When to declare an incident:
- Service downtime > 5 minutes
- Data loss detected
- Security breach suspected
- Multiple component failures

### Communication Channels

1. **Internal:**
   - Slack: #incidents channel
   - PagerDuty: Automated alerts
   - Email: incident-team@company.com

2. **External:**
   - Status page: status.marimo-erp.com
   - Customer email: support@company.com
   - Social media: @MarimoERP

### Status Updates

- **Initial notification:** Within 15 minutes of incident detection
- **Progress updates:** Every 30 minutes during active incident
- **Resolution notification:** Within 15 minutes of resolution
- **Post-mortem:** Within 48 hours of resolution

## Post-Incident

### Post-Mortem Template

```markdown
# Incident Post-Mortem

## Incident Summary
- **Date:** YYYY-MM-DD
- **Duration:** X hours Y minutes
- **Severity:** Critical/High/Medium/Low
- **Impact:** Number of users affected, services down

## Timeline
- HH:MM - Incident detected
- HH:MM - Team notified
- HH:MM - Root cause identified
- HH:MM - Fix applied
- HH:MM - Service restored
- HH:MM - Incident resolved

## Root Cause
[Detailed description]

## Resolution
[What was done to resolve]

## Impact
- Users affected: X
- Data loss: Yes/No
- Revenue impact: $X

## Action Items
1. [ ] Prevent recurrence
2. [ ] Improve monitoring
3. [ ] Update runbooks
4. [ ] Infrastructure changes

## Lessons Learned
[What we learned]
```

### Continuous Improvement

After each incident:
1. Conduct blameless post-mortem
2. Update DR procedures
3. Improve monitoring and alerts
4. Add automated tests
5. Share learnings with team

## Appendix

### Useful Commands

```bash
# Quick status check
kubectl get all -n marimo-erp

# Check resource usage
kubectl top pods -n marimo-erp

# View recent events
kubectl get events -n marimo-erp --sort-by='.lastTimestamp'

# Describe all pods
kubectl describe pods -n marimo-erp

# Get logs from all pods
kubectl logs -n marimo-erp -l app=api-gateway --tail=100

# Port forward for testing
kubectl port-forward -n marimo-erp svc/api-gateway-service 8080:8080
```

### External Resources

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [PostgreSQL Backup and Recovery](https://www.postgresql.org/docs/current/backup.html)
- [Disaster Recovery Best Practices](https://cloud.google.com/architecture/dr-scenarios-planning-guide)

---

**Document Version:** 1.0
**Last Updated:** 2024-01-15
**Next Review:** 2024-04-15
**Owner:** DevOps Team
