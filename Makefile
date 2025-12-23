up:
	docker compose up -d

down:
	docker compose down

reset: down
	docker compose down -v
	docker compose up -d
	@sleep 5
	make migrate

migrate:
	docker compose run --rm migrate up

test:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run

fmt:
	go fmt ./...
	goimports -w .

logs:
	docker compose logs -f

clean:
	docker compose down -v
	rm -f coverage.out coverage.html

rabbit-ui:
	@echo "RabbitMQ Management UI: http://localhost:15672 (guest/guest)"

consume:
	go run cmd/consumer/main.go
