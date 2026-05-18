# Dawg Patrol API

REST API built with Go for real-time robot control via WebSockets. Includes JWT authentication, role-based access control, rate limiting, and a seed for local development.

---

## Tech Stack

- **Go** — main language
- **PostgreSQL** — database
- **JWT** — authentication (HS256, configurable expiry)
- **Gorilla WebSocket** — real-time robot telemetry
- **Chi** — HTTP router with middleware support

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

### 5. Start the server

```bash
air
```

The server will be available at `http://localhost:8080`.

### 6. Seed the database *(optional but recommended)*

Populates the database with 25 test users and 30 robots. Logs credentials to `cmd/seed/logs/seed.log` so you can use them right away. After seeding, also starts a telemetry mock that connects as a robot and sends live frames for 30 seconds — the server must already be running for this step.

```bash
make seed
```

> The `reservationID` and `robotID` used by the telemetry mock are written to `seed.log`. Use them in Bruno to test the user-telemetry WebSocket endpoint.

---

## Makefile Commands

|            Command             | Description                           |
|--------------------------------|---------------------------------------|
| `make migrate-up`              | Run all pending migrations            |
| `make migrate-down <n>`        | Roll back `n` migrations              |
| `make migrate-create <name>`   | Create a new migration file           |
| `make migrate-force <version>` | Force migration to a specific version |
| `make seed`                    | Seed the database and mock telemetry  |

---

## Domain

### User Roles

| Role       | Description                                                                  |
|------------|------------------------------------------------------------------------------|
| `VIEWER`   | Default role on registration. Can see robot count and request a role upgrade |
| `OPERATOR` | Can view robot data, reserve robots, and open WebSocket sessions             |
| `ADMIN`    | Full access — can register robots, see all categories, manage operators      |

### Robot Categories

The category is derived from the last capital letter in the robot name (e.g. `NoisyA1` → `ASSISTANT`).

| Letter | Category    |
|--------|-------------|
| `A`    | `ASSISTANT` |
| `S`    | `SUMO`      |
| `R`    | `RACER`     |

### Robot Statuses

| Status     | Meaning                           |
|------------|-----------------------------------|
| `IDLE`     | Available for reservation         |
| `IN_USE`   | Currently reserved                |
| `CHARGING` | Charging, not available           |
| `OFFLINE`  | Not available                     |

---

## API Endpoints

### Authentication

Protected endpoints require a `Bearer` token in the `Authorization` header:

```
Authorization: Bearer <your_jwt_token>
```

Roles in the table below indicate the **minimum role** required.

---

### Public Endpoints

#### `GET /v1/health`

Health check. No authentication required.

**Response `200`**
```json
{ "status": "ok" }
```

---

#### `POST /v1/register`

Register a new user. Rate limited to **10 requests/min per IP**.

**Request body**
```json
{
  "username": "shadowbyte",
  "email": "shadow@example.com",
  "password": "Shadow#2026"
}
```

| Field      | Type   | Rules                          |
|------------|--------|--------------------------------|
| `username` | string | required, max 30 chars         |
| `email`    | string | required, max 250 chars        |
| `password` | string | required, min 12, max 30 chars |

**Response `201`**
```json
{
  "data": {
    "id": "uuid",
    "username": "shadowbyte",
    "role": "VIEWER",
    "created_at": "2025-05-18T12:00:00Z"
  }
}
```

**Errors**

| Status | Condition            |
|--------|----------------------|
| `400`  | Validation failed    |
| `409`  | Email already exists |

---

#### `POST /v1/login`

Authenticate and receive a JWT token. Rate limited to **5 requests/min per IP**.

**Request body**
```json
{
  "email": "shadow@example.com",
  "password": "Shadow#2026"
}
```

**Response `200` — VIEWER**
```json
{
  "data": {
    "token": "<jwt>",
    "id": "uuid",
    "role": "VIEWER"
  }
}
```

**Response `200` — ADMIN / OPERATOR**
```json
{
  "data": {
    "token": "<jwt>",
    "refresh_token": "<jwt>",
    "id": "uuid",
    "role": "OPERATOR"
  }
}
```

