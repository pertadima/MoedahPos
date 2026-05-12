# Backend Architecture

The Go backend follows a layered architecture pattern:

## Layers

- **Handlers** (`internal/handler/`) - HTTP request handling
- **Services** (`internal/service/`) - Business logic
- **Repositories** (`internal/repository/`) - Data access
- **Domain** (`internal/domain/`) - Entity definitions

## Key Components

- [[API Endpoints]] via Chi router
- [[Authentication]] middleware
- Database migrations via goose

See also: [[Database Schema]], [[Configuration]]