# Инфраструктура

## Обзор

```
Интернет
   │
   ▼
Nginx (системный, Ubuntu) :80/:443
   │
   ├── admin.snzbeachvolleyball25.ru → /var/www/vp-bot-admin (статика React)
   └── api.snzbeachvolleyball25.ru   → Go app:8080
                                           │
                                           ├── Go HTTP сервер (Docker)
                                           └── PostgreSQL :5432 (Docker, не доступен снаружи)
```

## Сервер

- **IP:** 194.87.0.94
- **OS:** Ubuntu
- **Deploy пользователь:** `deploy`
- **SSH ключ:** `~/.ssh/vp_bot_deploy` (локально, алиас `vps`)

## Пути на сервере

| Путь | Что там |
|------|---------|
| `/home/deploy/vp-bot-server` | Go бэкенд (git репо + docker compose) |
| `/var/www/vp-bot-admin` | Frontend статика (деплоится через scp) |
| `/etc/nginx/sites-available/` | Nginx конфиги |
| `/etc/letsencrypt/live/` | SSL сертификаты |
| `/var/www/certbot` | Webroot для Certbot проверки |

## Домены и nginx

| Домен | Назначение | Nginx конфиг |
|-------|-----------|--------------|
| `api.snzbeachvolleyball25.ru` | Go API + Telegram бот | `/etc/nginx/sites-available/api...` |
| `admin.snzbeachvolleyball25.ru` | Админ панель (статика) | `/etc/nginx/sites-available/admin...` |

## Docker (бэкенд)

| Сервис | Образ | Назначение |
|--------|-------|-----------|
| app | ./Dockerfile | Go HTTP сервер |
| db | postgres:16-alpine | База данных |

Переменные окружения хранятся в `.env` на сервере (не в git). Шаблон — `.env.example`.

## CI/CD — Backend

**Файл:** `.github/workflows/deploy.yml`

При пуше в `dev`:
1. GitHub Actions → SSH на VPS
2. `git pull origin dev`
3. `docker compose up --build -d`

### GitHub Secrets (vp-bot-server)

| Secret | Значение |
|--------|----------|
| `VPS_HOST` | 194.87.0.94 |
| `VPS_USER` | deploy |
| `VPS_SSH_KEY` | приватный SSH ключ (`~/.ssh/vp_bot_deploy`) |

## CI/CD — Frontend

**Файл:** `vp-bot-admin/.github/workflows/deploy.yml`

При пуше в `dev`:
1. GitHub Actions: `npm ci` → `npm run build`
2. `scp dist/` → VPS `/var/www/vp-bot-admin/`

### GitHub Secrets (vp-bot-admin)

| Secret | Значение |
|--------|----------|
| `VPS_HOST` | 194.87.0.94 |
| `VPS_USER` | deploy |
| `VPS_SSH_KEY` | приватный SSH ключ (`~/.ssh/vp_bot_deploy`) |
| `VITE_API_URL` | `https://api.snzbeachvolleyball25.ru` |
| `VITE_CLINIC_SLUG` | `default` |

## SSL сертификаты

Выдаются бесплатно через Let's Encrypt (Certbot). Обновляются автоматически каждые 90 дней.

Получить новый сертификат для субдомена:
```bash
sudo certbot certonly --webroot -w /var/www/certbot -d <subdomain>
```

## Первый деплой

См. [first-deploy.md](./first-deploy.md)