> ADMIN and OPERATOR receive a `refresh_token` (JWT, valid for **7 days**) in addition to the short-lived access token (30 min). Store it and use `POST /v1/refresh` to rotate it when the access token expires.

**Errors**

| Status | Condition              |
|--------|------------------------|
| `400`  | Validation failed      |
| `401`  | Invalid email/password |

---

#### `POST /v1/refresh`

Issues a new access token and a new refresh token by validating and rotating an existing refresh JWT. No `Authorization` header required. Rate limited to **10 requests/min per IP**.

Only reachable by ADMIN and OPERATOR — VIEWER users never receive a refresh token on login.

**Request body**
```json
{
  "refresh_token": "<refresh_jwt>"
}
```

**Response `200`**
```json
{
  "data": {
    "token": "<new_jwt>",
    "refresh_token": "<new_refresh_jwt>",
    "id": "uuid",
    "role": "OPERATOR"
  }
}
```

> Each call **rotates** the refresh token — the old one is immediately invalidated and a new one is returned. Store the new `refresh_token` after every successful refresh.

**Errors**

| Status | Condition                                     |
|--------|-----------------------------------------------|
| `400`  | Missing or malformed body                     |
| `401`  | Invalid, expired, or already-used refresh token |

---

### Protected Endpoints

All endpoints below require `Authorization: Bearer <token>`. Rate limited to **100 requests/min per user**.

---

#### `GET /v1/home`

Returns the current user's summary and available idle robots. The robot data shown depends on the user's role.

**Response `200`**
```json
{
  "data": {
    "user": {
      "id": "uuid",
      "username": "shadowbyte",
      "role": "OPERATOR",
      "request_status": "NONE"
    },
    "num_robots": 12,
    "robots": []
  }
}
```

**Role-based robot visibility**

| Role       | `num_robots` counts   | `robots` array contains                      |
|------------|-----------------------|----------------------------------------------|
| `VIEWER`   | Racers + Sumos (idle) | `null` — no robot detail shown               |
| `OPERATOR` | Racers + Sumos (idle) | Racer and Sumo robot detail                  |
| `ADMIN`    | All types (idle)      | All robot detail including Assistants        |

**Errors**

| Status | Condition      |
|--------|----------------|
| `401`  | Invalid token  |
| `500`  | Internal error |

---

#### `GET /v1/profile`

Returns full profile data for the authenticated user along with available actions.

**Response `200`**
```json
{
  "data": {
    "profile": {
      "id": "uuid",
      "username": "shadowbyte",
      "email": "shadow@example.com",
      "role": "VIEWER",
      "is_active": true,
      "request_status": "NONE",
      "requested_at": null,
      "created_at": "2025-05-18T12:00:00Z",
      "updated_at": "2025-05-18T12:00:00Z"
    },
    "user_actions": {
      "action_update_password": true,
      "action_update_username": true,
      "action_request_role_update": {
        "action": true,
        "request_status": "NONE",
        "requested_at": "0001-01-01T00:00:00Z",
        "request_response": ""
      }
    }
  }
}
```

**Errors**

| Status | Condition     |
|--------|---------------|
| `401`  | Invalid token |

---

#### `POST /v1/profile/request-role-update`

Submits a role upgrade request from `VIEWER` to `OPERATOR`. Only one active request is allowed at a time.

**Response `201`**
```json
{
  "data": {
    "request_status": "PENDING",
    "requested_at": "2025-05-18T12:00:00Z"
  }
}
```

**Errors**

| Status | Condition                                    |
|--------|----------------------------------------------|
| `401`  | Invalid token                                |
| `404`  | User not found                               |
| `500`  | Request already pending or internal error    |

---

#### `POST /v1/robots/register-robot` — `ADMIN only`

Registers a new robot. If the serial number already exists, the record is updated (upsert).

**Request body**
```json
{
  "serial_number": "02:1A:C3:4D:5E:01",
  "name": "NoisyA1",
  "battery": 85
}
```

