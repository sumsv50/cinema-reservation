# Cinema Reservation System

A high-concurrency cinema seat reservation system with social distancing, built with Go, Gin, PostgreSQL, and Redis.

## Table of Contents
- [Introduction](#introduction)
- [Solution](#solution)
- [Setup Instructions](#setup-instructions)
- [API Documentation](#api-documentation)
- [Testing](#testing)
- [Project Structure](#project-structure)
- [Future Improvements](#future-improvements)

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

## Solution

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

📊 View Sequence Diagram: [Sequence Diagram](./sequenceDiagram.mmd)

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

### Prerequisites
Before running tests, you must set up the test environment:

1. Create a new cinema with minimum distance = 1
2. Update test configuration in the code with the following parameters:
    - `totalRequests`: Number of concurrent requests to simulate
    - `slug`: Cinema identifier
    - `rows`: Number of seat rows
    - `columns`: Number of seat columns
    - Additional cinema-specific parameters as needed

*Note: You may need to remove the rate limit middleware from the Gin router before running the tests. (command row 92-93 in main.go)*

### Running Tests
- **Test multiple reservations for the same seat** <br/>
  Ensures only the first reservation succeeds and the rest are marked as conflicts:
  
  ```sh
  go test -v ./test/reservation_same_seat_test.go
  ```

- **Test reservations for all seats (including duplicates and valid cases)** <br/>
  Ensures all seats are booked correctly, including duplicates and valid reservations:
  
  ```sh
  go test -v ./test/reservation_all_seats_test.go
  ```

### My Test Results
The system handled 10,000 concurrent requests successfully when tested on my local machine (MacBook Pro 2021, M1 chip, 16GB RAM).

![Test result 1](https://res.cloudinary.com/hpit/image/upload/v1750003010/FollMe/cinema-reservation/Screenshot_2025-06-15_at_18.26.08_wdkusx.png)

![Test Result 2](https://res.cloudinary.com/hpit/image/upload/v1750003009/FollMe/cinema-reservation/Screenshot_2025-06-15_at_18.25.43_yt0xta.png)

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

## Future Improvements
If a reservation cancellation has already been processed in the database but fails to cancel in Redis, a mechanism is needed to retry and synchronize data from the database back to Redis:

- Implement background jobs to automatically retry the cancellation in Redis.
- If all retry attempts fail, notify the admin so they can either retry manually or trigger a manual sync from the database to Redis.
