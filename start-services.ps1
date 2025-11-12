# Marimo ERP - Start All Services Script
# This script starts all microservices locally

$env:USE_POSTGRES="true"
$env:DB_HOST="localhost"
$env:DB_PORT="5432"
$env:DB_USER="postgres"
$env:DB_PASSWORD="secure-postgres-password-2024"
$env:DB_NAME="marimo_erp"
$env:JWT_SECRET="marimo-super-secure-jwt-secret-key-minimum-32-characters-long-2024"
$env:LOG_LEVEL="info"
$env:CONSUL_ADDR="localhost:8500"
$env:REDIS_ADDR="localhost:6379"
$env:RABBITMQ_URL="amqp://admin:secure-rabbitmq-password-2024@localhost:5672/"

Write-Host "Starting Marimo ERP Services..." -ForegroundColor Green

# Start Users Service (Port 8081)
Write-Host "Starting Users Service on :8081" -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$PSScriptRoot\services\users'; go run cmd/server/main.go"

Start-Sleep -Seconds 3

# Start Config Service (Port 8082)
Write-Host "Starting Config Service on :8082" -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$PSScriptRoot\services\config'; go run cmd/server/main.go"

Start-Sleep -Seconds 2

# Start Accounting Service (Port 8083)
Write-Host "Starting Accounting Service on :8083" -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$PSScriptRoot\services\accounting'; go run cmd/server/main.go"

Start-Sleep -Seconds 2

# Start Factory Service (Port 8084)
Write-Host "Starting Factory Service on :8084" -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$PSScriptRoot\services\factory'; go run cmd/server/main.go"

Start-Sleep -Seconds 2

# Start Shop Service (Port 8085)
Write-Host "Starting Shop Service on :8085" -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$PSScriptRoot\services\shop'; go run cmd/server/main.go"

Start-Sleep -Seconds 2

# Start Main Service (Port 8086)
Write-Host "Starting Main Service on :8086" -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$PSScriptRoot\services\main'; go run cmd/server/main.go"

Start-Sleep -Seconds 3

# Start Gateway (Port 8080)
Write-Host "Starting API Gateway on :8080" -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$PSScriptRoot\services\gateway'; go run cmd/server/main.go"

Start-Sleep -Seconds 3

# Start Frontend (Port 3000)
Write-Host "Starting Frontend on :3000" -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$PSScriptRoot\frontend'; npm start"

Write-Host ""
Write-Host "All services started!" -ForegroundColor Green
Write-Host ""
Write-Host "Access points:" -ForegroundColor Yellow
Write-Host "  Frontend:        http://localhost:3000" -ForegroundColor White
Write-Host "  API Gateway:     http://localhost:8080" -ForegroundColor White
Write-Host "  Consul UI:       http://localhost:8500" -ForegroundColor White
Write-Host "  RabbitMQ UI:     http://localhost:15672" -ForegroundColor White
Write-Host "  Grafana:         http://localhost:3001" -ForegroundColor White
Write-Host "  Prometheus:      http://localhost:9090" -ForegroundColor White
Write-Host "  Jaeger:          http://localhost:16686" -ForegroundColor White
Write-Host "  Kibana:          http://localhost:5601" -ForegroundColor White
Write-Host ""
Write-Host "Default credentials:" -ForegroundColor Yellow
Write-Host "  Email: admin@example.com" -ForegroundColor White
Write-Host "  Password: admin123" -ForegroundColor White
