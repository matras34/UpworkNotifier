#!/bin/bash
# Web SSH Service - Start Script

echo "🚀 Starting Web SSH Service..."
echo ""

# Activate virtual environment
if [ ! -d "venv" ]; then
    echo "❌ Virtual environment not found. Please run setup.sh first."
    exit 1
fi

source venv/bin/activate

# Run the server
python3 run.py
