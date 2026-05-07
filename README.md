# Dawg Patrol API

REST API built with Go for real-time robot control via WebSockets. Includes authentication, rate limiting, and a seed for local development.

---

## Tech Stack

- **Go** — main language
- **PostgreSQL** — database
- **JWT** — authentication
- **Docker & docker-compose** — local database setup
- **Air** — live reload for development
- **Bruno** — API collection for testing endpoints
- **direnv** — environment variable management
- **golang-migrate** — database migrations

---

## Prerequisites

- [Go](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/)
- [direnv](https://direnv.net/docs/installation.html)
- [Air](https://github.com/air-verse/air)
- [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)
- [Bruno](https://www.usebruno.com/) *(optional, for testing endpoints)*

---

## Getting Started

### 1. Clone the repo

```bash
git clone https://github.com/erickgreco/dawg-patrol.git
cd dawg-patrol
```

### 2. Set up environment variables

```bash
cp .envrc.example .envrc
```

Fill in your values in `.envrc`, then allow direnv:

```bash
direnv allow
```

### 3. Start the database

```bash
docker-compose up -d
```

### 4. Run migrations

```bash
make migrate-up
```

### 5. Seed the database *(optional but recommended)*

Populates the database with test users and robots. Logs credentials to `cmd/seed/logs/seed.log` so you can use them in the endpoints right away.

```bash
make seed
```

### 6. Start the server

```bash
air
```

The server will be available at `http://localhost:8080`.

---

## Makefile Commands

|            Command             | Description |
|--------------------------------|---------------------------------------|
| `make migrate-up`              | Run all pending migrations            |
| `make migrate-down <n>`        | Roll back `n` migrations              |
| `make migrate-create <name>`   | Create a new migration file           |
| `make migrate-force <version>` | Force migration to a specific version |
| `make seed`                    | Seed the database with test data      |

---

## API Endpoints

| Method |             Path                  | Auth |       Description       |
|--------|-----------------------------------|------|-------------------------|
| GET    | `/v1/health`                      |  ❌  | Health check            |
| POST   | `/v1/register`                    |  ❌  | Register a new user     |
| POST   | `/v1/login`                       |  ❌  | Login and receive a JWT |
| GET    | `/v1/home`                        |  ✅  | Home (protected)        |
| GET    | `/v1/profile`                     |  ✅  | Get current user info   |
| POST   | `/v1/profile/request-role-update` |  ✅  | Request a role update   |
| POST   | `/v1/robots/register-robot`       |  ✅  | Register a robot        |

> WebSocket endpoints for real-time robot control are in progress.

### Authentication

Protected endpoints require a `Bearer` token in the `Authorization` header:

```
Authorization: Bearer <your_jwt_token>
```

---

## Bruno Collection

The `bruno/dawg-patrol-api/` folder contains the API collection with pre-configured environment variables (`baseUrl`, `token`).

1. Open Bruno and load the collection from `bruno/dawg-patrol-api/`.
2. Select the `dawg-patrol-bru-env` environment.
3. Login via the `login` request — the script stores the token automatically.
4. Use any protected endpoint directly.

---

## Project Structure

```
.
├── cmd/
│   ├── api/            # Server entrypoint
│   ├── migrate/        # Migration files
│   │   └── migrations/
│   └── seed/           # Database seeder
├── internal/
│   ├── auth/           # JWT and middleware
│   ├── domain/         # Shared domain types
│   ├── home/           # Home feature
│   ├── robots/         # Robots feature
│   └── users/          # Users feature
├── pkg/
│   ├── db/             # Database connection
│   ├── env/            # Env loader
│   ├── json/           # JSON helpers
│   └── myerrors/       # Custom errors
├── bruno/              # Bruno API collection
├── air.toml            # Air live reload config
├── docker-compose.yml
├── Makefile
└── .envrc.example
```

---

## Environment Variables

|   Variable   |        Description           |                Example                   |
|--------------|------------------------------|------------------------------------------|
| `ADDR`       | Server address               |               `:8080`                    |
| `DB_ADDR`    | PostgreSQL connection string | `postgres://user:pass@localhost:5432/db` |
| `JWT_SECRET` | Base64-encoded secret key    |         `openssl rand -base64 32`        |
| `JWT_EXPIRY` | Token expiry duration        |                 `30m`                    |
