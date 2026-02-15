@echo off
REM Web SSH Service - Easy Setup Script for Windows

echo ============================================
echo   Web SSH Service - Easy Setup (Windows)
echo ============================================
echo.

REM Check Python
echo Checking Python version...
python --version >nul 2>&1
if errorlevel 1 (
    echo ERROR: Python is not installed!
    echo Please install Python 3.8 or higher from https://www.python.org/
    pause
    exit /b 1
)

for /f "tokens=2" %%i in ('python --version 2^>^&1') do set PYTHON_VERSION=%%i
echo Found Python %PYTHON_VERSION%
echo.

REM Create virtual environment
echo Creating virtual environment...
if not exist "venv" (
    python -m venv venv
    echo Virtual environment created
) else (
    echo Virtual environment already exists
)
echo.

REM Activate virtual environment
echo Activating virtual environment...
call venv\Scripts\activate.bat
echo.

REM Install dependencies
echo Installing dependencies (this may take a few minutes)...
python -m pip install --upgrade pip >nul 2>&1
pip install -r requirements.txt
echo Dependencies installed
echo.

REM Create .env file
if not exist ".env" (
    echo Creating configuration file...
    copy .env.example .env
    echo Configuration file created (.env)
    echo.
    echo WARNING: Edit .env file and set strong passwords before production!
) else (
    echo Configuration file already exists
)
echo.

echo ============================================
echo Setup complete!
echo ============================================
echo.
echo Next steps:
echo.
echo 1. Make sure PostgreSQL and Redis are running
echo 2. Run the server: start.bat
echo.
echo 3. Access the application:
echo    http://localhost:5000
echo.
echo 4. Default login:
echo    Email: admin@example.com
echo    Password: admin123
echo.
echo ============================================
pause
