#!/bin/bash
# stop.sh
set -e

stop_compose() {
    echo "Остановка $1..."
    docker-compose -f "$1" down
    if [ $? -eq 0 ]; then
        echo "✓ $1 успешно остановлен"
    else
        echo "✗ Ошибка при остановке $1"
        exit 1
    fi
}

stop_compose "docker-compose-kafka.yml"
stop_compose "docker-compose-pg.yml"

echo "Все сервисы остановлены успешно!"
