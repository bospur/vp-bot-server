# Кодогенерация TypeScript типов

## Зачем

В проекте три репозитория: Go бэкенд и два TypeScript фронта (admin, app).
Go structs из `internal/repository/` дублировались вручную в обоих фронтах — одни и те же поля писались три раза. При изменении модели нужно было помнить обновить типы в двух местах, что легко забыть.

**Решение:** генерировать TypeScript типы из Go structs автоматически и синхронизировать в оба фронта одной командой.

---

## Технология: tygo

[tygo](https://github.com/gzuidhof/tygo) — инструмент, который читает Go пакеты и генерирует TypeScript интерфейсы из struct-ов, используя json-теги как имена полей.

**Почему tygo, а не OpenAPI:**
- OpenAPI требует поддерживать отдельную спецификацию (YAML/JSON) или аннотировать весь код комментариями для `swaggo`
- tygo читает код напрямую — нет отдельного слоя документации
- На текущем этапе нужны только типы, без валидации запросов в рантайме
- Переход на OpenAPI в будущем не потребует переписывать фронты — типы совместимы

---

## Конфигурация

Файл `tygo.yaml` в корне `vp-bot-server`:

```yaml
packages:
  - path: "go-server/internal/repository"
    output_path: "generated/types.ts"
    type_mappings:
      time.Time: "string"
    exclude_types:
      - AnimalRepository
      - ArticleRepository
      - DoctorRepository
      - GroomingRepository
      - UserRepository
```

- `path` — Go пакет, из которого читаются structs
- `output_path` — куда генерировать TypeScript файл
- `type_mappings` — маппинг Go типов в TS (time.Time → string)
- `exclude_types` — исключить служебные struct-ы репозиториев (пустые интерфейсы в TS не нужны)

> **Известная проблема:** `exclude_types` в tygo v0.2.21 не исключает все типы корректно — часть Repository интерфейсов остаётся в выводе как пустые `interface {}`. Они не вредят: фронты их не импортируют, TS компилятор на них не ругается.

---

## Makefile команды

```makefile
make generate     # только генерация → generated/types.ts
make sync-types   # генерация + копирование в оба фронта
```

`sync-types` копирует `generated/types.ts` в:
- `../vp-bot-admin/src/generated/types.ts`
- `../vp-bot-app/src/generated/types.ts`

Предполагает что все три репозитория лежат рядом в одной папке.

---

## Установка tygo

```bash
go install github.com/gzuidhof/tygo@latest
```

После установки бинарник будет в `~/go/bin/tygo`. Убедись что `~/go/bin` в `PATH`:

```bash
export PATH=$PATH:~/go/bin
```

---

## Что генерируется

Из пакета `internal/repository` генерируются типы:

| Go struct | TypeScript interface | Где используется |
|---|---|---|
| `Animal` | `Animal` | admin, app |
| `Category` | `Category` | admin, app |
| `AnimalInput` | `AnimalInput` | admin |
| `CategoryInput` | `CategoryInput` | admin |
| `Article` | `Article` | admin, app |
| `ArticleInput` | `ArticleInput` | admin |
| `Doctor` | `Doctor` | admin, app |
| `DoctorInput` | `DoctorInput` | admin |
| `DoctorSchedule` | `DoctorSchedule` | admin |
| `DoctorScheduleException` | `DoctorScheduleException` | admin |
| `ClinicSettings` | `ClinicSettings` | admin |
| `ScheduleEntry` | `ScheduleEntry` | admin, app |
| `GroomingBreed` | `GroomingBreed` | admin |
| `GroomingBreedInput` | `GroomingBreedInput` | admin |
| `GroomingTemplateSlot` | `GroomingTemplateSlot` | admin |
| `GroomingTemplateInput` | `GroomingTemplateInput` | admin |
| `GroomingAppointment` | `GroomingAppointment` | admin |
| `GroomingAppointmentInput` | `GroomingAppointmentInput` | admin |
| `User` | `User` | admin |

---

## Где лежат сгенерированные файлы

```
vp-bot-server/generated/types.ts   ← генерируется, в .gitignore
vp-bot-admin/src/generated/types.ts  ← копируется, хранится в git
vp-bot-app/src/generated/types.ts    ← копируется, хранится в git
```

`generated/` в `vp-bot-server` добавлен в `.gitignore` — генерировать нужно локально.
В фронтах файл **хранится в git** — это позволяет видеть diff при изменении моделей в PR.

---

## Процесс при изменении модели

1. Изменить Go struct в `internal/repository/*.go`
2. Запустить из корня `vp-bot-server`:
   ```bash
   make sync-types
   ```
3. Проверить diff в `src/generated/types.ts` в обоих фронтах
4. Обновить использования типов в коде если изменились поля
5. Закоммитить изменения в каждом репо

---

## Возможный следующий шаг: OpenAPI

Когда команда вырастет или появится потребность в документации API и валидации запросов в рантайме — можно перейти на OpenAPI:

- `swaggo/swag` — генерирует OpenAPI spec из Go аннотаций
- `openapi-typescript` — генерирует TS типы из OpenAPI spec
- Типы будут совместимы с текущими — миграция фронтов сведётся к переименованию импортов

Это не срочно: tygo покрывает текущие потребности и не создаёт технического долга.
