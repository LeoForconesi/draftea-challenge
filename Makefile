up:
	docker compose up -d --build
	echo "Migrating database..."
	@sleep 5
	make migrate

down:
	docker compose down

reset: down
	docker compose down -v
	docker compose up -d
	@sleep 5
	make migrate

migrate:
	make migrate-up

migrate-up:
	docker compose run --rm migrate up

migrate-down:
	docker compose run --rm migrate down ${N:-1}

migrate-version:
	docker compose run --rm migrate version

migrate-force:
	@if [ -z "${V}" ]; then \
		echo "V is required. Example: make migrate-force V=2"; \
		exit 1; \
	fi
	docker compose run --rm migrate force ${V}

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
