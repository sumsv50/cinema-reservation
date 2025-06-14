# Cinema Reservation System

A high-concurrency cinema seat reservation system with social distancing, built with Go, Gin, PostgreSQL, and Redis.

## Table of Contents
- [Features](#features)
- [Setup Instructions](#setup-instructions)
- [API Documentation](#api-documentation)
- [Assumptions and Design Decisions](#assumptions-and-design-decisions)
- [Testing](#testing)
- [Environment Variables](#environment-variables)
- [Project Structure](#project-structure)
- [License](#license)

---

## Introduction
This project focuses on solving a computational problem where the system must handle a high volume of concurrent seat reservation and cancellation requests. To simplify the core logic, we deliberately omit the following aspects:
- Practical Application Details: The reservation system does not associate seats with specific movies or showtimes. Instead, seats are reserved based solely on their position within a cinema.
- Authentication: User authentication and authorization are not included in the current scope.

Main features:
- Create cinema layouts with configurable name, rows, columns, and minimum seat distance.
- Reserve and cancel seats with atomicity and social distancing enforcement.
- High concurrency support using Redis and Lua scripts.
- Rate limiting and robust error handling.
- Health check endpoint for service monitoring.

---

## 💡 Solution

Offloads the critical, concurrent-sensitive seat check to Redis for speed and atomicity. Redis stores seat reservations per cinema screen using a 2D layout as a hash.

Atomic Lua script in Redis:
- Validates the seat block is free
- Checks for social distancing (Manhattan distance)
- Reserves the seats if valid

Go backend:
- Validate inputs (seat position, cinema exists, group size)
- Delegate reservation to Redis
- Persist reservation in DB
- Respond to clients
- Logging, rate limiting

📊 View Sequence Diagram: [Authentication Sequence Diagram](./sequenceDiagram.mmd)

>Redis is the gatekeeper — it enforces safety. <br>
Go app is the orchestrator — it handles flow, persistence, and responses.

---

## Setup Instructions

### Prerequisites
- Go 1.23+
- Docker & Docker Compose

### 1. Clone the repository
```sh
git clone https://github.com/sumsv50/cinema-reservation.git
cd cinema-reservation
```

### 2. Start dependencies (PostgreSQL & Redis)
```sh
docker-compose up -d
```

### 3. Configure environment variables
Copy `.env.example` to `.env` and adjust as needed:
```
cp .env.example .env
```
- `DATABASE_URL`: PostgreSQL connection string
- `REDIS_URL`: Redis connection string

### 4. Run the server
```sh
go run ./cmd/server/main.go
```
The server will start on `localhost:8080` by default.

---

## API Documentation

### Health Check
- `GET /health`  
  Returns the health status of the service and its dependencies.

### Cinema Management
- Configure Cinema Layout (create a new cinema):
  - **Path:** `POST /api/v1/cinemas`
  - **Body:**  
    ```json
    {
      "name": "Grand Cinema Downtown",
      "rows": 10,
      "columns": 15,
      "min_distance": 2
    }
    ```
  - **Response:** Created cinema details.

- Query Available Seats:
  - **Path:** `GET /api/v1/cinemas/{slug}/seats?number_of_seats=3`
  - Returns available seat blocks for a group.

- Check Available Seats:
  - **Path:** `POST /api/v1/cinemas/{slug}/seats/check-availability`
  - **Body:**  
    ```json
    {
      "seats": [
        {"row": 1, "column": 2},
        {"row": 1, "column": 3}
      ]
    }
    ```
  - **Response:** List of available seats from the request.

### Reservation
- Reserve Seats:
  - **Path:** `POST /api/v1/reservations`
  - **Body:**  
    ```json
    {
      "cinema_slug": "grand-cinema-downtown",
      "note": "Friends night",
      "seats": [
        {"row": 1, "column": 2},
        {"row": 1, "column": 3}
      ]
    }
    ```
  - **Response:** Cancel details.

- Cancel Reservation:
  - **Path:** `DELETE /api/v1/reservations`
  - **Body:**  
    ```json
    {
      "cinema_slug": "grand-cinema-downtown",
      "seats": [
        {"row": 1, "column": 2}
      ]
    }
    ```
  - **Response:** Success message.

---


## Testing
- Run all tests:
  ```sh
  go test ./test/...
  ```
- Includes:
  - High-concurrency seat reservation tests.
  - All-seats and same-seat edge case tests.

---

## Project Structure
- `cmd/server/` — Main entrypoint
- `internal/handlers/` — HTTP handlers
- `internal/services/` — Business logic
- `internal/repositories/` — Data access
- `internal/models/` — Data models
- `internal/database/` — DB/Redis setup
- `internal/scripts/` — Lua scripts for Redis
- `internal/middleware/` — Gin middleware
- `internal/utils/` — Utilities and error handling
- `test/` — Integration and concurrency tests

