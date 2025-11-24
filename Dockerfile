FROM golang:1.25-alpine AS builder

WORKDIR /usr/local/src

# качаем зависимости зависимости
COPY ["./go.mod", "./go.sum", "./"]
RUN go mod download

#билдим проект в бинарник
COPY . .
RUN go build  -o ./bin/app ./cmd/pr-service/main.go


FROM alpine AS runner
WORKDIR /app

# Копируем бинарник, .env файл и миграции
COPY --from=builder /usr/local/src/bin/app /app/app
COPY  migrations /app/migrations
COPY .env /app/.env

CMD ["/app/app"]