# Queue Service

Simple HTTP queue service with Postgres storage.

## Requirements
- Docker + Docker Compose

## Quick Start

1. Start Postgres, run migrations, and start the app:
```sh
docker compose up --build
```

2. Service will be available on `http://localhost:8080`.

## Migrations

Migrations are applied automatically by the `migrate` service on `docker compose up`.

To rollback one migration step:
```sh
docker compose run --rm migrate_down
```

## API
OpenAPI spec is in `api/openapi.yml`.

Endpoints:
- `POST /v1/queues/{queueName}/push`
- `POST /v1/queues/{queueName}/pop`
- `POST /v1/tasks/{taskId}/ack`
- `POST /v1/tasks/{taskId}/nack`
