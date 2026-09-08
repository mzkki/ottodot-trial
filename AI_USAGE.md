# AI Usage Report (`AI_USAGE.md`)

This document outlines how AI tools were utilized during the design, development, testing, and refinement of the Ottodot Trial Class Booking system, including places of disagreement, architectural corrections, workflow reflections, and verification methodology.

---

## 1. Which AI Tools Were Used

- **Antigravity CLI / Claude 3.5 Sonnet / Gemini 3.8**: Used as the primary agentic pair programmer for repository scaffolding, Clean Architecture structuring, refactoring, and test suite generation.
- **GitHub Copilot / LLM Code Completion**: Used for inline auto-completion of repetitive Go boilerplate (error checking, struct mapping, SQL statements, and HTML template markup).

---

## 2. What AI Was Used For

1. **Architecture & Boilerplate Scaffolding**:
   - Analyzing existing Go reference boilerplates (`mauth`) to adopt standardized practices: Wire dependency injection, Gin HTTP routing, zap logging, and GORM repository patterns.
2. **Exhaustive Unit & Concurrency Test Generation**:
   - Synthesizing comprehensive mock repositories (`mock_payment.go`, mock repositories in `booking_usecase_mock_test.go`) to achieve **100.0% statement test coverage** across all branches in `internal/usecase`.
   - Designing unit test fixtures that reproduce the exact last-seat concurrency race condition (`TestBookingUseCaseRaceConditionDuringPayment`).
3. **Database Migrations & Seed Automation**:
   - Creating initial PostgreSQL migration scripts with foreign keys, compound indexes, and unique partial indexes (`idx_unique_active_booking`).
   - Scripting the automated database reset tooling (`scripts/db-fresh.sh`) supporting dual isolated databases (`db_ottodot_trial` and `db_ottodot_trial_test`).
4. **Interactive API Documentation & Security Middleware**:
   - Generating Swag / OpenAPI 2.0 / 3.0 annotations on HTTP handlers and integrating the modern Scalar UI at `/docs` and `/api/v1/docs/`.
   - Implementing security middleware including IP-based rate limiting to protect public booking and payment endpoints from brute-force and resource starvation attacks.

---

## 3. One Place Where AI Helped Move Faster

### **Exhaustive Branch Coverage & Concurrency Testing**
Writing tests for Clean Architecture use cases in Go often involves an enormous amount of boilerplate: setting up mock repository implementations, stubbing individual error returns, simulating database transaction failures, and asserting state transitions.

AI dramatically accelerated:
- Generating exhaustive test cases for every failure branch in `booking_usecase.go` (e.g., repository query errors, transaction rollback triggers, gateway timeouts, context cancellation).
- Writing the concurrent test simulating the **last-seat race condition**, where two bookings compete for the single remaining seat.

What would normally take 4–6 hours of manual mock coding was accomplished in under an hour, resulting in **100.0% statement coverage** on the core usecase package.

---

## 4. Places Where I Disagreed With, Corrected, or Rejected AI Output

### **1. Coupling Payment Service Directly Inside the Usecase Layer**
- **The Issue:** In the initial generation, the AI coupled payment execution logic directly inside `internal/usecase`, mixing business rules with third-party payment gateway mechanics.
- **The Correction:** Rejected this coupling to enforce strict **Clean Architecture & Dependency Inversion (DIP)**. I separated the payment concern entirely from the core usecase:
  - The business layer (`internal/usecase`) now interacts solely via an abstract `PaymentGateway` consumer interface.
  - A standalone payment service / adapter was created under `internal/infrastructure/payment/`, allowing pluggable, direct integration with external payment providers (e.g., Stripe, Midtrans) or mocks without polluting domain rules.

### **2. N+1 Query in Class Listing (Memory & Latency Spike)**
- **The Issue:**
  - **Location:** `internal/delivery/http/booking/handler.go:80-82`
  - **Details:** In `GetClasses`, for every class returned, the handler executed `CountConfirmedForClass(ctx, cl.ID)`. Listing 20 classes triggered 21 sequential database round-trips, introducing severe latency and connection pool pressure.
- **The Correction:** Refactored the data retrieval mechanism to batch-query confirmed booking counts using a single `GROUP BY trial_class_id` aggregation query, reducing complexity from $O(N)$ database round-trips to $O(1)$.

### **3. Missing Endpoint Rate Limiting & Abuse Prevention**
- **The Issue:** The initial routing implementation exposed open booking and payment confirmation endpoints without throughput safeguards, leaving the application vulnerable to seat hoarding bots, race exploitation, and denial-of-service (DoS).
- **The Correction:** Enforced an IP-based token-bucket/sliding-window rate-limiting middleware across sensitive HTTP endpoints to throttle rapid repetitive hits, protect PostgreSQL connections, and isolate external payment provider quotas.

### **4. Test and Runtime Database Collision**
- **The Issue:** The AI initially executed test runners against default environment variables, risking test pollution on local runtime seed data.
- **The Correction:** Enforced strict database separation (`db_ottodot_trial` vs `db_ottodot_trial_test`) and updated `scripts/db-fresh.sh` to ensure testing and demo environments run in complete isolation.

