# Blue-Green Deployment Strategy

Blue-Green deployment позволяет обновлять приложение без даунтайма, переключаясь между двумя идентичными окружениями.

## Принцип работы

1. **Blue** - текущая production версия
2. **Green** - новая версия для тестирования
3. Переключение трафика с Blue на Green после валидации
4. Откат на Blue при проблемах

## Использование

### 1. Deploy новой версии в Green

```bash
# Обновить image в green deployment
kubectl apply -f k8s/blue-green/api-gateway-green.yml

# Проверить статус
kubectl get pods -n marimo-erp -l color=green
kubectl wait --for=condition=ready pod -l app=api-gateway,color=green -n marimo-erp --timeout=300s
```

### 2. Тестирование Green окружения

```bash
# Временный доступ к green через port-forward
kubectl port-forward -n marimo-erp svc/api-gateway-green 9090:8080

# Или через тестовый ingress
curl -H "Host: green.marimo-erp.com" https://your-cluster/health
```

### 3. Переключение трафика на Green

```bash
# Обновить service selector на green
kubectl patch service api-gateway-service -n marimo-erp -p '{"spec":{"selector":{"color":"green"}}}'

# Проверить что трафик идет на green
kubectl describe svc api-gateway-service -n marimo-erp
```

### 4. Мониторинг после переключения

```bash
# Проверить логи
kubectl logs -n marimo-erp -l app=api-gateway,color=green --tail=100 -f

# Проверить метрики
kubectl top pods -n marimo-erp -l color=green
```

### 5. Откат на Blue (если нужно)

```bash
# Быстро переключить обратно на blue
kubectl patch service api-gateway-service -n marimo-erp -p '{"spec":{"selector":{"color":"blue"}}}'
```

### 6. Очистка старой версии

```bash
# После успешного деплоя удалить blue
kubectl delete deployment api-gateway-blue -n marimo-erp

# Переименовать green в blue для следующего деплоя
kubectl label deployment api-gateway-green color=blue --overwrite -n marimo-erp
```

## Automated Script

Используйте `scripts/blue-green-deploy.sh` для автоматического деплоя:

```bash
./scripts/blue-green-deploy.sh api-gateway marimo-erp/api-gateway:v1.2.3
```
