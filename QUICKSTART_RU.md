# 🎯 БЫСТРЫЙ СТАРТ ЗА 2 МИНУТЫ

## Вариант 1: Docker (Рекомендуется) 🐳

### Шаг 1: Установите Docker
- Скачайте: https://www.docker.com/products/docker-desktop
- Установите и запустите Docker Desktop

### Шаг 2: Запустите приложение
**Windows:**
1. Дважды кликните `install.bat`
2. Дождитесь окончания установки

**Mac/Linux:**
1. Откройте терминал
2. `./install.sh`

### Шаг 3: Готово!
Откройте http://localhost в браузере

---

## Вариант 2: Python (Без Docker) 🐍

### Шаг 1: Установите нужное ПО
1. **Python 3.11+**: https://www.python.org/downloads/
2. **PostgreSQL**: https://www.postgresql.org/download/

### Шаг 2: Запустите
**Windows:**
```
cd backend-python
setup.bat
start.bat
```

**Mac/Linux:**
```bash
cd backend-python
./setup.sh
./start.sh
```

### Шаг 3: Готово!
Откройте http://localhost:5000 в браузере

---

## 🔐 Первый вход

```
Email:    admin@example.com
Password: admin123
```

**⚠️ Сразу смените пароль!**

---

## 💡 Как подключиться к серверу

1. Нажмите **"+ New Connection"**
2. Заполните:
   - Название: `Мой сервер`
   - Хост: `192.168.1.100` (ваш IP)
   - Порт: `22`
   - Логин: `root` (ваш логин)
   - Пароль: `******` (ваш пароль)
3. Нажмите **"Save and Connect"**
4. **Готово!** Терминал откроется

---

## 🆘 Проблемы?

### Docker не запускается
- Убедитесь что Docker Desktop запущен
- Перезагрузите компьютер

### Python версия
- Нужен Python 3.8 или новее
- Проверьте: `python --version`

### База данных
- Убедитесь что PostgreSQL запущен
- Порт 5432 должен быть свободен

### Порт занят
В файле `.env` измените:
```
PORT=5001
```

---

## 📚 Больше информации

- Полная инструкция: `README_RU.md`
- English version: `README_WEBSSH.md`
- Для разработчиков: `backend-python/README.md`

---

## ✨ Возможности

✅ SSH терминал в браузере  
✅ Работает на Windows/Mac/Linux  
✅ Сохранение подключений  
✅ Безопасное хранение паролей  
✅ Работает с телефона  
✅ Многопользовательский режим  

---

**Сделано просто! 🚀**
