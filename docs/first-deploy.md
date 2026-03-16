# Первый деплой (выполняется один раз)

## 1. Настройка SSH ключей

На своём Mac генерируем ключ специально для деплоя:

```bash
ssh-keygen -t ed25519 -C "deploy@vp-bot" -f ~/.ssh/vp_bot_deploy
```

Это создаст два файла:
- `~/.ssh/vp_bot_deploy` — приватный ключ (добавим в GitHub Secrets)
- `~/.ssh/vp_bot_deploy.pub` — публичный ключ (добавим на сервер)

## 2. Добавляем публичный ключ на сервер

```bash
# Копируем публичный ключ на сервер
ssh-copy-id -i ~/.ssh/vp_bot_deploy.pub deploy@194.87.0.94

# Проверяем что подключение работает
ssh -i ~/.ssh/vp_bot_deploy deploy@194.87.0.94
```

## 3. Добавляем Secrets в GitHub

Открываем: `https://github.com/bospur/vp-bot-server/settings/secrets/actions`

Добавляем три секрета:
- `VPS_HOST` → `194.87.0.94`
- `VPS_USER` → `deploy`
- `VPS_SSH_KEY` → содержимое файла `~/.ssh/vp_bot_deploy` (приватный ключ)

```bash
# Скопировать содержимое приватного ключа в буфер обмена
cat ~/.ssh/vp_bot_deploy | pbcopy
```

## 4. Первый раз клонируем репо на сервере

```bash
# Подключаемся к серверу
ssh -i ~/.ssh/vp_bot_deploy deploy@194.87.0.94

# Клонируем репо
git clone https://github.com/bospur/vp-bot-server.git /home/deploy/vp-bot-server

cd /home/deploy/vp-bot-server
```

## 5. Создаём .env на сервере

```bash
# На сервере, в папке проекта
cp .env.example .env
nano .env   # заполняем реальными значениями
```

## 6. Получаем SSL сертификат (Certbot)

```bash
# Устанавливаем Certbot на сервер (выполняется от root или через sudo)
sudo apt install certbot -y

# Временно останавливаем nginx если запущен
sudo systemctl stop nginx 2>/dev/null || true

# Получаем сертификат (standalone режим — Certbot поднимает временный сервер)
sudo certbot certonly --standalone -d snzbeachvolleyball25.ru

# Проверяем что сертификаты создались
sudo ls /etc/letsencrypt/live/snzbeachvolleyball25.ru/
```

## 7. Запускаем всё

```bash
# В папке проекта на сервере
docker compose up --build -d

# Проверяем что контейнеры запустились
docker compose ps

# Смотрим логи
docker compose logs -f
```

## 8. Проверяем

Открываем в браузере: `https://snzbeachvolleyball25.ru/hello`

Должен вернуть: `{"message":"Hello from Go server"}`
