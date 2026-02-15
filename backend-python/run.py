#!/usr/bin/env python3
"""
Web SSH Service - Python Backend
Simple script to run the server
"""

import sys
import os

# Add app directory to path
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from app import create_app, socketio
from config import Config

if __name__ == '__main__':
    print("=" * 60)
    print("  Web SSH Service - Python Backend")
    print("=" * 60)
    print()
    print(f"Starting server on {Config.HOST}:{Config.PORT}")
    print(f"Debug mode: {Config.DEBUG}")
    print()
    print("Default login: admin@example.com / admin123")
    print()
    print("Press Ctrl+C to stop")
    print("=" * 60)
    print()
    
    app = create_app()
    
    # Run with SocketIO
    socketio.run(
        app,
        host=Config.HOST,
        port=Config.PORT,
        debug=Config.DEBUG,
        use_reloader=Config.DEBUG
    )
