# Moedah POS API Documentation

Base URL: `/api/v1`

## General Response Formats

**Success Response**
```json
{
  "success": true,
  "data": { ... }
}
```

**Error Responses**

*400 Bad Request*
```json
{
  "success": false,
  "error": "BAD_REQUEST",
  "message": "Invalid input format"
}
```

*401 Unauthorized*
```json
{
  "success": false,
  "error": "UNAUTHORIZED",
  "message": "Missing or invalid token"
}
```

*404 Not Found*
```json
{
  "success": false,
  "error": "NOT_FOUND",
  "message": "Resource not found"
}
```

*500 Internal Server Error*
```json
{
  "success": false,
  "error": "INTERNAL_SERVER_ERROR",
  "message": "An unexpected error occurred"
}
```

---

## 1. Auth Module

### Login
- **Endpoint**: `POST /auth/login`
- **Description**: Authenticate user and receive JWT.
- **Request Body**:
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```
- **Success Response**:
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOi...",
    "user": {
      "id": "uuid",
      "name": "Jane Doe",
      "email": "user@example.com",
      "role": "cashier"
    }
  }
}
```
- **Notes**: Returns 401 if invalid credentials.

### Refresh Token
- **Endpoint**: `POST /auth/refresh`
- **Description**: Refresh an expired JWT.
- **Request Body**:
```json
{
  "refresh_token": "eyJhbGciOi..."
}
```
- **Success Response**:
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOi..."
  }
}
```

### Get Current User
- **Endpoint**: `GET /auth/me`
- **Description**: Get currently logged in user info.
- **Headers**: `Authorization: Bearer <token>`
- **Success Response**:
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "Jane Doe",
    "email": "user@example.com",
    "role": "cashier"
  }
}
```

---

## 2. Product Module

### List Products
- **Endpoint**: `GET /stores/{storeId}/products`
- **Description**: Retrieve a paginated list of products for a store.
- **Headers**: `Authorization: Bearer <token>`
- **Query Params**:
  - `page` (optional, default: 1)
  - `limit` (optional, default: 10)
  - `search` (optional)
- **Success Response**:
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "uuid",
        "name": "Kopi Susu",
        "sku": "KP-001",
        "category_id": "uuid",
        "category_name": "Beverages",
        "price": 25000,
        "cost_price": 10000,
        "use_global_tax": true,
        "tax_percentage": 11,
        "stock": 50,
        "is_active": true
      }
    ],
    "total": 1
  }
}
```

### Create Product
- **Endpoint**: `POST /stores/{storeId}/products`
- **Description**: Create a new product.
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
```json
{
  "name": "Kopi Susu",
  "sku": "KP-001",
  "category_id": "uuid",
  "price": 25000,
  "cost_price": 10000,
  "stock_alert_level": 10,
  "is_active": true
}
```
- **Success Response**:
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "Kopi Susu",
    "sku": "KP-001",
    "category_id": "uuid",
    "price": 25000,
    "cost_price": 10000,
    "is_active": true,
    "created_at": "2024-05-10T10:00:00Z"
  }
}
```

### Update Product
- **Endpoint**: `PUT /stores/{storeId}/products/{productId}`
- **Description**: Update an existing product.
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
```json
{
  "name": "Kopi Susu Large",
  "price": 28000,
  "cost_price": 12000,
  "use_global_tax": true
}
```
- **Success Response**: (Same as Create Product)
- **Notes**: Returns 404 if product ID doesn't exist.

### Delete Product
- **Endpoint**: `DELETE /stores/{storeId}/products/{productId}`
- **Description**: Soft delete or remove a product.
- **Headers**: `Authorization: Bearer <token>`
- **Success Response**:
```json
{
  "success": true,
  "data": null
}
```

---

## 3. Transaction Module

