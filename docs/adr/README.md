# Architecture Decision Records (ADR)

This directory contains records of significant architectural decisions made in the Marimo ERP project.

## What is an ADR?

An Architecture Decision Record (ADR) is a document that captures an important architectural decision made along with its context and consequences.

## ADR Template

Each ADR follows this structure:
- **Title**: Short noun phrase
- **Status**: Proposed, Accepted, Deprecated, Superseded
- **Context**: What is the issue we're facing?
- **Decision**: What are we going to do about it?
- **Consequences**: What becomes easier or harder as a result?

## Index

| ADR | Title | Status |
|-----|-------|--------|
| [ADR-001](./001-microservices-architecture.md) | Microservices Architecture | Accepted |
| [ADR-002](./002-multi-tenancy-strategy.md) | Multi-tenancy Strategy | Accepted |
| [ADR-003](./003-event-driven-messaging.md) | Event-Driven Messaging with RabbitMQ | Accepted |
| [ADR-004](./004-api-gateway-pattern.md) | API Gateway Pattern | Accepted |
| [ADR-005](./005-jwt-authentication.md) | JWT-based Authentication | Accepted |
| [ADR-006](./006-postgresql-database.md) | PostgreSQL as Primary Database | Accepted |
| [ADR-007](./007-kubernetes-deployment.md) | Kubernetes for Container Orchestration | Accepted |
| [ADR-008](./008-react-frontend.md) | React for Frontend Development | Accepted |
| [ADR-009](./009-prometheus-monitoring.md) | Prometheus for Monitoring | Accepted |
| [ADR-010](./010-blue-green-deployment.md) | Blue-Green Deployment Strategy | Accepted |

## Creating New ADRs

When creating a new ADR:

1. Copy the template from `template.md`
2. Number it sequentially (e.g., `011-your-decision.md`)
3. Fill in all sections
4. Get review from team
5. Update this index
6. Commit to repository

## References

- [Michael Nygard's ADR](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
- [ADR GitHub Organization](https://adr.github.io/)