---

## 5. What I Would Change About My AI Workflow Next Time

If I were to approach this challenge again with an AI pair programmer:

1. **Interface-First & Schema-First Prompting**:
   - Define all Go interfaces (`usecase/interfaces.go`), gateway abstractions, and SQL schema contracts *before* asking the AI to write any implementation or repository code. This prevents duplicate type declarations, eliminates refactoring churn, and keeps boundaries crisp from minute one.
2. **Batch Query & Performance Constraints in Initial Prompt**:
   - Explicitly instruct the AI up-front to avoid looping queries ($N+1$) in handlers/usecases and require aggregate queries (`GROUP BY`, bulk lookups) as a non-negotiable acceptance criterion.
3. **Smaller, Single-Responsibility Iteration Prompts**:
   - Rather than asking the AI to implement multi-layer features (DB + Usecase + Handlers + Templates) in single passes, break tasks down into:
     1. Database schema & migrations
     2. Infrastructure adapters (DB repos, payment provider integrations)
     3. Usecase business logic + unit tests
     4. Delivery layer (HTTP handlers, rate limiting, middleware)
     5. UI templates / Docs
   - Verifying each step with `go test` and `go vet` before moving to the next layer drastically reduces debugging time.

---

## 6. How the Final Implementation Was Verified

The solution underwent rigorous automated and manual verification:

### **A. Automated Testing & 100.0% Statement Coverage**
- **Test Execution**: Ran `make test` targeting the isolated test database (`db_ottodot_trial_test`).
- **Usecase Statement Coverage**:
  ```text
  === Core Business Logic Coverage (internal/usecase) ===
  internal/usecase/booking_usecase.go:131:    NewBookingUseCase                100.0%
  internal/usecase/booking_usecase.go:152:    NewBookingUseCaseWithInterfaces  100.0%
  internal/usecase/booking_usecase.go:175:    GetAllParents                    100.0%
  internal/usecase/booking_usecase.go:180:    GetStudentsByParent              100.0%
  internal/usecase/booking_usecase.go:185:    GetAvailableClasses              100.0%
  internal/usecase/booking_usecase.go:191:    GetAvailableClassesWithStatus    100.0%
  internal/usecase/booking_usecase.go:243:    GetActiveBookingsByStudent       100.0%
  internal/usecase/booking_usecase.go:256:    GetAllClasses                    100.0%
  internal/usecase/booking_usecase.go:261:    GetClassRoster                   100.0%
  internal/usecase/booking_usecase.go:266:    GetClassByID                     100.0%
  internal/usecase/booking_usecase.go:271:    GetBookingByID                   100.0%
  internal/usecase/booking_usecase.go:276:    CountConfirmedForClass           100.0%
  internal/usecase/booking_usecase.go:281:    CreateBooking                    100.0%
  internal/usecase/booking_usecase.go:327:    ProcessPayment                   100.0%
  ---------------------------------------------------------
  total:                                      (statements)                     45.1%
  ```
- **Concurrency Test (`TestBookingUseCaseRaceConditionDuringPayment`)**: Verified that when two pending bookings attempt to confirm the final slot in a class of capacity 1, the first request succeeds while the second fails with `ErrClassFull` and marks the booking `cancelled`.
- **Router & Middleware Tests**: Verified all HTTP endpoints, rate limiting, request ID tracing, panic recovery, and template rendering via `cmd/api/app_test.go` and `internal/delivery/http/middleware/middleware_test.go`.

### **B. Database Integrity & Reseed Verification**
- Executed `make db-fresh-all` to drop and recreate both databases.
- Verified via SQL queries that:
  - 4 parents and 8 children are seeded.
  - Trial Class 1 ("Introduction to Robotics") has 3 pre-seeded confirmed bookings, leaving **exactly 1 slot remaining** for live race-condition demonstrations.
  - Unique partial index `idx_unique_active_booking` rejects duplicate pending/confirmed bookings for the same child and class.

### **C. End-to-End Manual & UI Verification**
- Started the server via `make run` on `http://localhost:8080`.
- **Parent Switcher & Dynamic Catalog**: Toggled between Sarah Johnson, Michael Chen, Amanda Miller, and David Tan using HTMX; confirmed child list and class availability updated without full page reloads.
- **Payment Success & Failure Flows**:
  - Booked a seat and clicked **Pay & Confirm** (`simulate_failure = false`) → verified status turned green (`confirmed`) and student appeared on `/admin/roster`.
  - Clicked **Simulate Failed Payment** (`simulate_failure = true`) → verified status changed to `payment_failed` with a descriptive reason and a retry button.
- **Rate Limit Verification**: Triggered burst requests on the booking endpoints; confirmed HTTP `429 Too Many Requests` responses when thresholds were breached.
- **Interactive Documentation**: Opened `http://localhost:8080/docs` and confirmed the Scalar API Reference renders all endpoints, models, and example payloads interactively.
