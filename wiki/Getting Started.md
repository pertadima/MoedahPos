# Getting Started

## Prerequisites

- Go 1.23+
- Node.js 22+
- PostgreSQL
- Docker (optional)

## Backend Setup

```bash
cd backend
go mod download
go run cmd/api/main.go
```

## Frontend Setup

```bash
cd frontend
npm install
npm run dev
```

## Docker Setup

Use [[Deployment]] with docker-compose.

See also: [[Configuration]], [[Project Overview]]