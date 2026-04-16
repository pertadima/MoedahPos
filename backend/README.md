# Moedah POS Backend

Moedah POS is a modern, feature-rich point-of-sale system designed for retail and restaurant businesses. This backend provides a robust RESTful API built with Go and PostgreSQL, focusing on high performance, data integrity, and scalability.

## 🚀 Key Features

- **Multi-Store Management**: Effortlessly handle multiple store locations with centralized control.
- **Advanced Inventory**: FIFO-based stock batch tracking, automated stock levels, and movement history.
- **Restaurant Optimization**: Specialized support for table management, menu ingredients (BOM), and KDS (Kitchen Display System) status tracking.
- **Financial Management**: Integrated tracking for incomes, expenses, recurring costs, and cash flow analysis.
- **Granular RBAC**: Flexible Role-Based Access Control allowing precise permission management per store.
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