### Create Transaction (Checkout)
- **Endpoint**: `POST /stores/{storeId}/transactions`
- **Description**: Process a POS checkout.
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
```json
{
  "customer_id": "uuid (optional)",
  "payment_method": "CASH",
  "total_amount": 25000,
  "discount_amount": 0,
  "tax_amount": 2750,
  "grand_total": 27750,
  "amount_paid": 50000,
  "change_amount": 22250,
  "items": [
    {
      "product_id": "uuid",
      "quantity": 1,
      "price": 25000,
      "subtotal": 25000
    }
  ]
}
```
- **Success Response**:
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "invoice_number": "INV-20240510-0001",
    "status": "COMPLETED",
    "grand_total": 27750,
    "created_at": "2024-05-10T10:05:00Z"
  }
}
```
- **Notes**: Validates `amount_paid` >= `grand_total` for CASH. Deducts stock automatically.

### Get Transaction History
- **Endpoint**: `GET /stores/{storeId}/transactions`
- **Description**: List past transactions.
- **Headers**: `Authorization: Bearer <token>`
- **Query Params**:
  - `page` (optional)
  - `limit` (optional)
  - `start_date` (optional, YYYY-MM-DD)
  - `end_date` (optional, YYYY-MM-DD)
  - `status` (optional)
- **Success Response**:
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "uuid",
        "invoice_number": "INV-20240510-0001",
        "status": "COMPLETED",
        "grand_total": 27750,
        "payment_method": "CASH",
        "created_at": "2024-05-10T10:05:00Z"
      }
    ],
    "total": 1
  }
}
```

---

## 4. Purchase Order Module

### Create Purchase Order
- **Endpoint**: `POST /stores/{storeId}/purchase-orders`
- **Description**: Draft a new PO.
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
```json
{
  "supplier_id": "uuid",
  "expected_date": "2024-05-15T00:00:00Z",
  "notes": "Weekly supply",
  "items": [
    {
      "product_id": "uuid",
      "quantity": 100,
      "cost_price": 10000
    }
  ]
}
```
- **Success Response**:
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "po_number": "PO-20240510-001",
    "status": "DRAFT",
    "total_amount": 1000000
  }
}
```

### Receive Purchase Order
- **Endpoint**: `POST /stores/{storeId}/purchase-orders/{poId}/receive`
- **Description**: Mark PO as received (full or partial) and add stock.
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
```json
{
  "received_items": [
    {
      "po_item_id": "uuid",
      "received_quantity": 100
    }
  ]
}
```
- **Success Response**:
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "status": "RECEIVED"
  }
}
```
- **Notes**: Updates product inventory based on received quantity and recalculates HPP using FIFO.

### Pay Purchase Order
- **Endpoint**: `POST /stores/{storeId}/purchase-orders/{poId}/payments`
- **Description**: Record payment for a PO.
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
```json
{
  "amount": 1000000,
  "payment_method": "BANK_TRANSFER",
  "reference_number": "TRX-998877",
  "notes": "Paid in full"
}
```
- **Success Response**:
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "po_id": "uuid",
    "amount": 1000000,
    "status": "COMPLETED"
  }
}
```

---

## 5. Income Module

### Log General Income
- **Endpoint**: `POST /stores/{storeId}/incomes`
- **Description**: Record a non-sales income.
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
```json
{
  "amount": 500000,
  "category_id": "uuid",
  "description": "Funding deposit",
  "payment_method": "BANK_TRANSFER",
  "date": "2024-05-10T12:00:00Z"
}
```
- **Success Response**:
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "amount": 500000,
    "status": "COMPLETED"
  }
}
```

### Get Incomes
- **Endpoint**: `GET /stores/{storeId}/incomes`
- **Description**: Retrieve income logs.
- **Headers**: `Authorization: Bearer <token>`
- **Query Params**:
  - `page` (optional)
  - `limit` (optional)
  - `start_date` (optional, YYYY-MM-DD)
  - `end_date` (optional, YYYY-MM-DD)
