# API Endpoints

## REST API

The backend exposes REST endpoints via Chi router:

- `POST /api/auth/login` - User authentication
- `GET /api/products` - List products
- `POST /api/transactions` - Create transaction
- `GET /api/customers` - List customers
- `GET /api/reports/*` - Various reports

## Authentication

Uses JWT tokens. See [[Authentication]] for details.

See also: [[Backend Architecture]]