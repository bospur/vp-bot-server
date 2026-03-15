# Бэкенд

## Архитектура

```
main.go                  — точка входа, инициализация, роуты
internal/
├── db/
│   └── db.go            — подключение к БД, запуск миграций
├── repository/
│   ├── animals.go       — SQL запросы: animals, categories
│   └── articles.go      — SQL запросы: articles, article_categories
├── handler/
│   ├── animals.go       — HTTP хендлеры: GET animals, categories
│   ├── articles.go      — HTTP хендлеры: GET articles
│   └── admin.go         — HTTP хендлеры: авторизация, CRUD
├── middleware/
│   └── auth.go          — JWT middleware
└── bot/
    └── bot.go           — Telegram бот
migrations/
├── 001_create_animals.up/down.sql
├── 002_create_categories.up/down.sql
├── 003_create_articles.up/down.sql
└── 004_add_multitenancy.up/down.sql
```

## Слои приложения

```
HTTP запрос
    │
    ▼
middleware (Auth) — проверка JWT для /api/admin/*
    │
    ▼
handler — читает запрос, вызывает repository, отвечает JSON
    │
    ▼
repository — SQL запросы к БД
    │
    ▼
PostgreSQL
```

## Публичный API

| Метод | URL | Описание |
|-------|-----|----------|
| GET | `/api/animals` | Список животных |
| GET | `/api/animals/{slug}/categories` | Категории животного |
| GET | `/api/animals/{animalSlug}/categories/{categorySlug}/articles` | Статьи категории |
| GET | `/api/articles/{slug}` | Одна статья |

## Admin API

| Метод | URL | Защита | Описание |
|-------|-----|--------|----------|
| POST | `/api/admin/login` | — | Получить JWT токен |
| POST | `/api/admin/animals` | JWT | Создать животное |
| PUT | `/api/admin/animals/{id}` | JWT | Обновить животное |
| DELETE | `/api/admin/animals/{id}` | JWT | Удалить животное |
| POST | `/api/admin/articles` | JWT | Создать статью |
| PUT | `/api/admin/articles/{id}` | JWT | Обновить статью |
| DELETE | `/api/admin/articles/{id}` | JWT | Удалить статью |
| POST | `/api/admin/articles/{id}/categories/{categoryId}` | JWT | Привязать статью к категории |
| DELETE | `/api/admin/articles/{id}/categories/{categoryId}` | JWT | Отвязать статью от категории |

## Авторизация

JWT токен передаётся в заголовке:
```
Authorization: Bearer <token>
```
Токен действует 24 часа.

## Мультитенантность

Каждая сущность привязана к `clinic_id`.

Публичный API: клиника определяется по slug в URL.
Admin API: клиника определяется из JWT токена пользователя.

## Переменные окружения

| Переменная | Описание |
|-----------|----------|
| `DATABASE_URL` | Строка подключения к PostgreSQL |
| `TELEGRAM_BOT_TOKEN` | Токен Telegram бота |
| `ADMIN_LOGIN` | Логин администратора (временно, будет заменён на таблицу users) |
| `ADMIN_PASSWORD` | Пароль администратора (временно) |
| `JWT_SECRET` | Секрет для подписи JWT токенов |
