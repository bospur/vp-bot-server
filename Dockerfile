# ── Stage 1: сборка ──────────────────────────────────────────────────────────
# Берём официальный образ Go для компиляции
FROM golang:1.24-alpine AS builder

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

EXPOSE 8080

CMD ["./server"]
