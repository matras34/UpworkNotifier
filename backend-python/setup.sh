#!/bin/bash
# Web SSH Service - Easy Setup Script for Beginners

echo "============================================"
echo "  Web SSH Service - Easy Setup"
echo "============================================"
echo ""

# Check Python version
echo "🔍 Checking Python version..."
if ! command -v python3 &> /dev/null; then
    echo "❌ Python 3 is not installed!"
    echo "Please install Python 3.8 or higher from https://www.python.org/"
    exit 1
fi

PYTHON_VERSION=$(python3 --version 2>&1 | awk '{print $2}')
echo "✅ Found Python $PYTHON_VERSION"
echo ""

# Check if we're in the right directory
if [ ! -f "requirements.txt" ]; then
    echo "❌ Please run this script from the backend-python directory"
    exit 1
fi

# Create virtual environment
echo "📦 Creating virtual environment..."
if [ ! -d "venv" ]; then
    python3 -m venv venv
    echo "✅ Virtual environment created"
else
    echo "✅ Virtual environment already exists"
fi
echo ""

# Activate virtual environment
echo "🔧 Activating virtual environment..."
source venv/bin/activate
echo "✅ Virtual environment activated"
echo ""

# Install dependencies
echo "📥 Installing dependencies (this may take a few minutes)..."
pip install --upgrade pip > /dev/null 2>&1
pip install -r requirements.txt
echo "✅ Dependencies installed"
echo ""

# Create .env file
if [ ! -f ".env" ]; then
    echo "⚙️  Creating configuration file..."
    cp .env.example .env
    echo "✅ Configuration file created (.env)"
    echo ""
    echo "⚠️  IMPORTANT: Edit .env file and set strong passwords before production!"
else
    echo "✅ Configuration file already exists"
fi
echo ""

# Check PostgreSQL
echo "🔍 Checking PostgreSQL..."
if command -v psql &> /dev/null; then
    echo "✅ PostgreSQL found"
else
    echo "⚠️  PostgreSQL not found. You can:"
    echo "   1. Install PostgreSQL locally, or"
    echo "   2. Use Docker (see docker-compose.yml), or"
    echo "   3. Use a hosted PostgreSQL service"
fi
echo ""

# Check Redis
echo "🔍 Checking Redis..."
if command -v redis-cli &> /dev/null; then
    echo "✅ Redis found"
else
    echo "⚠️  Redis not found. You can:"
    echo "   1. Install Redis locally, or"
    echo "   2. Use Docker (see docker-compose.yml)"
fi
echo ""

echo "============================================"
echo "✅ Setup complete!"
echo "============================================"
echo ""
echo "📖 ЧТО ДЕЛАТЬ ДАЛЬШЕ?"
echo ""
echo "🇷🇺 Подробная инструкция на русском:"
echo "   📄 ../ЧТО_ДЕЛАТЬ_ДАЛЬШЕ.md"
echo "   📄 ../ШПАРГАЛКА.md (быстрая справка)"
echo ""
echo "🇬🇧 English guide:"
echo "   📄 ../README_WEBSSH.md"
echo ""
echo "⚡ Быстрый старт:"
echo ""
echo "1. Установите PostgreSQL и Redis"
echo "   (см. ЧТО_ДЕЛАТЬ_ДАЛЬШЕ.md)"
echo ""
echo "2. Запустите сервер:"
echo "   ./start.sh"
echo ""
echo "3. Откройте браузер:"
echo "   http://localhost:5000"
echo ""
echo "4. Войдите:"
echo "   Email: admin@example.com"
echo "   Password: admin123"
echo ""
echo "============================================"
