# Инфраструктура

## Обзор

```
Интернет
   │
   ▼
[Nginx :80/:443]  ← SSL termination, редирект HTTP→HTTPS
   │
   ▼
[Go app :8080]    ← бизнес-логика, API, Telegram webhook
   │
   ▼
[PostgreSQL :5432] ← база данных (недоступна снаружи)
```

Все сервисы работают в Docker контейнерах и общаются по внутренней сети `internal`.
Снаружи доступны только порты 80 и 443 (через Nginx).

## Сервер

- **IP:** 194.87.0.94
- **OS:** Ubuntu
- **Домен:** snzbeachvolleyball25.ru
- **Deploy пользователь:** deploy
- **Путь проекта:** /home/deploy/vp-bot-server

## Сервисы (docker-compose.yml)

| Сервис | Образ | Назначение |
|--------|-------|-----------|
| app | ./Dockerfile | Go HTTP сервер |
| db | postgres:16-alpine | База данных |
| nginx | nginx:alpine | Reverse proxy + SSL |

## Переменные окружения

Хранятся в файле `.env` на сервере (не в git).
Шаблон — `.env.example`.

## CI/CD

**Файл:** `.github/workflows/deploy.yml`

При пуше в ветку `dev`:
1. GitHub Actions подключается к VPS по SSH
2. Делает `git pull origin dev`
3. Запускает `docker compose up --build -d`

### GitHub Secrets (нужно настроить в репозитории)

| Secret | Значение |
|--------|----------|
| `VPS_HOST` | 194.87.0.94 |
| `VPS_USER` | deploy |
| `VPS_SSH_KEY` | приватный SSH ключ пользователя deploy |

## Первый деплой (выполняется вручную один раз)

См. [first-deploy.md](./first-deploy.md)

## SSL сертификат

Выдаётся бесплатно через Let's Encrypt (Certbot).
Автоматически обновляется каждые 90 дней.
