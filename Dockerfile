# ── Stage 1: сборка ──────────────────────────────────────────────────────────
# Берём официальный образ Go для компиляции
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Копируем файлы зависимостей и скачиваем их
# (отдельным слоем — Docker кэширует если go.mod не менялся)
COPY go.mod ./
RUN go mod download

# Копируем весь исходный код и собираем бинарник
COPY . .
RUN go build -o server .

# ── Stage 2: runtime ─────────────────────────────────────────────────────────
# Минимальный образ без Go toolchain — только бинарник
FROM alpine:3.19

WORKDIR /app

# Копируем собранный бинарник из stage 1
COPY --from=builder /app/server .
# Копируем папку с миграциями — они нужны при запуске сервера
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./server"]