| Field           | Type    | Rules                                               |
|-----------------|---------|-----------------------------------------------------|
| `serial_number` | string  | Format `XX:XX:XX:XX:XX:XX` (hyphens are normalized) |
| `name`          | string  | Pattern `^[A-Z][a-z]+[A-Z][0-9]+$` e.g. `NoisyA1`  |
| `battery`       | integer | 0–100                                               |

The robot **category is inferred from the last capital letter** in the name:
- `A` → `ASSISTANT`, `S` → `SUMO`, `R` → `RACER`

**Response `202`**
```json
{
  "data": {
    "id": "uuid",
    "serial_number": "02:1A:C3:4D:5E:01",
    "name": "NoisyA1",
    "type": "ASSISTANT",
    "status": "IDLE",
    "battery": 85,
    "last_seen_at": "2025-05-18T12:00:00Z"
  }
}
```

**Errors**

| Status | Condition                               |
|--------|-----------------------------------------|
| `401`  | Invalid token                           |
| `403`  | Insufficient role (not ADMIN)           |
| `422`  | Invalid serial number, name, or battery |

---

#### `GET /v1/robots/idle-robots` — `ADMIN / OPERATOR`

Returns all robots currently in `IDLE` status, grouped by category.

**Response `200`**
```json
{
  "data": {
    "AssistantRobots": [],
    "SumoRobots": [],
    "RacerRobots": [
      {
        "id": "uuid",
        "serial_number": "02:1A:C3:4D:5E:06",
        "name": "BlazeR1",
        "type": "RACER",
        "status": "IDLE",
        "battery": 78,
        "last_seen_at": "2025-05-18T12:00:00Z"
      }
    ]
  }
}
```

**Errors**

| Status | Condition         |
|--------|-------------------|
| `401`  | Invalid token     |
| `403`  | Insufficient role |

---

#### `PATCH /v1/robots/{robotID}/reserve-robot` — `ADMIN / OPERATOR`

Reserves a robot for the authenticated user. The robot must be `IDLE` with battery ≥ 20%. Only one active reservation per robot and per user is allowed at a time. The reservation expires in **30 minutes** unless a WebSocket session keeps it alive.

**URL params**

| Param     | Description       |
|-----------|-------------------|
| `robotID` | UUID of the robot |

**Response `200`**
```json
{
  "data": {
    "reservation_id": "uuid",
    "user_id": "uuid",
    "robot_id": "uuid",
    "expires_at": "2025-05-18T12:30:00Z",
    "active": true,
    "created_at": "2025-05-18T12:00:00Z",
    "status": "IN_USE",
    "last_seen_at": "2025-05-18T12:00:00Z"
  }
}
```

**Errors**

| Status | Condition                             |
|--------|---------------------------------------|
| `401`  | Invalid token                         |
| `403`  | Insufficient role                     |
| `404`  | Robot not found                       |
| `409`  | Robot unavailable or battery too low  |

---

### WebSocket Endpoints

Both WS endpoints require `Authorization: Bearer <token>` in the **connection headers** and sit behind `RobotContextMiddleware` + `ReservationCtxMiddleware`, which validate the robot and the reservation before upgrading the connection.

**Reservation lifecycle**

```
reserve-robot ──► active=true, ws_started_at=null, expires_at=NOW()+30m
     │
     ├── no WS in 5 min  ──► cleanup worker: active=false, robot → IDLE
     │
     └── robot connects (ws/robot-telemetry)
              │
              ├── ws_started_at = NOW()
              ├── KeepReservationAlive extends expires_at every 5 min
              └── robot disconnects ──► active=false immediately, robot → IDLE
```

---

#### `GET /v1/robots/{robotID}/{reservationID}/ws/robot-telemetry`

WebSocket endpoint for the **robot** side of the session. Registers the robot in the hub and starts two goroutines: `KeepReservationAlive` (extends the reservation every 5 minutes) and `MockTelemetry` (generates fake telemetry frames every 2 seconds, useful during development).

**URL params**

| Param           | Description                    |
|-----------------|--------------------------------|
| `robotID`       | UUID of the reserved robot     |
| `reservationID` | UUID of the active reservation |

