# План миграции в монорепо

## Контекст

Текущая структура — три отдельных репозитория:
- `vp-bot-server` — Go бэкенд + Telegram бот
- `vp-bot-admin` — Админ панель (Vite + React + MUI)
- `vp-bot-app` — Telegram Mini App (Vite + React + TG UI)

Цель — объединить в один приватный монорепозиторий с Turborepo.

**Почему монорепо:**
- Кросс-платформенные фичи в одном PR и одном деплое
- Агентам (Claude, Copilot) проще работать с единой кодовой базой
- Локальная разработка через один `docker-compose up`
- Общий пакет типов `packages/types` без внешних инструментов
- Команда фулстеков — не нужно переключаться между репами

**Почему приватная репа:**
Продукт планируется к продаже как SaaS для ветклиник — исходники должны быть закрыты.

---

## Целевая структура

```
vp-bot/                          ← новый монорепозиторий
├── apps/
│   ├── server/                  ← Go бэкенд (из vp-bot-server)
│   ├── admin/                   ← Админ панель (из vp-bot-admin)
│   └── app/                     ← Mini App (из vp-bot-app)
├── packages/
│   └── types/                   ← общие TypeScript типы
│       ├── package.json
│       └── src/
│           ├── index.ts
│           ├── animals.ts
│           ├── articles.ts
│           ├── doctors.ts
│           ├── grooming.ts
│           └── users.ts
├── turbo.json                   ← конфигурация Turborepo
├── package.json                 ← root package.json (workspaces)
├── docker-compose.yml           ← единый запуск всего локально
├── .github/
│   └── workflows/               ← CI/CD для каждого приложения
└── .gitignore
```

---

## Этапы миграции

### Этап 0. Подготовка (до начала миграции)
**Цель:** чистое состояние всех трёх репо перед переносом.

- [ ] Смержить все открытые PR в `dev` во всех трёх репо
- [ ] Убедиться что `dev` задеплоен и всё работает на проде
- [ ] Закрыть/отменить ветки `chore/tygo-codegen` — tygo не переносим
- [ ] Сделать финальный `git pull` во всех трёх репо

---

### Этап 1. Создание монорепо
**Цель:** пустой репозиторий с правильной структурой и Turborepo.

**1.1 Создать новый приватный репозиторий** на GitHub (`vp-bot` или другое название).

**1.2 Инициализировать структуру локально:**
```bash
mkdir vp-bot && cd vp-bot
git init
git remote add origin <url новой репы>
```

**1.3 Создать root `package.json` с workspaces:**
```json
{
  "name": "vp-bot",
  "private": true,
  "workspaces": ["apps/*", "packages/*"],
  "scripts": {
    "dev": "turbo dev",
    "build": "turbo build",
    "lint": "turbo lint"
  },
  "devDependencies": {
    "turbo": "latest"
  }
}
```

**1.4 Создать `turbo.json`:**
```json
{
  "$schema": "https://turbo.build/schema.json",
  "tasks": {
    "build": {
      "dependsOn": ["^build"],
      "outputs": ["dist/**"]
    },
    "dev": {
      "cache": false,
      "persistent": true
    },
    "lint": {}
  }
}
```

**1.5 Установить Turborepo:**
```bash
npm install
```

---

### Этап 2. Перенос приложений
**Цель:** перенести код трёх репо в папки `apps/`, сохранив git-историю.

Для каждого приложения используем `git subtree` — это сохраняет всю историю коммитов.

**2.1 Добавить remote-ы старых репо:**
```bash
git remote add server https://github.com/bospur/vp-bot-server.git
git remote add admin  https://github.com/bospur/vp-bot-admin.git
git remote add app    https://github.com/bospur/vp-bot-app.git
```

**2.2 Перенести каждое приложение через subtree:**
```bash
git fetch server && git subtree add --prefix=apps/server server dev --squash
git fetch admin  && git subtree add --prefix=apps/admin  admin  dev --squash
git fetch app    && git subtree add --prefix=apps/app    app    main --squash
```

> `--squash` сжимает историю в один коммит. Если хочешь сохранить полную историю — убери флаг, но merge-коммит будет большим.

**2.3 Проверить структуру:**
```
apps/server/  ← весь код vp-bot-server
apps/admin/   ← весь код vp-bot-admin
apps/app/     ← весь код vp-bot-app
```

---

### Этап 3. Создание пакета типов
**Цель:** вынести общие TypeScript типы в `packages/types`, убрать дублирование.

**3.1 Создать пакет:**
```bash
mkdir -p packages/types/src
```

