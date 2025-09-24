# Library

![интерфейсы](docs/scheme/Interfaces.png)
![реализация](docs/scheme/Implementations.png)

## Документация
[README.md](./docs/README.md) в ./docs

## Разработка

![разработка](docs/scheme/developing.png)

## Унификация технологий
* Структура проекта [go-clean-template](https://github.com/evrone/go-clean-template)
* Для логирования [zap](https://github.com/uber-go/zap)
* Для валидации [protoc-gen-validate](https://github.com/bufbuild/protoc-gen-validate)
* Для поддержики REST-to-gRPC API [gRPC gateway](https://grpc-ecosystem.github.io/grpc-gateway/)

## Тестирование в CI
* Код тестов можно посмотреть в файле [integration_test.go](./integration-test/integration_test.go)
* Для прохождения тестов необходимо записать переменные окружения для инициализации конфига, см: [README](./cmd/library/README.md)

## Авторские заметки
* Для понимания Makefile и easyp можно заглянуть в Go-CT-Learning репозиторий (private).