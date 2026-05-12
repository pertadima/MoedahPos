# Testing

## Backend Tests

Go tests in each package:
```bash
go test ./...
```

Run with coverage:
```bash
go test -coverprofile=coverage.out ./...
```

## Frontend Tests

```bash
cd frontend
npm run lint
npm run type-check
```

See also: [[Backend Architecture]]