**3.2 `packages/types/package.json`:**
```json
{
  "name": "@vp-bot/types",
  "version": "0.0.1",
  "private": true,
  "main": "./src/index.ts",
  "types": "./src/index.ts",
  "exports": {
    ".": "./src/index.ts"
  }
}
```

**3.3 Разбить типы по файлам** (взять из существующих domain/types.ts и сгенерированных tygo файлов):

- `src/animals.ts` — Animal, Category, AnimalInput, CategoryInput
- `src/articles.ts` — Article, ArticleInput, ArticleStatus
- `src/doctors.ts` — Doctor, DoctorInput, DoctorSchedule, DoctorScheduleException, ClinicSettings, ScheduleEntry
- `src/grooming.ts` — GroomingBreed, GroomingBreedInput, GroomingTemplateSlot, GroomingTemplateInput, GroomingAppointment, GroomingAppointmentInput
- `src/users.ts` — User (базовый тип, без role enum — роли специфичны для admin)
- `src/index.ts` — реэкспорт всего

**3.4 Подключить пакет в admin и app:**

В `apps/admin/package.json` и `apps/app/package.json` добавить:
```json
{
  "dependencies": {
    "@vp-bot/types": "*"
  }
}
```

**3.5 Заменить импорты** в обоих приложениях:
```typescript
// было
import type { Doctor } from '../modules/doctors/domain/types';

// стало
import type { Doctor } from '@vp-bot/types';
```

> Делать постепенно, файл за файлом. Не обязательно за один PR.

---

### Этап 4. Настройка локальной разработки
**Цель:** один `docker-compose up` запускает всё.

**4.1 Создать `docker-compose.yml` в корне монорепо:**
```yaml
services:
  db:
    image: postgres:16
    environment:
      POSTGRES_DB: vp_bot
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - pg_data:/var/lib/postgresql/data

  server:
    build: ./apps/server
    env_file: ./apps/server/.env.local
    ports:
      - "8080:8080"
    depends_on:
      - db

  admin:
    build: ./apps/admin
    ports:
      - "5173:5173"

  app:
    build: ./apps/app
    ports:
      - "5174:5174"

volumes:
  pg_data:
```

**4.2 Настроить `turbo dev`** чтобы запускал admin и app параллельно (server через Docker).

---

### Этап 5. CI/CD
**Цель:** деплой каждого приложения только при изменении его кода.

**5.1 Структура workflows:**
```
.github/workflows/
├── deploy-server.yml   ← триггер: apps/server/**
├── deploy-admin.yml    ← триггер: apps/admin/** или packages/types/**
└── deploy-app.yml      ← триггер: apps/app/** или packages/types/**
```

**5.2 Path-based триггеры:**
```yaml
on:
  push:
    branches: [main]
    paths:
      - 'apps/server/**'
      - '.github/workflows/deploy-server.yml'
```

> Это ключевой момент — без path filters при любом коммите будут деплоиться все три приложения.

**5.3 Перенести секреты** из трёх старых репо в новое:
- `VPS_HOST`, `VPS_USER`, `VPS_SSH_KEY`
- `VITE_API_URL` и другие env переменные

---

### Этап 6. Обновление старых репозиториев
**Цель:** заархивировать старые репо, чтобы не было путаницы.

- [ ] Добавить в README каждого старого репо: `⚠️ Репозиторий перенесён в vp-bot (монорепо)`
- [ ] Заархивировать репозитории на GitHub (Settings → Archive)
- [ ] Обновить ссылки в документации

---

## Порядок выполнения этапов

```
Этап 0  →  Этап 1  →  Этап 2  →  Этап 3  →  Этап 4  →  Этап 5  →  Этап 6
Подготовка  Структура  Перенос   Типы      Docker     CI/CD      Архив
  (1 день)  (1 день)  (1 день)  (2 дня)   (1 день)   (2 дня)    (1 час)
```

Итого: ~8-9 рабочих дней. Не срочно — можно делать параллельно с разработкой фич.

---

## Риски и как их митигировать

| Риск | Митигация |
|---|---|
| CI/CD не триггерится нужно | Протестировать path filters на тестовых коммитах до удаления старых репо |
| Потеря git истории | Использовать subtree без --squash или сохранить старые репо в архиве |
| Конфликты зависимостей Go + Node | Go не входит в npm workspaces, `apps/server` управляет своими зависимостями через `go.mod` независимо |
| Секреты в старых репо | Ротировать SSH ключи после переноса |

---

## Что НЕ переносим

- `tygo.yaml` и `generated/types.ts` — заменяет `packages/types`
- Отдельные `.github/workflows` из старых репо — пишем заново под монорепо структуру
