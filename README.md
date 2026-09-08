# OttoDot Trial Class Booking System

A robust, production-grade trial class booking and payment management system built in Go, adhering strictly to **Clean Architecture** (Ports & Adapters) and the **Dependency Inversion Principle (DIP)**.

The system features a dynamic **HTMX + Tailwind CSS** frontend, a REST API, deterministic payment simulation, concurrency-safe seat reservation, tiered rate limiting, and 100% statement test coverage on core business logic.

---

## Table of Contents
1. [How to Run the Solution](#how-to-run-the-solution)
2. [What Was Built](#what-was-built)
3. [Time Spent](#time-spent)
4. [Assumptions Made](#assumptions-made)
5. [Key Architecture & Backend Decisions](#key-architecture--backend-decisions)
6. [What Was Deliberately Cut](#what-was-deliberately-cut)
7. [What to Monitor After Release](#what-to-monitor-after-release)
8. [What to Do Next with More Time](#what-to-do-next-with-more-time)

---

## How to Run the Solution

### Prerequisites
- **Go**: 1.22 or higher
- **PostgreSQL 16+** or **Podman** / **Docker**
- **golang-migrate**: (`brew install golang-migrate` or via binary)
- **Google Wire**: (`go install github.com/google/wire/cmd/wire@latest`)

---

### Step 1: Start PostgreSQL (Docker / Podman)

Start the PostgreSQL container, which automatically initializes both the system database (`db_ottodot_trial`) and the test database (`db_ottodot_trial_test`):

```bash
# Using Makefile (auto-detects podman or docker)
make docker-up
# or
make podman-up
```

*Alternatively, run with docker-compose directly:*
```bash
docker compose up -d
```

---

### Step 2: Run Database Migrations

Apply database migrations to set up tables, unique indexes, foreign keys, and seed data:

```bash
# Migrate development/system database
make migrate-up

# Migrate dedicated test database
make migrate-test-up
```

#### Reset / Fresh Database
To drop all existing tables and re-run all migrations and initial seed data from scratch:

```bash
# Fresh development/system database
make db-fresh
# or directly via script: ./scripts/db-fresh.sh system

# Fresh test database
make db-fresh-test

# Fresh BOTH databases
make db-fresh-all
```

---

### Step 3: Run the Application

Start the HTTP web server:

```bash
make run
```
*(Or compile binary with `make build` and run `./bin/ottodot-trial`)*

The server will start on port **`8080`**:
- **Parent Booking UI**: [`http://localhost:8080`](http://localhost:8080)
- **Admin Roster UI**: [`http://localhost:8080/admin/roster`](http://localhost:8080/admin/roster)
- **Interactive API Documentation (Scalar UI)**: [`http://localhost:8080/docs`](http://localhost:8080/docs) (or [`http://localhost:8080/api/v1/docs/`](http://localhost:8080/api/v1/docs/))
- **OpenAPI Specification**: [`http://localhost:8080/api/v1/docs/openapi.json`](http://localhost:8080/api/v1/docs/openapi.json)
- **Health Check API**: [`http://localhost:8080/api/health`](http://localhost:8080/api/health)

To regenerate OpenAPI documentation after modifying route annotations:
```bash
make swag
```

---

### Step 4: Run Tests

The test suite runs against the dedicated test database (`db_ottodot_trial_test`), completely isolated from development data:

```bash
# Run all tests with coverage output and usecase summary
make test

# View interactive HTML visual coverage report in your browser
go tool cover -html=coverage.out
```

---

## What Was Built

1. **Clean Architecture Core (`internal/`)**:
   - **`domain/`**: Entities (`Parent`, `Student`, `TrialClass`, `Booking`, `PaymentAttempt`), statuses (`pending`, `confirmed`, `cancelled`, `payment_failed`), and relationships.
   - **`usecase/`**: Orchestrates business rules, owns consumer interfaces (`ParentRepo`, `StudentRepo`, `TrialClassRepo`, `BookingRepo`, `PaymentAttemptRepo`, `PaymentGateway`, `TxManager`), and manages booking state transitions.
   - **`repository/`**: GORM infrastructure providers fulfilling repository contracts.
   - **`infrastructure/payment/`**: Infrastructure adapter implementing `usecase.PaymentGateway` with deterministic failure simulation and latency control.
   - **`delivery/http/`**: Gin HTTP handlers, structured responses, template rendering, and middleware.

2. **Interactive HTMX Frontend**:
   - **Parent & Child Switcher**: Select a parent to load their enrolled children asynchronously via HTMX (`GET /api/v1/students`).
   - **Real-time Class Catalog**: Instant spots left computation and "Book Now" / "Enrolled" badges (`GET /api/v1/classes`).
   - **Checkout & Payment Drawer**: Shows booking details and lets the evaluator trigger a normal payment (`simulate_failure=false`) or a simulated failure (`simulate_failure=true`).
   - **Live Polling / Inline Feedback**: Displays real-time confirmation banners or failure retry options without full page reloads.
   - **Admin Class Roster**: View confirmed student rosters and enrollment counts per trial class (`/admin/roster`).

3. **Concurrency-Safe Last-Seat Reservation**:
   - Transactional pessimistic row-level locking (`SELECT ... FOR UPDATE`) serializes simultaneous booking attempts for the same class, preventing overbooking.
   - Auto-cancellation when the last seat is taken by a competing request.

4. **Tiered Rate Limiting Middleware**:
   - **Global Limit**: 50 req/sec (burst 100).
   - **Sensitive Route Limit**: 5 req/sec (burst 10) on checkout routes (`POST /api/v1/bookings`, `POST /api/v1/bookings/:id/pay`).
   - **Dual Content Negotiation**: Returns `429 Too Many Requests` with user-friendly HTMX alert banners for web clients and structured JSON for REST clients.
   - **Automatic Cleanup**: Background routine evicts stale visitor IPs every 10 minutes to prevent memory leaks.

5. **Security & Production Hardening**:
   - HTTP server connection timeouts (`ReadHeaderTimeout: 5s`, `ReadTimeout: 15s`, `WriteTimeout: 15s`, `IdleTimeout: 60s`) preventing Slowloris attacks.
   - Security headers middleware (`X-Content-Type-Options`, `X-Frame-Options`, `X-XSS-Protection`, `Referrer-Policy`).
   - Strict UUID input validation to reject malformed parameters before database querying.
   - Graceful shutdown closing the HTTP listener and PostgreSQL connection pool.

---

## Time Spent

| Phase / Component | Description | Time Spent |
| :--- | :--- | :---: |
| **Architecture Scaffolding & DB Migrations** | Initialized Clean Architecture boilerplate, designed relational schema, partial index constraints, and seeds. | ~20 min (0.33 h) |
| **Core Usecase & Concurrency Strategy** | Implemented usecase orchestration, concurrency locking (`FOR UPDATE`), and business error branches. | ~35 min (0.58 h) |
| **Architecture Decoupling & Optimization** | Decoupled `PaymentGateway` into infrastructure layer, fixed N+1 listing query, and isolated test DB. | ~25 min (0.42 h) |
| **Exhaustive Unit & Concurrency Testing** | Leveraged AI to generate mocks and edge cases to achieve 100% statement coverage on usecase logic. | ~30 min (0.50 h) |
| **Delivery Handlers & HTMX Frontend** | Wired Gin handlers, implemented token-bucket rate limiting middleware, and built HTMX UI views. | ~25 min (0.42 h) |
| **E2E Verification & API Documentation** | Verified race conditions, payment failure simulations, DB resets, and generated Scalar/Swagger docs. | ~15 min (0.25 h) |
| **Total** | | **~2.5 h (150 min)** |
---

## Assumptions Made

1. **Student Active Booking Rule**:
   - A student can only have **one active booking** (`pending` or `confirmed`) per trial class at any given time.
   - Enforced both at the usecase validation layer and via a PostgreSQL partial unique index:
     ```sql
     CREATE UNIQUE INDEX idx_unique_active_booking
         ON bookings(student_id, trial_class_id)
         WHERE status IN ('pending', 'confirmed');
     ```
2. **Deterministic Mock Payment**:
   - Real payment gateway API keys (e.g., live Midtrans/Stripe credentials) were not required for this trial.
   - Instead, the UI exposes a "Simulate Failed Payment" toggle so evaluators can deterministically verify both success and failure flows without real credit cards.
3. **Fixed Trial Pricing**:
   - Each trial class is priced at a flat rate of **$500.00** (`50000` cents).
4. **Sessionless Parent Selection**:
   - To provide the best reviewer experience, user authentication (login/session cookies) was omitted in favor of a parent selector dropdown. Any parent can be selected to simulate their view and their children.
5. **Class Capacity**:
   - Trial classes have small capacities (e.g. 4 students) to make race condition testing immediate and observable.

---

## Key Architecture & Backend Decisions

### 1. Consumer-Owned Interfaces (DIP)
Following idiomatic Go Clean Architecture:
- The **high-level module** (`internal/usecase`) defines the contracts it requires (`ParentRepo`, `StudentRepo`, `TrialClassRepo`, `BookingRepo`, `PaymentAttemptRepo`, `PaymentGateway`, `TxManager`).
- The **low-level modules** (`internal/repository` and `internal/infrastructure/payment`) implement these contracts implicitly.
- Handlers (`internal/delivery/http`) depend on `BookingUseCaseInterface`, enabling total decoupling and isolated unit testing.

### 2. Concurrency Strategy: Option 1 vs. 2-Phase Reservation
- **Strategy Used (Option 1 - Pessimistic Row Lock)**:
  - Inside a transaction, `bookingRepo.FindByIDForUpdate` and `classRepo.FindByIDForUpdate` lock the rows while the mock payment is executed.
  - Guarantees strict serializability and mathematical zero-overbooking for this demonstration.
- **Production Recommendation (2-Phase Reservation)**:
  - In a high-throughput system with real third-party gateways (e.g. Stripe, Midtrans), holding database locks during external network calls can exhaust the connection pool.
  - The production pattern would use a 2-Phase Reservation:
    - *Phase A*: Acquire brief lock, mark seat as `reserved_pending_payment` with a 10-minute TTL.
    - *Phase B*: Release lock, invoke payment gateway outside the database transaction.
    - *Phase C*: Confirm booking upon payment success or webhook callback; release seat on failure/timeout.

### 3. Elimination of N+1 Query Bottleneck
- Rather than running separate `CountConfirmedForClass` queries for every trial class in a loop, the class catalog uses `CountConfirmedGroupedByClass`:
  ```sql
  SELECT trial_class_id, count(*) FROM bookings WHERE status = 'confirmed' GROUP BY trial_class_id;
  ```
- This reduces database round-trips from $O(N)$ to $O(1)$.

### 4. Pluggable Payment Infrastructure
- Decoupled into `internal/infrastructure/payment/mock_payment.go`.
- If integrating a real provider (e.g., Midtrans Snap or Stripe Elements) in the future, only a new file in `internal/infrastructure/payment/` needs to be created. The usecase and domain remain untouched.

### 5. Dedicated Test Database
- `make test` targets `db_ottodot_trial_test`, keeping test runs, fixtures, and truncations completely isolated from the runtime system database `db_ottodot_trial`.

---

## What Was Deliberately Cut

1. **Full Authentication / User Management**:
   - Replaced username/password and JWT authentication with a parent switcher dropdown. This prioritizes evaluators testing the core trial booking domain without credential friction.
2. **Distributed Redis Cache & Locking**:
   - Used PostgreSQL row-level locks and Go in-memory token bucket rate limiting rather than adding Redis. This keeps the application self-contained, lightweight, and easy to run locally.
3. **Asynchronous Message Queue (Kafka/RabbitMQ)**:
   - Synchronous mock processing was favored over an asynchronous worker queue for simplicity and immediacy of evaluation.
4. **Live Payment Gateway SDKs**:
   - Proprietary SDKs were skipped in favor of a clean `PaymentGateway` port that allows plug-and-play integration later.

---

## What to Monitor After Release

1. **Database Lock Wait Time & Deadlocks**:
   - `pg_stat_activity` and `pg_locks` to detect lock contention on popular trial classes near capacity.
2. **Payment Funnel Conversion & Failures**:
   - Ratio of bookings created in `pending` vs. finalized to `confirmed` vs. `payment_failed`.
   - Alerting if `payment_failed` rate spikes above a baseline threshold (gateway downtime).
3. **Rate Limiting Metrics**:
   - Track HTTP 429 response spikes to detect bot scraping or brute-force checkout abuse.
4. **Connection Pool Utilization**:
   - Monitor `db.Stats()` (in-use connections, wait duration, idle connections) to prevent pool starvation under peak load.
5. **Endpoint Latencies (p95 / p99)**:
   - Monitor `POST /api/v1/bookings/:id/pay` and `GET /api/v1/classes` latency distributions.

---

## What to Do Next with More Time

1. **2-Phase Reservation with Redis TTL**:
   - Implement temporary 10-minute seat holds with Redis keys expiring automatically if the user abandons payment, freeing seats without database locking.
2. **Real Payment Gateway Integration**:
   - Add `midtrans.go` or `stripe.go` implementing `PaymentGateway`, complete with webhook signature verification for asynchronous notifications.
3. **Waitlist Feature**:
   - Allow parents to join a waitlist when a class hits `MaxCapacity`. If a booking is cancelled or payment fails, automatically notify or promote the next student in line.
4. **Calendar & Email Notifications**:
   - Send confirmation emails with `.ics` calendar invites upon payment confirmation.
5. **End-to-End Automated Browser Tests**:
   - Integrate Playwright or Cypress to run automated end-to-end booking flow tests against the HTMX UI.
