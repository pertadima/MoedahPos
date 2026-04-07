# Moedah POS Backend

This is the backend API for the Moedah POS system, built with Go and PostgreSQL.

## Prerequisites

- Go 1.22 or later
- PostgreSQL 16 or later (for local development)
- Docker and Docker Compose (for containerized development)

## Quick Start with Docker (Recommended)

1. **Clone the repository and navigate to the project root:**
   ```bash
   cd /path/to/moedah-pos
   ```

2. **Copy the environment file:**
   ```bash
   cp backend/.env.example backend/.env
   ```

3. **Start the services:**
   ```bash
   docker-compose up --build
   ```

   This will:
   - Start a PostgreSQL database
   - Build and run the Go backend API
   - Run database migrations automatically

4. **Access the API:**
   - API: http://localhost:8080
   - Database: localhost:5432 (user: moedah, password: moedahsecret, db: moedah_pos)

## Local Development Setup

### 1. Install Dependencies

**PostgreSQL:**
- macOS: `brew install postgresql`
- Ubuntu: `sudo apt install postgresql postgresql-contrib`
- Or use Docker: `docker run --name postgres -e POSTGRES_PASSWORD=password -d -p 5432:5432 postgres:16`

**Go:**
- Download from https://golang.org/dl/ or use your package manager

### 2. Set up PostgreSQL Database

```bash
# Start PostgreSQL service (if using system install)
sudo systemctl start postgresql  # Linux
brew services start postgresql   # macOS

# Create database and user
createdb moedah_pos
createuser moedah
psql -c "ALTER USER moedah PASSWORD 'moedahsecret';"
psql -c "GRANT ALL PRIVILEGES ON DATABASE moedah_pos TO moedah;"
```

### 3. Configure Environment

```bash
cd backend
cp .env.example .env
# Edit .env if needed (default values should work for local setup)
```

### 4. Install Go Dependencies

```bash
go mod download
```

### 5. Run Database Migrations

```bash
go run cmd/api/main.go
```

The application will run migrations on startup if `RUN_MIGRATIONS=true` in your `.env` file.

### 6. Run the Application

```bash
go run cmd/api/main.go
```

The API will be available at http://localhost:8080

## Database Schema

See [database.md](database.md) for the complete database schema diagram.

## API Documentation

The API uses REST endpoints. Key features include:

- User authentication and authorization
- Store management
- Product and inventory management
- Purchase orders
- Transaction processing
- Restaurant mode with tables and menu items

## Development Commands

```bash
# Run with hot reload (requires air or similar)
air

# Run tests
go test ./...

# Run linter
golangci-lint run

# Build for production
go build -ldflags="-w -s" -o server ./cmd/api
```

## Environment Variables

See `.env.example` for all available configuration options.

Key variables:
- `APP_ENV`: development/production
- `DB_*`: Database connection settings
- `JWT_SECRET`: Secret key for JWT tokens
- `RUN_MIGRATIONS`: Whether to run migrations on startup

## Troubleshooting

**Database connection issues:**
- Ensure PostgreSQL is running
- Check DB credentials in `.env`
- For Docker: wait for postgres healthcheck

**Port conflicts:**
- Change `APP_PORT` in `.env` if 8080 is in use

**Migration errors:**
- Ensure database exists and user has permissions
- Check migration files in `migrations/` directory