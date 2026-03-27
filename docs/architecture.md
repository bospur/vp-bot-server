# Архитектура проекта

## Репозитории

| Репо | Назначение | Стек |
|------|-----------|------|
| `vp-bot-server` | Бэкенд + Telegram бот | Go, PostgreSQL |
| `vp-bot-app` | Telegram Mini App | Vite, React 18, TanStack Query |
| `vp-bot-admin` | Админ панель | Vite, React, MUI |

## Схема системы

```
Telegram
   │
   ├── Mini App (vp-bot-app) ← app.snzbeachvolleyball25.ru
   │      │
   │      └── GET /api/clinics/*
   │
   └── Bot (long polling)
          │
          └── внутри Go сервера

Браузер (врачи / администраторы)
   │
   └── Admin Panel → https://admin.snzbeachvolleyball25.ru
          │
          ├── POST /api/admin/login
          └── /api/admin/* (JWT Bearer)

Интернет
   │
   ▼
Nginx (системный, Ubuntu VPS)
   │
   ├── admin.snzbeachvolleyball25.ru → /var/www/vp-bot-admin (статика)
   ├── app.snzbeachvolleyball25.ru   → /var/www/vp-bot-app (статика)
   └── api.snzbeachvolleyball25.ru   → Go app:8080
                                           │
                                           ├── Go HTTP сервер
                                           └── PostgreSQL (Docker)
```

## Структура данных

```
clinics
  └── users (admin/editor)
  └── animals
       └── categories
            └── article_categories (M2M)
                     └── articles  (content — HTML от TipTap)
  └── doctors
       └── doctor_schedules (еженедельные слоты)
       └── doctor_schedule_exceptions (исключения на дату)
  └── clinic_settings (schedule_display_weeks)
  └── grooming_breeds (коллекция пород, duration, price)
  └── grooming_weekly_template (рабочие дни грумера по шаблону)
  └── grooming_appointments (записи на конкретную дату)
```

Публичный API: клиника по `clinicSlug` в URL
Admin API: клиника из JWT (`clinic_id`)
Telegram бот: клиника из `CLINIC_SLUG` env

## CI/CD

### Backend (vp-bot-server)
```
git push origin dev
       │
       ▼
GitHub Actions
       │
       └── SSH → VPS
              ├── git pull origin dev
              └── docker compose up --build -d
```

### Frontend (vp-bot-admin / vp-bot-app)
```
git push origin dev
       │
       ▼
GitHub Actions
       │
       ├── npm ci
       ├── npm run build  (VITE_API_URL, VITE_CLINIC_SLUG из secrets)
       └── scp dist/ → VPS:/var/www/<repo>/
```

## Формат контента статей

Статьи хранятся как HTML-строка (генерируется TipTap в админке).
Telegram бот конвертирует HTML → Telegram HTML через `htmlToTelegram()` в `internal/bot/htmlformat.go`.
Mini App рендерит HTML напрямую.
