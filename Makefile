include .env
export

migrate-up:
	goose -dir db/migrations postgres "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" up

migrate-down:
	goose -dir db/migrations postgres "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" down

db-up: 
	docker run --rm --name my_postgres -e POSTGRES_HOST_AUTH_METHOD=trust -e POSTGRES_USER=$(DB_USER) -e POSTGRES_DB=$(DB_NAME) -p $(DB_PORT):5432 -d postgres:14.3

db-down:
	docker stop my_postgres

db-restart:
	make db-down || make db-up

run:
	go run ./cmd/main.go