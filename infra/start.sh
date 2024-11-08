#!/bin/bash
# start.sh
set -e

check_file() {
    if [ ! -f "$1" ]; then
        echo "Ошибка: Файл $1 не найден"
        exit 1
    fi
}

start_compose() {
    echo "Запуск $1..."
    docker-compose -f "$1" up -d
    if [ $? -eq 0 ]; then
        echo "✓ $1 успешно запущен"
    else
        echo "✗ Ошибка при запуске $1"
        exit 1
    fi
}

check_file "docker-compose-kafka-v2.yml"
check_file "docker-compose-pg.yml"

start_compose "docker-compose-kafka-v2.yml"
start_compose "docker-compose-pg.yml"

echo "Все сервисы запущены успешно!"