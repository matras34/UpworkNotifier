#!/bin/bash
# Web SSH Service - One-Click Installer with Docker

echo "============================================"
echo "  Web SSH - Установка за 1 клик"
echo "  Web SSH - One-Click Installation"
echo "============================================"
echo ""

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker не установлен / Docker is not installed!"
    echo ""
    echo "Установите Docker / Install Docker:"
    echo "  Windows/Mac: https://www.docker.com/products/docker-desktop"
    echo "  Linux: curl -fsSL https://get.docker.com | sh"
    echo ""
    exit 1
fi

# Check if docker-compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "❌ docker-compose не установлен / docker-compose is not installed!"
    echo ""
    echo "Обычно устанавливается вместе с Docker Desktop"
    echo "Usually installed with Docker Desktop"
    echo ""
    exit 1
fi

echo "✅ Docker найден / Docker found"
echo ""

# Go to project root
cd "$(dirname "$0")"

echo "🚀 Запуск всех сервисов / Starting all services..."
echo ""
echo "Это может занять несколько минут при первом запуске..."
echo "This may take a few minutes on first run..."
echo ""

# Start services
docker-compose up -d

# Wait for services to be ready
echo ""
echo "⏳ Ожидание запуска сервисов / Waiting for services to start..."
sleep 10

echo ""
echo "============================================"
echo "✅ Установка завершена! / Installation complete!"
echo "============================================"
echo ""
echo "🌐 Откройте в браузере / Open in browser:"
echo "   http://localhost"
echo ""
echo "🔐 Вход по умолчанию / Default login:"
echo "   Email: admin@example.com"
echo "   Password: admin123"
echo ""
echo "📋 Управление / Management:"
echo "   Остановить / Stop:   docker-compose stop"
echo "   Перезапустить / Restart: docker-compose restart"
echo "   Удалить / Remove:    docker-compose down"
echo "   Логи / Logs:         docker-compose logs -f"
echo ""
echo "============================================"
