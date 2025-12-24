# draftea-challenge

![Coverage](coverage-badge.svg)

Production-like wallet and payments API built with Go, Gin, PostgreSQL, and RabbitMQ.
The idea is to provide a solution for a coding challenge that demonstrates best practices in designing and implementing a robust, maintainable, and scalable service.

## Considerations
- You'll need Docker and Docker Compose installed and running in your local environment.
- By default this service will launch in docker. You can change this by setting the `APP_ENV` environment variable to `local` to run outside of docker.
This is useful for debugging or development.
- You can find configuration files under the `config/` directory. As said in the previous point, by default it will use `config.docker.yaml` when running in docker and `config.local.yaml` when running locally.
- Once you run the application with make up you can find the following services available:
  - Swagger Ui is available at [http://localhost:8080/docs/index.html](http://localhost:8080/docs/index.html).
  - You can access the mock payment gateway at [http://localhost:8081](http://localhost:8081).
  - RabbitMQ management UI is available at [http://localhost:15672](http://localhost:15672)
  - The API is exposed at [http://localhost:8080](http://localhost:8080).
- A full Postman collection is provided in [`/postman/collections`](postman/collections/DrafteaChallenge.postman_collection.json) you can import this Json into your postman or REST Client of choice to explore the endpoints. Take into consideration that {{$guid}} from postman is used for UUID generation in requests, and for idempotency keys, you can replace this to any UUID generator of your choice or manually create one. In the same folter you can find environment variables that will help you set up the requests.
- This app implementation uses idempotency key in order to avoid duplicate processing of requests. The key must be provided in the `Idempotency-Key` header for endpoints that support it (e.g., payment processing). The service will return the same response for requests with the same idempotency key. So if you want to create a new payment, make sure to use a different key each time. If you use the postman collection provided, this is handled automatically, but it's nice to take this into consideration if you want to test duplication handling.
- I recommend to take a look to Makefile scripts to see if there's any other command that could be useful for you, since some commands are not documented here for brevity.

## Docs
- [Challenge description](docs/challenge.md)
- [Architecture overview](docs/architecture/system-overview.md)
- [Service design](docs/architecture/service-design.md)
- [Database schema](docs/database/schema.md)
- [Tech stack](docs/decisions/tech-stack.md)
- [Error handling](docs/operations/error-handling.md)
- [Runbook](docs/operations/runbook.md)
- [About Mock Gateway](docs/architecture/about-mock-gateway.md)
- [Improvements](docs/improvements/improvements.md)
  - [cloud-deployment.md](docs/improvements/cloud-deployment.md)
  - [integration-tests.md](docs/improvements/integration-tests.md)
  - [gateway-resilience.md](docs/improvements/gateway-resilience.md)
  - [performance-optimization.md](docs/improvements/performance-optimization.md)
  - [not-included-yet.md](docs/improvements/not-included-yet.md) 👈🏻 (recommended)
- [OpenAPI spec](docs/openapi.yaml)
- [AI usage](docs/ai/usage.md) 🦾🤖(interesting read about how I used AI to help me build this project)

## Quickstart
- `make up`
- `make migrate`
- `make logs`

## Migrations
Migrations use `golang-migrate` with versioned SQL files under `migrations/`:
- Apply: `make migrate` (uses `docker compose run --rm migrate up`)
- Rollback: `docker compose run --rm migrate down 1`

## Docker Compose Flow
- `make up` starts Postgres, RabbitMQ, mock gateway, API, relay, and consumers, and then will run migrations.
- `make migrate` runs the migration container against the Postgres service.

## Configuration
Configuration precedence is:
1) YAML config at the path specified by `CONFIG_FILE` (optional).
2) Environment variables override YAML values.
3) Built-in defaults.

Included files:
- `config/config.local.yaml` for running locally.
- `config/config.docker.yaml` for docker-compose.
- If `CONFIG_FILE` is not set, the app selects a default based on `APP_ENV`:
  - `local` -> `config/config.local.yaml`
  - `docker` -> `config/config.docker.yaml`
  - `stage` -> `config/config.stage.yaml`
  - `prod` -> `config/config.prod.yaml`

## Test-Only Endpoints
I've created some test-only endpoints to facilitate local testing and visibility:
- `GET /healthz` : health check. Returns 200 OK if the service is healthy.
- `POST /wallets/{user_id}/top-up` : adds balance to a wallet for testing purposes.
- `GET /wallets`: Lists wallets and balances (paginated), useful for testing if you don't want to access the DB directly.

## API Key (Test Only)
If `app.api_key` is set, include the header `X-API-Key: <your-api-key>` on requests. app.api_key is only set in local and docker configs for testing convenience.
In production, use proper authN/authZ (e.g., OAuth2/JWT + user service).

---
## Included
- Idempotent payment processing with `Idempotency-Key` header.
- Outbox pattern with RabbitMQ for async event publishing.
- Mock payment gateway with simulated delays and failures.
- Clean/Hexagonal architecture with layered design.
- Docker Compose setup for local development.
- Comprehensive documentation and runbook.
- Makefile with useful commands.
- OpenAPI spec and Swagger UI.
- CORS support.
- Patterns like circuit breaker and exponential backoff for gateway resilience.
- Builders, Factories, and Dependency Injection for maintainability.
- DDD principles for domain modeling.
- Hexagonal/Clean architecture for separation of concerns.
- Database migrations with versioned SQL files.
- PostregSQL and RabbitMQ as core infrastructure.
- Postman collection for easy API exploration.
- Usage of IA for coding assistance and documentation.
---

## Some tools used in this project:
- [Mermaid](https://mermaid-js.github.io/mermaid/#/) - for diagrams in documentation.
- [Swagger UI](https://swagger.io/tools/swagger-ui/) - for interactive API docs.
- [Golang-Migrate](https://github.com/golang-migrate/migrate) - for database migrations.
- [Excalidraw in my local](https://github.com/excalidraw/excalidraw) - for architecture diagrams.
- [Postman](https://www.postman.com/) - for API testing and collections.
- [ChatGPT-4](https://chat.openai.com/) - for coding assistance and documentation help.
- [Dbdiagram](https://dbdiagram.io/d) - for database schema design.
- [GoLand](https://www.jetbrains.com/go/) - as the main IDE for development.
- [Docker](https://www.docker.com/) - for containerization and local environment setup.
- [DBeaver](https://dbeaver.io/) - for database management and querying locally.
- [CodeQL](https://codeql.github.com/) - for static code analysis and security checks.
- [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) - for vulnerability scanning of Go dependencies.
- [GitHub Actions](https://github.com/features/actions) - for CI/CD workflows.
---

## Contact
Any questions or feedback are welcome! Feel free to open an issue or reach out directly. Enjoy exploring the code! (⌐ ͡■ ͜ʖ ͡■)

**Leonardo A. Forconesi**<br>**(Backend Engineer)**

   <a href="https://www.linkedin.com/in/leonardo-forconesi/">
    <img src="docs/resources/LinkedIn_icon.svg.png" alt="LinkedIn" width="24" height="24">
   </a>

