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
│   ├── doctors.go       — SQL запросы: doctors, schedules, exceptions, settings
│   ├── users.go         — SQL запросы: users
│   └── grooming.go      — SQL запросы: grooming_breeds, template, appointments
├── handler/
│   ├── animals.go       — HTTP хендлеры: GET animals, categories (публичный)
│   ├── articles.go      — HTTP хендлеры: GET articles (публичный)
│   ├── admin.go         — HTTP хендлеры: авторизация, CRUD + admin GET
│   ├── doctors.go       — HTTP хендлеры: врачи, расписание, исключения, настройки
│   ├── grooming.go      — HTTP хендлеры: груминг (породы, шаблон, записи)
│   └── helpers.go       — общие утилиты (writeJSON)
├── middleware/
│   ├── auth.go          — JWT middleware
│   ├── auth_test.go     — unit тесты
│   └── cors.go          — CORS middleware
└── bot/
    ├── bot.go           — Telegram бот (хендлеры, навигация, setChatMenuButton)
    ├── doctors.go       — команды врачей и расписания
    └── htmlformat.go    — конвертер HTML → Telegram HTML
migrations/
├── 001_create_animals.up/down.sql
├── 002_create_categories.up/down.sql
├── 003_create_articles.up/down.sql
├── 004_add_multitenancy.up/down.sql
├── 005_add_article_status.up/down.sql
├── 006_create_doctors.up/down.sql   — doctors, doctor_schedules, exceptions, clinic_settings
└── 007_create_grooming.up/down.sql  — grooming_breeds, grooming_weekly_template, grooming_appointments
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
| GET | `/api/clinics/{clinicSlug}/animals/{a}/categories/{c}/articles` | Статьи категории |
| GET | `/api/clinics/{clinicSlug}/articles/{slug}` | Одна статья |
| GET | `/api/clinics/{clinicSlug}/doctors` | Опубликованные врачи |
| GET | `/api/clinics/{clinicSlug}/schedule` | Расписание на период `{ entries, settings }` |
| GET | `/uploads/{filename}` | Фото врачей (статика) |

## Admin API

### Авторизация

| Метод | URL | Описание |
|-------|-----|----------|
| POST | `/api/admin/login` | Получить JWT токен |

### Животные, категории, статьи

| Метод | URL | Описание |
|-------|-----|----------|
| GET/POST | `/api/admin/animals` | Список / создать |
| PUT/DELETE | `/api/admin/animals/{id}` | Обновить / удалить |
| GET/POST | `/api/admin/categories` | Список / создать |
| PUT/DELETE | `/api/admin/categories/{id}` | Обновить / удалить |
| GET/POST | `/api/admin/articles` | Список / создать |
| GET/PUT/DELETE | `/api/admin/articles/{id}` | Один / обновить / удалить |
| PATCH | `/api/admin/articles/{id}/status` | Сменить статус (только admin) |
| POST/DELETE | `/api/admin/articles/{id}/categories/{categoryId}` | Привязать / отвязать категорию |

### Врачи и расписание

| Метод | URL | Описание |
|-------|-----|----------|
| GET/POST | `/api/admin/doctors` | Список / создать |
| GET/PUT/DELETE | `/api/admin/doctors/{id}` | Один / обновить / удалить |
| PATCH | `/api/admin/doctors/{id}/status` | Сменить статус (только admin) |
| POST | `/api/admin/doctors/{id}/photo` | Загрузить фото (multipart, 5 МБ) |
| GET/POST | `/api/admin/doctors/{id}/schedule` | Слоты расписания / добавить |
| DELETE | `/api/admin/doctors/{id}/schedule/{slotId}` | Удалить слот |
| GET/PUT | `/api/admin/doctors/{id}/schedule/exceptions` | Исключения / upsert |
| DELETE | `/api/admin/doctors/{id}/schedule/exceptions/{exceptionId}` | Удалить исключение |
| GET/PATCH | `/api/admin/settings` | Настройки клиники (schedule_display_weeks) |

### Груминг

| Метод | URL | Описание |
|-------|-----|----------|
| GET/POST | `/api/admin/grooming/breeds` | Коллекция пород / создать |
| PUT/DELETE | `/api/admin/grooming/breeds/{id}` | Обновить / удалить породу |
| GET | `/api/admin/grooming/template` | Шаблон рабочей недели |
| PUT | `/api/admin/grooming/template` | Upsert слота (по `day_of_week`) |
| DELETE | `/api/admin/grooming/template/{dayOfWeek}` | Удалить рабочий день |
| GET | `/api/admin/grooming/appointments?month=YYYY-MM` | Записи за месяц |
| POST | `/api/admin/grooming/appointments` | Создать запись |
| DELETE | `/api/admin/grooming/appointments/{id}` | Удалить запись |

#### Бизнес-логика POST /appointments

1. Берёт `duration` из `grooming_breeds`
2. Проверяет что `date` — рабочий день по шаблону (`EXTRACT(DOW FROM date)`) → 400 если нет
3. Вычисляет `end_time = start_time + duration * interval '1 minute'` через PostgreSQL
4. Проверяет что `start_time >= time_from` и `end_time <= time_to` → 400 если за пределами
5. Проверяет пересечение: `start_time < end_existing AND end_time > start_existing` → 409 если занято

### Пользователи

| Метод | URL | Описание |
|-------|-----|----------|
| GET/POST | `/api/admin/users` | Список / создать (только admin) |
| DELETE | `/api/admin/users/{id}` | Удалить (только admin) |

## Авторизация

JWT токен передаётся в заголовке:
```
Authorization: Bearer <token>
```
Токен действует 24 часа и содержит `user_id`, `clinic_id`, `role`.
Пароли хранятся в БД в виде bcrypt хешей.

## Форматы данных

### day_of_week
PostgreSQL-конвенция: `0=Вс, 1=Пн, 2=Вт, 3=Ср, 4=Чт, 5=Пт, 6=Сб`
Используется в `doctor_schedules` и `grooming_weekly_template`.

### Контент статей
Поле `content` хранит HTML-строку от TipTap.
Бот конвертирует HTML → Telegram HTML через `htmlToTelegram()` в `internal/bot/htmlformat.go`.

## Первый пользователь

При первом запуске если таблица `users` пуста — создаётся admin пользователь
из переменных `ADMIN_LOGIN` и `ADMIN_PASSWORD`.

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
| `PUBLIC_URL` | нет | Базовый URL для фото врачей |
| `APP_URL` | нет | URL Telegram Mini App для кнопки в боте |
| `UPLOADS_DIR` | нет | Путь к папке загрузок (default: ./uploads) |
| `ADMIN_LOGIN` | только при первом запуске | Логин первого admin пользователя |
| `ADMIN_PASSWORD` | только при первом запуске | Пароль первого admin пользователя |
