# Database Schema

PostgreSQL database with migrations managed by goose.

## Tables

- `products` - Product catalog
- `transactions` - POS sales records
- `users` - Staff and admin accounts
- `customers` - Customer profiles
- `loyalty_points` - Loyalty program tracking

## Migrations

Located in `backend/migrations/` with 40+ migration files covering schema evolution.

See also: [[Backend Architecture]], [[Loyalty System]]