.PHONY: dev
dev: 
	go run main.go server --config conf.toml

.PHONY: services
services:
	docker-compose up -d