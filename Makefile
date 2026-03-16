.PHONY: setup fmt vet build

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
