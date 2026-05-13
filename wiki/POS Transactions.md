# POS Transactions

Core business flow for point-of-sale operations.

## Transaction Flow

1. Select products
2. Apply discounts (item/cart level)
3. Calculate totals with tax
4. Process payment
5. Update inventory
6. Award loyalty points (if customer identified)

## Key Entities

- `Transaction` - Sale record
- `TransactionItem` - Individual line items
- `PaymentRecord` - Payment details

See also: [[Database Schema]], [[Loyalty System]]