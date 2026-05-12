# Authentication

JWT-based authentication system.

## Flow

1. User submits credentials to `/api/auth/login`
2. Server validates and returns JWT token
3. Client includes token in Authorization header
4. Middleware validates token on protected routes

## Middleware

Located in `backend/internal/middleware/auth.go`

See also: [[API Endpoints]], [[Backend Architecture]]