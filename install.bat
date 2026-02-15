@echo off
REM Web SSH Service - One-Click Installer with Docker

echo ============================================
echo   Web SSH - Установка за 1 клик
echo   Web SSH - One-Click Installation
echo ============================================
echo.

REM Check if Docker is installed
docker --version >nul 2>&1
if errorlevel 1 (
    echo ERROR: Docker не установлен / Docker is not installed!
    echo.
    echo Скачайте Docker Desktop / Download Docker Desktop:
    echo https://www.docker.com/products/docker-desktop
    echo.
    pause
    exit /b 1
)

REM Check if docker-compose is installed
docker-compose --version >nul 2>&1
if errorlevel 1 (
    echo ERROR: docker-compose не установлен / docker-compose is not installed!
    echo.
    echo Обычно устанавливается вместе с Docker Desktop
    echo Usually installed with Docker Desktop
    echo.
    pause
    exit /b 1
)

echo Docker найден / Docker found
echo.

echo Запуск всех сервисов / Starting all services...
echo.
echo Это может занять несколько минут при первом запуске...
echo This may take a few minutes on first run...
echo.

REM Start services
docker-compose up -d

REM Wait for services
echo.
echo Ожидание запуска сервисов / Waiting for services...
timeout /t 10 /nobreak >nul

echo.
echo ============================================
echo Установка завершена! / Installation complete!
echo ============================================
echo.
echo Откройте в браузере / Open in browser:
echo    http://localhost
echo.
echo Вход по умолчанию / Default login:
echo    Email: admin@example.com
echo    Password: admin123
echo.
echo Управление / Management:
echo    Остановить / Stop:       docker-compose stop
echo    Перезапустить / Restart: docker-compose restart
echo    Удалить / Remove:        docker-compose down
echo    Логи / Logs:             docker-compose logs -f
echo.
echo ============================================
pause
