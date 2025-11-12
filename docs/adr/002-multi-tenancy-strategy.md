# ADR-002: Multi-tenancy Strategy

## Status
Accepted

## Date
2024-01-20

## Context

Our ERP system needs to support multiple organizations (tenants) while:
- Ensuring complete data isolation between tenants
- Minimizing infrastructure costs
- Allowing easy scaling as tenants grow
- Supporting different subscription tiers
- Enabling custom domains for enterprise clients

We evaluated three multi-tenancy approaches:

1. **Separate Database per Tenant**
   - Complete isolation
   - Highest cost
   - Difficult to manage at scale
   - Easy backup/restore per tenant

2. **Separate Schema per Tenant**
   - Good isolation
   - Moderate cost
   - Connection pool limitations
   - Schema migration complexity

3. **Shared Schema with tenant_id**
   - Lowest cost
   - Requires careful query filtering
   - Easy to scale
   - Risk of data leakage if not implemented correctly

## Decision

We will implement a **shared schema with tenant_id** approach with the following safeguards:

### Implementation Details

1. **Database Design**
   - Add `tenant_id UUID` column to all tenant-specific tables
   - Create composite indexes: `(tenant_id, id)`, `(tenant_id, created_at)`
   - Use Row-Level Security (RLS) policies in PostgreSQL

2. **Tenant Resolution**
   Priority order:
   - `X-Tenant-ID` header (for API calls)
   - `X-Tenant-Slug` header
   - Custom domain lookup
   - Subdomain extraction

3. **Data Access**
   - Implement `TenantAwareDB` wrapper
   - Automatically inject `tenant_id` in all queries
   - Context-based tenant resolution
   - Middleware for HTTP requests

4. **Subscription Management**
   - Trial: 14 days, limited features
   - Starter: $29/month, 5 users
   - Professional: $99/month, 25 users
   - Enterprise: Custom pricing, unlimited users

### Code Example

```go
// Automatic tenant filtering
db := tenancy.TenantAwareDB(ctx, baseDB)
var users []User
db.Where("role = ?", "admin").Find(&users)
// Automatically adds: WHERE tenant_id = ? AND role = ?
```

## Consequences

### Positive

- **Cost Efficient**: Single database for all tenants
- **Easy Scaling**: No database proliferation
- **Simple Backups**: Single database backup
- **Fast Onboarding**: New tenants immediately available
- **Cross-tenant Analytics**: Easier for system-wide reporting
- **Resource Sharing**: Efficient use of database connections

### Negative

- **Data Leakage Risk**: Bugs could expose data across tenants
- **Query Complexity**: Must always filter by tenant_id
- **Noisy Neighbor**: One tenant can affect others' performance
- **Limited Customization**: Harder to provide tenant-specific schema changes
- **Backup Granularity**: Can't easily backup single tenant

### Mitigation Strategies

1. **Code Review**: Strict review for all database queries
2. **Testing**: Comprehensive tests with multiple tenants
3. **RLS Policies**: PostgreSQL row-level security as safety net
4. **Monitoring**: Track queries without tenant_id filter
5. **Resource Limits**: Implement usage limits per subscription tier
6. **Automated Checks**: Linter rules to enforce tenant filtering

### Security Measures

```sql
-- Row-Level Security Policy
CREATE POLICY tenant_isolation ON users
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

ALTER TABLE users ENABLE ROW LEVEL SECURITY;
```

## Alternatives Considered

### Separate Database per Tenant
**Rejected** because:
- High operational complexity (1000 tenants = 1000 databases)
- Expensive at scale
- Schema migration nightmares
- Connection pool exhaustion

### Separate Schema per Tenant
**Rejected** because:
- PostgreSQL limits on schemas
- Complex search path management
- Still has connection pool issues
- Migration complexity across hundreds of schemas

## Related Decisions

- [ADR-006](./006-postgresql-database.md) - PostgreSQL Database
- [ADR-005](./005-jwt-authentication.md) - JWT Authentication

## Migration Path

If data isolation becomes critical:
1. Move enterprise clients to dedicated schemas/databases
2. Hybrid approach: shared for small tenants, isolated for large
3. Tools already in place for tenant data export

## References

- [Multi-tenant SaaS Database Tenancy Patterns](https://docs.microsoft.com/en-us/azure/sql-database/saas-tenancy-app-design-patterns)
- [PostgreSQL Row Level Security](https://www.postgresql.org/docs/current/ddl-rowsecurity.html)
- [Force.com Multi-Tenant Architecture](https://developer.salesforce.com/page/Multi_Tenant_Architecture)
