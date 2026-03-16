# Бэкенд

## Архитектура

```
main.go                  — точка входа, инициализация, роуты, seed пользователя
internal/
├── db/
│   └── db.go            — подключение к БД, запуск миграций
├── repository/
│   ├── animals.go       — SQL запросы: animals, categories (CRUD)
│   ├── articles.go      — SQL запросы: articles, article_categories (CRUD)
│   └── users.go         — SQL запросы: users
├── handler/
│   ├── animals.go       — HTTP хендлеры: GET animals, categories (публичный)
│   ├── articles.go      — HTTP хендлеры: GET articles (публичный)
│   ├── admin.go         — HTTP хендлеры: авторизация, CRUD + admin GET
│   └── helpers.go       — общие утилиты (writeJSON)
├── middleware/
│   ├── auth.go          — JWT middleware
│   ├── auth_test.go     — unit тесты
│   └── cors.go          — CORS middleware
└── bot/
    ├── bot.go           — Telegram бот (хендлеры, навигация)
    └── htmlformat.go    — конвертер HTML → Telegram HTML
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
middleware (CORS, Auth) — CORS для всех, JWT для /api/admin/*
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
| GET | `/api/clinics/{clinicSlug}/animals` | Список животных клиники |
| GET | `/api/clinics/{clinicSlug}/animals/{slug}/categories` | Категории животного |
| GET | `/api/clinics/{clinicSlug}/animals/{animalSlug}/categories/{categorySlug}/articles` | Статьи категории |
| GET | `/api/clinics/{clinicSlug}/articles/{slug}` | Одна статья |

## Admin API

| Метод | URL | Защита | Описание |
|-------|-----|--------|----------|
| POST | `/api/admin/login` | — | Получить JWT токен |
| GET | `/api/admin/articles` | JWT | Все статьи клиники |
| GET | `/api/admin/articles/{id}` | JWT | Одна статья по id |
| GET | `/api/admin/articles/{id}/categories` | JWT | Категории статьи |
| POST | `/api/admin/animals` | JWT | Создать животное |
| PUT | `/api/admin/animals/{id}` | JWT | Обновить животное |
| DELETE | `/api/admin/animals/{id}` | JWT | Удалить животное |
| POST | `/api/admin/categories` | JWT | Создать категорию |
| PUT | `/api/admin/categories/{id}` | JWT | Обновить категорию |
| DELETE | `/api/admin/categories/{id}` | JWT | Удалить категорию |
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
Токен действует 24 часа и содержит `user_id`, `clinic_id`, `role`.
Пароли хранятся в БД в виде bcrypt хешей.

## Формат контента статей

Поле `content` хранит HTML-строку, генерируемую TipTap в админке.
Бот конвертирует HTML → Telegram HTML через `htmlToTelegram()` в `internal/bot/htmlformat.go`.

Поддерживаемые теги: `<h1>–<h3>`, `<p>`, `<strong>`, `<em>`, `<s>`, `<ul>`, `<ol>`, `<li>`, `<code>`, `<pre>`, `<br>`.

## Первый пользователь

При первом запуске если таблица `users` пуста — создаётся admin пользователь
из переменных `ADMIN_LOGIN` и `ADMIN_PASSWORD`. После создания env переменные
можно убрать.

## Мультитенантность

- Публичный API: клиника по `clinicSlug` в URL
- Admin API: клиника из JWT (`clinic_id`)
- Telegram бот: клиника из `CLINIC_SLUG` env

## Переменные окружения

| Переменная | Обязательная | Описание |
|-----------|-------------|----------|
| `DATABASE_URL` | да | Строка подключения к PostgreSQL |
| `TELEGRAM_BOT_TOKEN` | да | Токен Telegram бота |
| `CLINIC_SLUG` | да | Slug клиники для Telegram бота |
| `JWT_SECRET` | да | Секрет для подписи JWT токенов |
| `ADMIN_LOGIN` | только при первом запуске | Логин первого admin пользователя |
| `ADMIN_PASSWORD` | только при первом запуске | Пароль первого admin пользователя |