- **Success Response**:
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "uuid",
        "amount": 500000,
        "category_name": "Deposit",
        "description": "Funding deposit",
        "date": "2024-05-10T12:00:00Z"
      }
    ],
    "total": 1
  }
}
```

---

## 6. Expense Module

### Log Expense
- **Endpoint**: `POST /stores/{storeId}/expenses`
- **Description**: Log a store expense.
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
```json
{
  "amount": 150000,
  "category_id": "uuid",
  "description": "Electricity Bill",
  "payment_method": "CASH",
  "date": "2024-05-10T14:00:00Z"
}
```
- **Success Response**:
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "amount": 150000,
    "status": "PAID"
  }
}
```

### Get Expenses
- **Endpoint**: `GET /stores/{storeId}/expenses`
- **Description**: Retrieve expense logs.
- **Headers**: `Authorization: Bearer <token>`
- **Query Params**:
  - `page` (optional)
  - `limit` (optional)
  - `start_date` (optional, YYYY-MM-DD)
  - `end_date` (optional, YYYY-MM-DD)
- **Success Response**:
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "uuid",
        "amount": 150000,
        "category_name": "Utilities",
        "description": "Electricity Bill",
        "date": "2024-05-10T14:00:00Z"
      }
    ],
    "total": 1
  }
}
```

---

## 7. Cash Flow Module

### Get Cash Flow Report
- **Endpoint**: `GET /stores/{storeId}/reports/cash-flow`
- **Description**: Generate aggregated cash flow metrics.
- **Headers**: `Authorization: Bearer <token>`
- **Query Params**:
  - `start_date` (YYYY-MM-DD)
  - `end_date` (YYYY-MM-DD)
- **Success Response**:
```json
{
  "success": true,
  "data": {
    "total_cash_in": 527750,
    "total_cash_out": 1150000,
    "net_cash": -622250,
    "total_sales_in": 27750,
    "total_other_in": 500000,
    "cash_in_by_method": {
      "CASH": 27750,
      "BANK_TRANSFER": 500000
    },
    "rows": [
      {
        "date": "2024-05-10",
        "cash_in": 527750,
        "cash_out": 1150000,
        "net_cash": -622250,
        "sales_in": 27750,
        "other_in": 500000,
        "cash_in_by_method": {
          "CASH": 27750,
          "BANK_TRANSFER": 500000
        }
      }
    ]
  }
}
```

### Get Cash Flow Detailed Drill-Down
- **Endpoint**: `GET /stores/{storeId}/reports/cash-flow/detail`
- **Description**: Get individual entries affecting cash flow.
- **Headers**: `Authorization: Bearer <token>`
- **Query Params**:
  - `date` (YYYY-MM-DD)
- **Success Response**:
```json
{
  "success": true,
  "data": [
    {
      "type": "SALE",
      "label": "INV-20240510-0001",
      "amount": 27750,
      "payment_method": "CASH",
      "category": null,
      "notes": null,
      "timestamp": "2024-05-10T10:05:00Z"
    },
    {
      "type": "EXPENSE",
      "label": "Utilities",
      "amount": -150000,
      "payment_method": "CASH",
      "category": "Utilities",
      "notes": "Electricity Bill",
      "timestamp": "2024-05-10T14:00:00Z"
    }
  ]
}
```

---

## 8. Activity Log Module

### Get Activity Logs
- **Endpoint**: `GET /stores/{storeId}/activity-logs`
- **Description**: Retrieve audit trail logs.
- **Headers**: `Authorization: Bearer <token>`
- **Query Params**:
  - `page` (optional)
  - `limit` (optional)
  - `user_id` (optional)
  - `module` (optional)
  - `action_type` (optional)
- **Success Response**:
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "uuid",
        "user_id": "uuid",
        "user_name": "Jane Doe",
        "store_id": "uuid",
        "action_type": "CREATE",
        "module": "TRANSACTION",
        "reference_id": "uuid",
        "metadata": {
          "invoice_number": "INV-20240510-0001",
          "total": 27750
        },
        "created_at": "2024-05-10T10:05:00Z"
      }
    ],
    "total": 1
  }
}
```
