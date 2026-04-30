# Moedah POS Backend

Moedah POS is a modern, feature-rich point-of-sale system designed for retail and restaurant businesses. This backend provides a robust RESTful API built with Go and PostgreSQL, focusing on high performance, data integrity, and scalability.

## 🚀 Key Features

- **Multi-Store Management**: Effortlessly handle multiple store locations with centralized control.
- **Advanced Inventory**: FIFO-based stock batch tracking, automated stock levels, and movement history.
- **Restaurant Optimization**: Specialized support for table management, menu ingredients (BOM), and KDS (Kitchen Display System) status tracking.
- **Financial Management**: Integrated tracking for incomes, expenses, recurring costs, and cash flow analysis.
- **Granular RBAC**: Flexible Role-Based Access Control using a modular naming convention (`{module}:{action}`). Allows precise permission management (read, write, update, delete) per system module.
- **Comprehensive Auditing**: Detailed activity logs for all critical system actions.

## 🛠 Tech Stack & Dependencies

- **Language**: [Go 1.22+](https://golang.org/)
- **Database**: [PostgreSQL 16+](https://www.postgresql.org/)
- **API Router**: [Chi v5](https://github.com/go-chi/chi) - Minimalist and idiomatic.
- **SQL Extensions**: [Sqlx](https://github.com/jmoiron/sqlx) - For cleaner database interactions.
- **Migrations**: [Goose](https://github.com/pressly/goose) - Reliable database versioning.
- **Validation**: [Validator v10](https://github.com/go-playground/validator) - Strict request body validation.
- **Logging**: [Zerolog](https://github.com/rs/zerolog) - Fast, structured JSON logging.
- **Security**: JWT for authentication and Argon2/Bcrypt for password hashing.

---

## 🏛 Architecture

The backend is built using a strict layered architecture pattern to enforce separation of concerns, improve testability, and keep the application strictly decoupled.

1. **Handler/Delivery Layer (`/internal/handler`)**: Receives HTTP input, extracts variables via the `Chi` router, validates raw JSON requests via `validator`, and passes them to the Service layer.
2. **Service Layer (`/internal/service`)**: Contains all core business logic, RBAC processing, formatting, and calculation logic. Services use interfaces to talk to Repositories, making them 100% mockable for testing.
3. **Repository Layer (`/internal/repository`)**: Solely responsible for interacting with the `PostgreSQL` database utilizing `sqlx`. No business logic exists here.
4. **Domain & DTO (`/internal/domain`, `/internal/dto`)**: Contains canonical entity structs mapping directly to DB schemas and data transfer objects for standardized JSON API requests/responses.

---

## 🔌 Core API Endpoints

The application relies on REST-compliant routing structures grouped by domains. *(Note: All endpoints aside from public Auth require a valid `Bearer` JSON Web Token).*

**Auth & Users**
- `POST /api/v1/auth/login` - Authenticate user
- `POST /api/v1/auth/register` - Create new admin tenant
- `POST /api/v1/auth/refresh` - Cycle access tokens
- `GET /api/v1/auth/me` - Read session info

**Store & Sub-tenants**
- `GET/POST /api/v1/stores`
- `GET/PUT/DELETE /api/v1/stores/{storeId}`
- `GET/POST/PUT /api/v1/stores/{storeId}/members` - Manage RBAC assignments per store

**Core Retail Module**
- `CRUD /api/v1/stores/{storeId}/categories` 
- `CRUD /api/v1/stores/{storeId}/products`
- `CRUD /api/v1/stores/{storeId}/customers`
- `GET /api/v1/stores/{storeId}/stock/levels`
- `POST /api/v1/stores/{storeId}/stock/adjust` - Mutates stocks with tracked audit history

**Offline Sync Engine**
- `GET /api/v1/stores/{storeId}/sync/pull` - Fetch delta-compressed database rows using `?since={unix_timestamp}`.

**Transaction & Checkout Module**
- `POST /api/v1/stores/{storeId}/transactions` - Execute final checkout
- `POST /api/v1/stores/{storeId}/transactions/drafts` - Save active carts into holds

**Financial Modules (B2B)**
- `CRUD /api/v1/suppliers`
- `CRUD /api/v1/stores/{storeId}/purchase-orders`
- `CRUD /api/v1/stores/{storeId}/expenses`
- `CRUD /api/v1/stores/{storeId}/incomes`

---

## 🚦 Getting Started

### Prerequisites

- Go 1.22 or later
- PostgreSQL 16 or later
- Docker & Docker Compose (optional, for containerized setup)

### Option 1: Quick Start with Docker (Recommended)

1. **Clone and navigate:**
   ```bash
   cd moedah-pos/backend
   ```

2. **Configure environment:**
   ```bash
   cp .env.example .env
   ```

3. **Launch services:**
   ```bash
   docker-compose up --build
   ```
   *This automatically handles database setup, migrations, and starts the API on port `8080`.*

### Option 2: Local Development

1. **Setup Database:**
   Ensure PostgreSQL is running and create the database:
   ```sql
   CREATE DATABASE moedah_pos;
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Run Migrations:**
   The app runs migrations automatically on startup if `RUN_MIGRATIONS=true` is set in `.env`. Alternatively:
   ```bash
   go run cmd/api/main.go
   ```

4. **Start the API:**
   ```bash
   go run cmd/api/main.go
   ```
   API will be available at `http://localhost:8080`.

---

## 📊 Database ERD

The database schema is maintained in a dedicated Mermaid file for better version control and visibility.

- **Diagram File**: [database.mmd](database.mmd)

---

## 🛠 Development Commands

```bash
# Hot reload development
air

# Run all tests
go test ./... -v

# Code linting
golangci-lint run

# Build for production
go build -o server ./cmd/api
```

---

## 📜 License

This project is proprietary and confidential. Unauthorized copying is strictly prohibited.