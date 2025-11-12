# ADR-003: Event-Driven Messaging with RabbitMQ

## Status
Accepted

## Date
2024-01-22

## Context

In our microservices architecture, services need to communicate asynchronously for:
- Decoupling services
- Handling long-running operations
- Publishing events to multiple consumers
- Ensuring message delivery reliability
- Supporting different messaging patterns (pub/sub, work queues, RPC)

We evaluated several message brokers:

| Broker | Pros | Cons |
|--------|------|------|
| **RabbitMQ** | Mature, feature-rich, AMQP standard, flexible routing | Lower throughput than Kafka |
| **Apache Kafka** | High throughput, persistent logs, great for streams | Complex setup, overkill for our use case |
| **Redis Pub/Sub** | Simple, fast | No persistence, no delivery guarantees |
| **NATS** | Lightweight, fast | Limited features, smaller ecosystem |
| **AWS SQS** | Managed, reliable | Vendor lock-in, higher cost, latency |

## Decision

We will use **RabbitMQ** as our primary message broker for asynchronous communication between services.

### Implementation Strategy

1. **Message Patterns**
   - **Work Queues**: Background job processing (email sending, report generation)
   - **Pub/Sub**: Event broadcasting (user.created, payment.succeeded)
   - **Topic Exchange**: Filtered event routing
   - **Dead Letter Queue**: Failed message handling

2. **Exchange Configuration**
   ```go
   // Events exchange (topic)
   events.exchange
   - user.created
   - user.updated
   - user.deleted
   - payment.succeeded
   - payment.failed

   // Tasks exchange (direct)
   tasks.exchange
   - send.email
   - generate.report
   - process.webhook
   ```

3. **Message Structure**
   ```json
   {
     "id": "uuid",
     "type": "user.created",
     "timestamp": "2024-01-22T10:00:00Z",
     "tenant_id": "uuid",
     "data": {},
     "metadata": {
       "correlation_id": "uuid",
       "user_id": "uuid"
     }
   }
   ```

4. **Reliability Features**
   - Publisher confirms
   - Consumer acknowledgments
   - Message persistence
   - Dead letter queues
   - Retry with exponential backoff

## Consequences

### Positive

- **Decoupling**: Services don't need to know about each other
- **Reliability**: Message persistence and delivery guarantees
- **Scalability**: Easy to add consumers for load distribution
- **Flexibility**: Multiple exchange types and routing patterns
- **Monitoring**: Built-in management UI and metrics
- **Standards**: AMQP protocol ensures interoperability
- **Ecosystem**: Large community and good client libraries

### Negative

- **Operational Complexity**: Need to manage RabbitMQ cluster
- **Learning Curve**: Understanding AMQP concepts
- **Single Point of Failure**: Need clustering for HA
- **Message Ordering**: Not guaranteed across partitions
- **Debugging**: Async flows harder to trace

### Mitigation Strategies

1. **High Availability**
   ```yaml
   # RabbitMQ cluster with 3 nodes
   - rabbitmq-1 (master)
   - rabbitmq-2 (mirror)
   - rabbitmq-3 (mirror)
   ```

2. **Monitoring**
   - Prometheus metrics exporter
   - Alert on queue depth > 1000
   - Track message rates and latencies

3. **Circuit Breaker**
   ```go
   // Prevent cascade failures
   if queue.Depth() > maxDepth {
       return ErrCircuitOpen
   }
   ```

4. **Distributed Tracing**
   - Propagate trace context in message metadata
   - Jaeger integration for async flow visualization

## Use Cases

### 1. Email Sending (Work Queue)
```go
// Publisher
rabbitMQ.Publish("tasks.exchange", "send.email", EmailTask{
    To: "user@example.com",
    Template: "welcome",
})

// Consumer
func ProcessEmailTask(msg EmailTask) error {
    return sendgrid.Send(msg)
}
```

### 2. Event Broadcasting (Pub/Sub)
```go
// Publisher
rabbitMQ.Publish("events.exchange", "user.created", UserCreatedEvent{
    UserID: "uuid",
    Email: "user@example.com",
})

// Multiple consumers
// - Analytics Service: Track user signup
// - Email Service: Send welcome email
// - Webhook Service: Notify external systems
```

### 3. Webhook Delivery (Dead Letter Queue)
```go
// Failed webhook goes to DLQ
if err := deliverWebhook(url, payload); err != nil {
    msg.Nack(false, false) // Send to DLQ
}

// DLQ consumer retries with backoff
```

## Performance Characteristics

- **Throughput**: ~20,000 messages/sec (single node)
- **Latency**: < 10ms (P95)
- **Message Size**: Up to 128MB (not recommended > 1MB)
- **Persistence**: Optional (we enable for critical messages)

## Alternatives Considered

### Apache Kafka
**Rejected** because:
- Too complex for our messaging volume
- Overkill for simple pub/sub
- Higher operational overhead
- We don't need log-based storage

### Redis Pub/Sub
**Rejected** because:
- No message persistence
- No delivery guarantees
- Fire-and-forget model unsuitable for critical operations

## Migration Path

If we outgrow RabbitMQ:
1. Add Kafka for high-volume event streaming
2. Keep RabbitMQ for task queues and RPC
3. Use both based on use case

## Related Decisions

- [ADR-001](./001-microservices-architecture.md) - Microservices Architecture
- [ADR-009](./009-prometheus-monitoring.md) - Prometheus Monitoring

## References

- [RabbitMQ Patterns](https://www.rabbitmq.com/getstarted.html)
- [Enterprise Integration Patterns](https://www.enterpriseintegrationpatterns.com/)
- [AMQP 0-9-1 Model](https://www.rabbitmq.com/tutorials/amqp-concepts.html)
