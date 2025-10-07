# golang - родительский образ, последней версии
FROM golang:latest AS builder

# установка рабочей директории для всех последующих команд
WORKDIR /application
# перенос всех файлов из текущей директории в рабочую директорию образа
COPY . .
# создание нового слоя (сборка под linux x86)
RUN make generate && GOOS=linux GOARCH=amd64 make build

FROM gcr.io/distroless/base-debian12 AS runtime
WORKDIR /app
COPY --from=builder /application/bin/library /app/library
# запуск приложения без аргументов
CMD ["/app/library"]
