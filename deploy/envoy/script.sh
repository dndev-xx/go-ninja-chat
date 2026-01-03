#!/bin/bash

# start-local-envoy.sh - Запуск Envoy локально для тестирования

set -e

echo "🚀 Запуск локального Envoy Gateway для тестирования"

# Проверка наличия Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Docker не найден. Пожалуйста, установите Docker."
    exit 1
fi

# Создаем директорию для тестирования
mkdir -p envoy-local-test
cd envoy-local-test

echo "📁 Настройка окружения..."

# Останавливаем существующие контейнеры
docker-compose down 2>/dev/null || true

echo "🏗️  Запуск Envoy и тестового backend..."

# Запуск сервисов
docker-compose up -d

echo "⏳ Ожидание запуска сервисов..."
sleep 10

# Проверка статуса
echo "📊 Статус контейнеров:"
docker-compose ps

echo ""
echo "✅ Envoy Gateway запущен!"
echo ""
echo "🌐 Доступные адреса:"
echo "  📊 Envoy Admin Panel:     http://localhost:9901"
echo "  🎯 Через Envoy Proxy:    http://localhost:10000"
echo "  🔗 Напрямую на Backend:  http://localhost:8080"
echo ""
echo "🧪 Тестирование соединения:"

# Тест прямого соединения
echo -n "  Backend (прямо): "
if curl -s http://localhost:8080 >/dev/null 2>&1; then
    echo "✅ OK"
else
    echo "❌ Недоступен"
fi

# Тест через Envoy
echo -n "  Backend (через Envoy): "
if curl -s http://localhost:10000 >/dev/null 2>&1; then
    echo "✅ OK"
else
    echo "❌ Недоступен"
fi

# Тест админ панели
echo -n "  Envoy Admin: "
if curl -s http://localhost:9901 >/dev/null 2>&1; then
    echo "✅ OK"
else
    echo "❌ Недоступен"
fi

echo ""
echo "📋 Полезные команды:"
echo "  docker-compose logs -f          # Просмотр логов"
echo "  docker-compose logs envoy-local # Логи только Envoy"
echo "  docker-compose down             # Остановка сервисов"
echo ""

# Автоматически открываем админ панель (если возможно)
if command -v xdg-open &> /dev/null; then
    echo "🔗 Открываем Envoy Admin Panel..."
    xdg-open http://localhost:9901
elif command -v open &> /dev/null; then
    echo "🔗 Открываем Envoy Admin Panel..."
    open http://localhost:9901
else
    echo "💡 Откройте http://localhost:9901 в браузере для доступа к админ панели"
fi

echo ""
echo "🎉 Готово! Envoy работает и готов к тестированию."
