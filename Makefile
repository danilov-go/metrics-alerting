.PHONY: generate-key build build-agent build-server

generate-key:
	@echo "Генерация RSA-ключей"
	@mkdir -p certs
	@openssl genrsa -out certs/private.pem 2048
	@openssl rsa -in certs/private.pem -pubout -out certs/public.pem
	@echo "Ключи успешно созданы в папке ./certs/"

build: build-agent build-server

build-agent:
	@echo "Сборка Агента"
	@go build -ldflags "-X 'main.buildVersion=$$(git describe --tags --always)' -X 'main.buildDate=$$(date +"%Y-%m-%d %H:%M:%S")' -X 'main.buildCommit=$$(git rev-parse --short HEAD)'" -o agent ./cmd/agent

build-server:
	@echo "Сборка Сервера"
	@go build -ldflags "-X 'main.buildVersion=$$(git describe --tags --always)' -X 'main.buildDate=$$(date +"%Y-%m-%d %H:%M:%S")' -X 'main.buildCommit=$$(git rev-parse --short HEAD)'" -o server ./cmd/server