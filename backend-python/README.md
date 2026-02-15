# Web SSH - Python Backend (Простая установка для начинающих / Easy Setup for Beginners)

## 🚀 Быстрый старт / Quick Start

### Windows

1. **Скачайте Python** (если еще не установлен):
   - Перейдите на https://www.python.org/downloads/
   - Скачайте Python 3.11 или новее
   - При установке поставьте галочку "Add Python to PATH"

2. **Установите PostgreSQL** (база данных):
   - Скачайте с https://www.postgresql.org/download/windows/
   - Установите с паролем `postgres`
   - Запомните порт (обычно 5432)

3. **Установите Redis** (опционально, можно без него):
   - Скачайте с https://github.com/microsoftarchive/redis/releases
   - Или используйте Docker

4. **Запустите приложение**:
   ```cmd
   cd backend-python
   setup.bat
   start.bat
   ```

### Linux / Mac

1. **Установите зависимости**:
   ```bash
   # Ubuntu/Debian
   sudo apt-get update
   sudo apt-get install python3 python3-pip python3-venv postgresql redis-server
   
   # Mac (с Homebrew)
   brew install python postgresql redis
   ```

2. **Запустите приложение**:
   ```bash
   cd backend-python
   chmod +x setup.sh start.sh
   ./setup.sh
   ./start.sh
   ```

### Доступ / Access

После запуска откройте в браузере:
- **Приложение**: http://localhost:5000
- **Вход по умолчанию**:
  - Email: `admin@example.com`
  - Password: `admin123`

## 🐳 Docker (Самый простой способ / Easiest Way)

Если у вас установлен Docker:

```bash
# Из корневой папки проекта
docker-compose up -d

# Откройте http://localhost
```

Всё! База данных, Redis и приложение запустятся автоматически.

## 📁 Структура проекта / Project Structure

```
backend-python/
├── setup.sh / setup.bat    # Скрипт установки
├── start.sh / start.bat     # Скрипт запуска
├── run.py                   # Главный файл
├── requirements.txt         # Зависимости Python
├── .env.example            # Пример настроек
├── app/
│   ├── __init__.py         # Инициализация приложения
│   ├── api/                # API endpoints
│   │   ├── auth.py         # Авторизация
│   │   ├── connections.py  # SSH соединения
│   │   └── ssh.py          # WebSocket SSH
│   ├── models/             # Модели базы данных
│   ├── utils/              # Утилиты
│   └── ssh_handler/        # SSH клиент
└── config.py               # Конфигурация
```

## ⚙️ Настройка / Configuration

Отредактируйте файл `.env`:

```bash
# База данных
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=webssh

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# Безопасность (ИЗМЕНИТЕ В ПРОДАКШЕНЕ!)
SECRET_KEY=your-secret-key-here
JWT_SECRET_KEY=your-jwt-secret-here
ENCRYPTION_KEY=your-32-char-encryption-key!!!
```

## 🔧 Устранение проблем / Troubleshooting

### Ошибка: "Python not found"
- **Windows**: Переустановите Python с галочкой "Add to PATH"
- **Linux**: `sudo apt-get install python3`
- **Mac**: `brew install python3`

### Ошибка: "Cannot connect to database"
1. Проверьте, что PostgreSQL запущен:
   ```bash
   # Linux
   sudo systemctl status postgresql
   
   # Mac
   brew services list
   
   # Windows
   services.msc (найдите PostgreSQL)
   ```

2. Создайте базу данных вручную:
   ```bash
   psql -U postgres
   CREATE DATABASE webssh;
   \q
   ```

### Ошибка: "Port 5000 already in use"
Измените порт в файле `.env`:
```
PORT=5001
```

### Redis недоступен
Redis опционален. Если не нужен, закомментируйте в коде или используйте Docker:
```bash
docker run -d -p 6379:6379 redis:alpine
```

## 📝 Функциональность / Features

✅ Терминал в браузере (xterm.js)  
✅ WebSocket соединение  
✅ Подключение по паролю  
✅ Подключение по приватному ключу  
✅ Сохранение подключений  
✅ Автоматическая проверка хост-ключей  
✅ Таймаут сессии  
✅ Многопользовательский режим  
✅ Логирование всех действий  

## 🔐 Безопасность / Security

- ❌ **НЕ используйте** пароли по умолчанию в продакшене!
- ✅ Измените `SECRET_KEY`, `JWT_SECRET_KEY`, `ENCRYPTION_KEY`
- ✅ Используйте HTTPS в продакшене
- ✅ Установите firewall
- ✅ Регулярно обновляйте зависимости

## 📦 Зависимости / Dependencies

- **Flask** - веб-фреймворк
- **Flask-SocketIO** - WebSocket поддержка
- **Paramiko** - SSH клиент
- **SQLAlchemy** - ORM для базы данных
- **PostgreSQL** - база данных
- **Redis** - кэш и сессии

## 🆘 Помощь / Help

### Часто задаваемые вопросы

**Q: Как добавить нового пользователя?**  
A: Используйте endpoint `/api/v1/auth/register` или добавьте через базу данных

**Q: Можно ли использовать SQLite вместо PostgreSQL?**  
A: Да, но потребуется изменить `SQLALCHEMY_DATABASE_URI` в config.py

**Q: Как добавить поддержку HTTPS?**  
A: Используйте nginx как reverse proxy или настройте SSL в Flask

**Q: Работает ли это на Raspberry Pi?**  
A: Да! Установите зависимости и запустите как обычно

## 📚 Дополнительная документация / Additional Docs

- Полная документация: `README_WEBSSH.md`
- Архитектура: `ARCHITECTURE.md`
- Docker: `docker-compose.yml`

## 🎯 Что дальше? / What's Next?

1. Измените пароль администратора
2. Настройте SSL/HTTPS
3. Настройте резервное копирование базы данных
4. Добавьте мониторинг
5. Масштабируйте с помощью Docker Swarm или Kubernetes

## 📄 Лицензия / License

MIT

---

**Сделано с ❤️ для начинающих / Made with ❤️ for beginners**