**Telemetry frame format** *(sent by robot, forwarded to paired user)*
```json
{
  "speed": 1.42,
  "direction": 210.5,
  "battery": 73,
  "timestamp": "2025-05-18T12:00:02Z"
}
```

**Behavior on disconnect**

When the robot disconnects, the hub immediately deactivates the reservation and sets the robot back to `IDLE` — no need to wait for the cleanup worker.

**Errors (HTTP, before upgrade)**

| Status | Condition                            |
|--------|--------------------------------------|
| `400`  | Invalid reservation or robot context |
| `401`  | Invalid token                        |
| `403`  | Insufficient role                    |

---

#### `GET /v1/robots/{robotID}/{reservationID}/ws/user-telemetry`

WebSocket endpoint for the **user** side of the session. Subscribes the user to an existing hub session identified by `reservationID`. Once subscribed, every telemetry frame sent by the robot is forwarded to the user in real time.

**URL params**

| Param           | Description                    |
|-----------------|--------------------------------|
| `robotID`       | UUID of the reserved robot     |
| `reservationID` | UUID of the active reservation |

> The robot must be connected via `ws/robot-telemetry` with the same `reservationID` for frames to be routed. If the robot connects after the user, frames start flowing automatically once it does.

**Behavior on disconnect**

When the user disconnects, the hub closes the robot's send channel. The robot's write pump detects this, sends a close frame, and terminates the robot connection — triggering the same immediate DB cleanup as a direct robot disconnect.

**Errors (HTTP, before upgrade)**

| Status | Condition                            |
|--------|--------------------------------------|
| `400`  | Invalid reservation or robot context |
| `401`  | Invalid token                        |
| `403`  | Insufficient role                    |

---

### Full WS Session Flow

```
1. PATCH /v1/robots/{robotID}/reserve-robot
        ↓
   reservation created: active=true, ws_started_at=null

2. GET  /v1/robots/{robotID}/{reservationID}/ws/robot-telemetry   (robot connects)
        ↓
   ws_started_at=NOW(), hub session created, KeepReservationAlive starts

3. GET  /v1/robots/{robotID}/{reservationID}/ws/user-telemetry    (user connects)
        ↓
   user subscribed to session, telemetry frames forwarded in real time

4. Either side disconnects
        ↓
   hub unregisters session, reservation deactivated, robot → IDLE
```

---

## Bruno Collection

The `bruno/dawg-patrol-api/` folder contains the full API collection with pre-configured environment variables (`base_url`, `token`, `refresh_token`, `robotID`, `reservationID`).

1. Open Bruno and load the collection from `bruno/dawg-patrol-api/`.
2. Select the `dawg-patrol-bru-env` environment.
3. Run `login-admin` or `login-operator` — the post-response script stores both `token` and `refresh_token` automatically.
4. Use any protected endpoint directly.
5. When the access token expires, run `refresh` — it rotates both tokens in the environment automatically.
6. For WebSocket testing, use the `ws-robot-telemetry` and `ws-user-telemetry` requests after reserving a robot.

---

## Project Structure

```
.
├── cmd/
│   ├── api/                # Server entrypoint and router
│   ├── migrate/            # Migration runner
│   │   └── migrations/     # SQL migration files
│   └── seed/               # Database seeder + telemetry mock
│       └── logs/           # seed.log with generated credentials
├── internal/
│   ├── apimiddleware/      # Auth, role, robot and reservation middleware
│   ├── auth/               # JWT token service
│   ├── domain/             # Shared domain constants (roles, categories, statuses)
│   ├── home/               # Home and reserve-robot feature
│   ├── robots/             # Robot registration, reservation, cleanup worker
│   ├── users/              # User registration, login, profile, role requests
│   └── websockets/         # Hub, robot/user clients, WS service and handler
├── pkg/
│   ├── db/                 # Database connection and pool config
│   ├── env/                # Environment variable loader
│   ├── json/               # JSON read/write helpers and validator
│   └── myerrors/           # Sentinel errors and HTTP error helpers
├── bruno/                  # Bruno API collection
├── air.toml                # Air live reload config
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
