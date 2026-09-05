# cmd/agent

В данной директории содержится код Агента, который скомпилируется в бинарное приложение.

## Сборка проекта

Для компиляции компонентов с автоматическим вшиванием версии, даты, коммита через флаги линкера (-ldflags -X), выполните команду из корня репозитория:

```bash
go build -ldflags "-X main.buildVersion=$(git describe --tags --always) -X 'main.buildDate=$(date +'%Y-%m-%d %H:%M:%S')' -X 'main.buildCommit=$(git rev-parse --short HEAD)'" -o agent ./cmd/agent
```