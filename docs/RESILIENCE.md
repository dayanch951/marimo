# Resilience Patterns - Руководство

## Обзор

Система Marimo ERP использует несколько паттернов устойчивости для обеспечения надежности и доступности.

## Circuit Breaker

### Описание

Circuit Breaker защищает систему от каскадных сбоев, быстро отклоняя запросы к недоступным сервисам.

### Использование

```go
package main

import (
    "github.com/dayanch951/marimo/shared/resilience"
    "time"
)

func main() {
    // Создание circuit breaker
    cb := resilience.NewCircuitBreaker(resilience.Settings{
        Name:        "payment-service",
        MaxRequests: 3,                  // Макс. запросов в half-open
        Interval:    60 * time.Second,   // Окно анализа
        Timeout:     30 * time.Second,   // Время до half-open
        Threshold:   5,                  // Мин. запросов для анализа
        FailureRate: 0.5,                // 50% error rate для открытия
        OnStateChange: func(name string, from, to resilience.State) {
            log.Printf("CB %s: %s -> %s", name, from, to)
        },
    })

    // Выполнение операции
    err := cb.Execute(func() error {
        return callExternalService()
    })

    if err == resilience.ErrCircuitOpen {
        // Circuit открыт, использовать fallback
        return useFallback()
    }

    // Получение результата
    result, err := cb.Call(func() (interface{}, error) {
        return fetchData()
    })
}
```

### Состояния

#### Closed (Закрыт)
- Нормальная работа
- Все запросы проходят
- Считаются ошибки

#### Open (Открыт)
- Сервис считается недоступным
- Запросы отклоняются немедленно
- Возвращается `ErrCircuitOpen`
- После timeout переход в half-open

#### Half-Open (Полуоткрыт)
- Тестирование восстановления
- Пропускается MaxRequests запросов
- При успехе → Closed
- При ошибке → Open

### Мониторинг

```go
// Получить текущее состояние
state := cb.State()

// Получить статистику
requests, successes, failures := cb.Counts()

// Сброс состояния
cb.Reset()
```

### Best Practices

1. **Настройка threshold**
   - Слишком низкий: частые ложные срабатывания
   - Слишком высокий: медленная реакция на проблемы
   - Рекомендуется: 5-10 запросов

2. **Failure rate**
   - 0.5 (50%) - хороший баланс
   - Критичные сервисы: 0.3 (30%)
   - Некритичные: 0.7 (70%)

3. **Timeout**
   - Время на восстановление сервиса
   - Рекомендуется: 30-60 секунд
   - Для внешних API: 60-120 секунд

4. **Fallback стратегии**
   - Кеширование последнего успешного ответа
   - Значения по умолчанию
   - Альтернативный сервис
   - Graceful degradation

## Retry Logic

### Описание

Автоматические повторные попытки при временных сбоях.

### Использование

```go
package main

import (
    "context"
    "github.com/dayanch951/marimo/shared/resilience"
    "time"
)

func main() {
    ctx := context.Background()

    // Стандартная политика
    policy := resilience.DefaultRetryPolicy()

    // Кастомная политика
    policy := resilience.RetryPolicy{
        MaxAttempts:  3,
        InitialDelay: 100 * time.Millisecond,
        MaxDelay:     10 * time.Second,
        Multiplier:   2.0,
        Jitter:       true,
    }

    // Простая retry
    err := resilience.Retry(ctx, policy, func() error {
        return makeHTTPRequest()
    })

    // Retry с результатом
    result, err := resilience.RetryWithResult(ctx, policy, func() (User, error) {
        return fetchUser(123)
    })
}
```

### Стратегии Backoff

#### Exponential Backoff (по умолчанию)
```
Попытка 1: 100ms
Попытка 2: 200ms (100 * 2^1)
Попытка 3: 400ms (100 * 2^2)
Попытка 4: 800ms (100 * 2^3)
```

#### Linear Backoff
```go
delay := resilience.LinearBackoff(attempt, 200*time.Millisecond)
// Попытка 1: 200ms
// Попытка 2: 400ms
// Попытка 3: 600ms
```

#### Constant Backoff
```go
delay := resilience.ConstantBackoff(500*time.Millisecond)
// Все попытки: 500ms
```

### Jitter

Добавление случайности для предотвращения "thundering herd":

```go
policy := resilience.RetryPolicy{
    InitialDelay: 100 * time.Millisecond,
    Jitter:       true, // ±5% случайности
}
```

### Retryable Errors

Определение ошибок, которые должны вызывать retry:

```go
var (
    ErrTimeout = errors.New("timeout")
    ErrNetwork = errors.New("network error")
)

policy := resilience.RetryPolicy{
    RetryableErrors: []error{ErrTimeout, ErrNetwork},
}
```

### HTTP Retry

Специальная обработка HTTP статус кодов:

```go
// Проверка retryable статуса
if resilience.IsRetryableHTTPStatus(statusCode) {
    // 408, 429, 500, 502, 503, 504
    // Выполнить retry
}
```

### Best Practices

1. **MaxAttempts**
   - Обычно: 3 попытки
   - Идемпотентные операции: 5 попыток
   - Неидемпотентные: 1-2 попытки

