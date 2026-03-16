# Архитектура проекта

## Репозитории

| Репо | Назначение | Стек |
|------|-----------|------|
| `vp-bot-server` | Бэкенд + Telegram бот | Go, PostgreSQL |
| `vp-bot-app` | Telegram Mini App | Vite, React |
| `vp-bot-admin` | Админ панель | Vite, React, MUI |

## Схема системы

```
Telegram
   │
   ├── Mini App (vp-bot-app)
   │      │
   │      └── GET /api/*
   │
   └── Bot (long polling)
          │
          └── внутри Go сервера

Браузер (врачи)
   │
   └── Admin Panel (vp-bot-admin)
          │
          ├── POST /api/admin/login
          └── /api/admin/* (JWT)

Интернет
   │
   ▼
Nginx (системный, Ubuntu)
   │
   ├── snzbeachvolleyball25.ru → pocketbase:8090 (VPN проект)
   └── api.snzbeachvolleyball25.ru → Go app:8080 (этот проект)
            │
            ├── Go HTTP сервер
            └── PostgreSQL (Docker)
```

## Мультитенантность

```
clinics
  └── users (admin/editor/viewer)
  └── animals
       └── categories
            └── article_categories
                     └── articles
```

Публичный API: `/api/clinics/{clinicSlug}/animals`
Admin API: клиника из JWT токена

## CI/CD

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
