.PHONY: setup fmt vet build generate sync-types

# Первоначальная настройка проекта (запускается один раз после клонирования)
setup:
	git config core.hooksPath .githooks
	chmod +x .githooks/pre-commit
	@echo "✅ Проект настроен"

# Форматирование кода
fmt:
	gofmt -w .

# Проверка кода
vet:
	go vet ./...

# Сборка
build:
	go build ./...

# Генерация TypeScript типов из Go structs
generate:
	tygo generate
	@echo "✅ Типы сгенерированы → generated/types.ts"

# Генерация и синхронизация типов в оба фронтовых проекта
sync-types: generate
	@ADMIN=../vp-bot-admin/src/generated; \
	APP=../vp-bot-app/src/generated; \
	mkdir -p $$ADMIN $$APP; \
	cp generated/types.ts $$ADMIN/types.ts; \
	cp generated/types.ts $$APP/types.ts; \
	echo "✅ Типы скопированы в vp-bot-admin и vp-bot-app"
