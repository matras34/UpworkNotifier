@echo off
REM Web SSH Service - Start Script (Windows)

echo Starting Web SSH Service...
echo.

REM Check if virtual environment exists
if not exist "venv" (
    echo ERROR: Virtual environment not found. Please run setup.bat first.
    pause
    exit /b 1
)

REM Activate virtual environment
call venv\Scripts\activate.bat

REM Run the server
python run.py

pause