2. **Timeout контекста**
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
   defer cancel()

   err := resilience.Retry(ctx, policy, fn)
   ```

3. **Логирование**
   ```go
   for attempt := 1; attempt <= maxAttempts; attempt++ {
       log.Printf("Attempt %d/%d", attempt, maxAttempts)
       err := fn()
       if err == nil {
           return nil
       }
   }
   ```

4. **Идемпотентность**
   - Убедитесь, что операция безопасна для повторения
   - Используйте idempotency keys для API
   - Проверяйте состояние перед retry

5. **Rate Limiting**
   - Учитывайте rate limits внешних API
   - Используйте exponential backoff
   - Добавляйте jitter

## Комбинирование паттернов

### Circuit Breaker + Retry

```go
func CallServiceWithResilience(serviceName string) (interface{}, error) {
    cb := getCircuitBreaker(serviceName)
    policy := resilience.DefaultRetryPolicy()
    ctx := context.Background()

    result, err := resilience.RetryWithResult(ctx, policy, func() (interface{}, error) {
        var res interface{}
        err := cb.Execute(func() error {
            var innerErr error
            res, innerErr = callService()
            return innerErr
        })
        return res, err
    })

    return result, err
}
```

### Circuit Breaker + Cache

```go
func GetUserWithFallback(userID string) (User, error) {
    cb := getCircuitBreaker("users")
    cache := getCache()

    var user User
    err := cb.Execute(func() error {
        return cache.GetOrSet(
            "user:"+userID,
            &user,
            5*time.Minute,
            func() (interface{}, error) {
                return fetchUserFromDB(userID)
            },
        )
    })

    if err == resilience.ErrCircuitOpen {
        // Fallback: попробовать из кеша
        if cacheErr := cache.Get("user:"+userID, &user); cacheErr == nil {
            return user, nil
        }
    }

    return user, err
}
```

## Мониторинг и Алерты

### Метрики

```go
// Prometheus метрики для circuit breaker
circuitBreakerState := prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "circuit_breaker_state",
        Help: "Current state of circuit breaker (0=closed, 1=open, 2=half-open)",
    },
    []string{"service"},
)

circuitBreakerRequests := prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "circuit_breaker_requests_total",
        Help: "Total requests through circuit breaker",
    },
    []string{"service", "result"},
)

// Обновление метрик
cb.OnStateChange = func(name string, from, to State) {
    circuitBreakerState.WithLabelValues(name).Set(float64(to))
}
```

### Алерты

```yaml
# Prometheus alerts
groups:
  - name: resilience
    rules:
      - alert: CircuitBreakerOpen
        expr: circuit_breaker_state > 0
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Circuit breaker {{ $labels.service }} is open"

      - alert: HighRetryRate
        expr: rate(retry_attempts_total[5m]) > 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High retry rate for {{ $labels.service }}"
```

## Тестирование

### Unit тесты для Circuit Breaker

```go
func TestCircuitBreaker(t *testing.T) {
    cb := resilience.NewCircuitBreaker(resilience.Settings{
        Name:        "test",
        Threshold:   3,
        FailureRate: 0.5,
    })

    // Проверка нормальной работы
    for i := 0; i < 5; i++ {
        err := cb.Execute(func() error {
            return nil
        })
        assert.NoError(t, err)
    }

    // Проверка открытия после ошибок
    for i := 0; i < 5; i++ {
        cb.Execute(func() error {
            return errors.New("error")
        })
    }

    assert.Equal(t, resilience.StateOpen, cb.State())

    // Проверка быстрого отклонения
    err := cb.Execute(func() error {
        t.Fatal("Should not be called")
        return nil
    })
    assert.Equal(t, resilience.ErrCircuitOpen, err)
}
```

### Integration тесты

```go
func TestRetryWithRealService(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Симулировать временную ошибку
        if rand.Float64() < 0.3 {
            w.WriteHeader(http.StatusServiceUnavailable)
            return
        }
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()

    policy := resilience.RetryPolicy{
        MaxAttempts: 3,
        InitialDelay: 10 * time.Millisecond,
    }

    err := resilience.Retry(context.Background(), policy, func() error {
        resp, err := http.Get(server.URL)
        if err != nil {
            return err
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            return fmt.Errorf("status: %d", resp.StatusCode)
        }
        return nil
    })

    assert.NoError(t, err)
}
```

## Troubleshooting

### Circuit Breaker постоянно открыт

**Причины:**
1. Threshold слишком низкий
2. Backend действительно недоступен
3. Timeout слишком короткий

**Решения:**
```go
// Увеличить threshold
cb := resilience.NewCircuitBreaker(resilience.Settings{
    Threshold: 10, // было 5
})

// Увеличить timeout
cb := resilience.NewCircuitBreaker(resilience.Settings{
    Timeout: 60 * time.Second, // было 30s
})

// Проверить логи backend сервиса
```

### Слишком много retry попыток

**Причины:**
1. MaxAttempts слишком высокий
2. Не определены RetryableErrors
3. Backend всегда возвращает retryable error

**Решения:**
```go
// Ограничить попытки
policy := resilience.RetryPolicy{
    MaxAttempts: 2, // было 5
}

// Определить retryable errors
policy := resilience.RetryPolicy{
    RetryableErrors: []error{ErrTimeout, ErrNetwork},
}
```

### Thundering Herd

**Причины:**
1. Jitter отключен
2. Все клиенты используют одинаковый backoff

**Решения:**
```go
// Включить jitter
policy := resilience.RetryPolicy{
    Jitter: true,
}

// Добавить случайную задержку перед началом
time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
```

## Дополнительные ресурсы

- [Architecture](./ARCHITECTURE.md) - общая архитектура
- [Observability](./OBSERVABILITY.md) - мониторинг
- [Testing](./TESTING.md) - тестирование

## Ссылки

- [Circuit Breaker Pattern](https://martinfowler.com/bliki/CircuitBreaker.html)
- [Retry Pattern](https://docs.microsoft.com/en-us/azure/architecture/patterns/retry)
- [Exponential Backoff](https://en.wikipedia.org/wiki/Exponential_backoff)
