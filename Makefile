.PHONY: dev
dev: 
	go run main.go server --config conf.toml

.PHONY: run-scheduler
run-scheduler:
	go run main.go scheduler --config conf.toml

.PHONY: run-notifier
run-notifier:
	go run main.go notifier --config conf.toml

.PHONY: services
services:
	docker-compose up -d