#!/bin/bash
# Скрипт для быстрого merge в main

echo "Создание локальной ветки main и merge..."

# Переключиться на main
git fetch origin main
git checkout -B main origin/main

# Merge нашей ветки
git merge claude/go-microservice-react-auth-011CV1SjuuSbv4wEBkXXK9RS --no-ff -m "Merge microservices architecture"

echo "Merge выполнен локально"
echo ""
echo "Для push в GitHub выполните:"
echo "git push origin main"
echo ""
echo "ВНИМАНИЕ: Если ветка main защищена, используйте Pull Request через веб-интерфейс"
