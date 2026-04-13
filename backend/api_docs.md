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

**Global Error Responses**
Unless specified otherwise, every endpoint can return:

*400 Bad Request*
```json
{
  "success": false,
  "error": "BAD_REQUEST",
  "message": "Invalid input format or validation failed"
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

*403 Forbidden*
```json
{
  "success": false,
  "error": "FORBIDDEN",
  "message": "Insufficient permissions"
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

### Register
- **Endpoint**: `POST /auth/register`
- **Description**: Register a new user.
- **Request Body**: `{"name": "...", "email": "...", "password": "..."}`
- **Success Response**: `{"success":true,"data":{"id":"uuid"}}`

### Login
- **Endpoint**: `POST /auth/login`
- **Description**: Authenticate user.
- **Request Body**: `{"email": "...", "password": "..."}`
- **Success Response**: `{"success":true,"data":{"token":"jwt...","user":{}}}`

### Refresh Token
- **Endpoint**: `POST /auth/refresh`
- **Description**: Refresh an expired JWT.
- **Request Body**: `{"refresh_token": "..."}`
- **Success Response**: `{"success":true,"data":{"token":"jwt..."}}`

### Logout
- **Endpoint**: `POST /auth/logout`
- **Description**: Invalidate current session.
- **Headers**: `Authorization: Bearer <token>`
- **Success Response**: `{"success":true,"data":null}`

### Get Current User
- **Endpoint**: `GET /auth/me`
- **Description**: Current user info.
- **Headers**: `Authorization: Bearer <token>`
- **Success Response**: `{"success":true,"data":{"id":"uuid","name":"Jane Doe","email":"user@example.com"}}`

---

## 2. Admin Module (SuperAdmin/Admin)

### List Roles
- **Endpoint**: `GET /admin/roles`
- **Description**: Get all available system roles.
- **Headers**: `Authorization: Bearer <token>`
- **Success Response**: `{"success":true,"data":[{"id":"uuid","name":"super_admin","permissions":[]}]}`

### User Management
- **`GET /admin/users`**: List users. Params: `page`, `limit`.
- **`POST /admin/users`**: Create user. Body: `{"name":"", "email":"", "password":"", "stores":[{"store_id":"","role_id":""}]}`
- **`GET /admin/users/{userId}`**: Get user details.
- **`PUT /admin/users/{userId}`**: Update user profile.
- **`POST /admin/users/{userId}/deactivate`**: Toggle user active status.
- **`POST /admin/users/{userId}/reset-password`**: Reset user password. Body: `{"password": ""}`
- **`PUT /admin/users/{userId}/stores`**: Update user store assignments. Body: `{"stores":[]}`

---

## 3. Global Categories (Income & Expense)

*Paths*: `/expense-categories`, `/income-categories`
*Endpoints* (Apply to both):
- **`GET /`**: List categories.
- **`POST /`**: Create category (SuperAdmin only). Body: `{"name":"...", "description":"..."}`
- **`PUT /{id}`**: Update category (SuperAdmin only).
- **`DELETE /{id}`**: Soft delete category (SuperAdmin only).

---

## 4. Suppliers Module

- **`GET /suppliers`**: List all global suppliers. Params: `page`, `limit`, `search`.
- **`POST /suppliers`**: Create supplier. Body: `{"name":"", "contact_name":"", "phone":"", "email":"", "address":""}`
- **`GET /suppliers/{supplierId}`**: Get supplier details.
- **`PUT /suppliers/{supplierId}`**: Update supplier.
- **`DELETE /suppliers/{supplierId}`**: Soft delete supplier.

---

## 5. Stores Module

- **`GET /stores`**: List accessible stores for current user.
- **`POST /stores`**: Create a new store. Body: `{"name":"", "address":"", "store_type":"retail", "currency":"IDR", "default_tax_percentage":11}`
- **`GET /stores/{storeId}`**: Get store context.
- **`PUT /stores/{storeId}`**: Update store info.
- **`DELETE /stores/{storeId}`**: Soft delete store.

### Store Members (`/stores/{storeId}/members`)
- **`GET /`**: List store members.
- **`POST /`**: Add member to store. Body: `{"user_id":"uuid", "role_id":"uuid"}`
- **`PUT /{userId}`**: Update member role. Body: `{"role_id":"uuid"}`
- **`DELETE /{userId}`**: Remove member from store.

---

## 6. Product Module (`/stores/{storeId}`)

### Product Categories
- **`GET /categories`**: List categories.
- **`POST /categories`**: Create category. Body: `{"name":"", "description":""}`
- **`PUT /categories/{categoryId}`**: Update category.
- **`DELETE /categories/{categoryId}`**: Delete category.

### Products
- **`GET /products`**: List products. Params: `page`, `limit`, `search`.
- **`POST /products`**: Create product.
- **`GET /products/barcode/{barcode}`**: Find product by barcode.
- **`GET /products/{productId}`**: Get product details.
- **`PUT /products/{productId}`**: Update product.
- **`DELETE /products/{productId}`**: Soft delete.

### Price History
- **`GET /products/{productId}/price-history`**: Product specific price history.
- **`GET /price-history`**: Store-wide price history changes. Params: `product_id`, `start_date`, `end_date`.

---

## 7. Customers Module (`/stores/{storeId}/customers`)

- **`GET /`**: List customers.
- **`GET /search`**: Search for autocomplete. Query: `q`.
- **`POST /`**: Create customer. Body: `{"name":"", "phone":"", "email":"", "address":""}`
- **`GET /{customerId}`**: Get specific customer.
- **`PUT /{customerId}`**: Update customer.
- **`DELETE /{customerId}`**: Delete customer.

---

## 8. Stock Module (`/stores/{storeId}`)

- **`GET /stock`**: Current stock levels for products.
- **`GET /stock/low`**: Products below minimum stock alert level.
- **`GET /stock/movements`**: History of stock movements IN/OUT.
- **`POST /stock/adjust`**: Manual adjustment. Body: `{"product_id":"", "quantity_change":10, "reason":"Damaged", "type":"OUT"}`
- **`PUT /stock/min`**: Set minimum stock level. Body: `{"product_id":"", "min_stock": 5}`
- **`GET /stock/batches`**: Active FIFO batches.
- **`GET /stock/batch-summary`**: Summary of batch values.
- **`GET /stock/{productId}`**: Detailed stock metrics per product.
- **`GET /adjustments`**: History of manual adjustments.
- **`POST /adjustments`**: Create manual adjustment log.

---

## 9. Transaction Module (`/stores/{storeId}/transactions`)

- **`GET /`**: List completed transactions.
- **`POST /`**: Direct Checkout. Body: `{"payment_method":"CASH","grand_total":100, "items":[]}`
- **`GET /draft`**: List active drafts (saved carts / restaurant tables).
- **`POST /draft`**: Create draft order.
- **`GET /{txnId}`**: Get specific transaction detail.
- **`PUT /{txnId}/draft`**: Update items in draft order.
- **`POST /{txnId}/pay`**: Settle draft/table order. Body: `{"amount_paid":100, "payment_method":"CASH"}`
- **`POST /{txnId}/void`**: Void a completed transaction.

---

## 10. Purchase Order Module (`/stores/{storeId}/purchase-orders`)

- **`GET /`**: List POs.
- **`POST /`**: Create PO draft. Body: `{"supplier_id":"", "expected_date":"", "items":[]}`
- **`GET /payables`**: List outstanding accounts payable.
- **`GET /{poId}`**: Get PO details.
- **`PUT /{poId}`**: Update PO draft.
- **`POST /{poId}/submit`**: Move PO from draft to submitted.
- **`POST /{poId}/receive`**: Receive stock to warehouse. Body: `{"received_items":[]}`
- **`DELETE /{poId}`**: Cancel PO.
- **`GET /{poId}/payments`**: View PO payments.
- **`POST /{poId}/payments`**: Pay PO outright. Body: `{"amount":100, "payment_method":"..."}`

### Termin (Installments)
- **`GET /{poId}/termins`**: List termin schedules.
- **`POST /{poId}/termins`**: Create payment schedule. Body: `{"termins":[{"due_date":"", "amount":0}]}`
- **`POST /{poId}/termins/{terminId}/payments`**: Pay specific termin installment.
- **`GET /{poId}/debt`**: Get debt summary for PO.
- **`GET /{poId}/document`**: Generate PO document metadata.

---

## 11. Reports Module (`/stores/{storeId}/reports`)

- **`GET /sales`**: Overall sales summary. Query: `start_date`, `end_date`.
- **`GET /sales/by-product`**: Sales grouped by product.
- **`GET /sales/by-cashier`**: Sales grouped by user.
- **`GET /stock-valuation`**: Total valuation of current stock based on average cost.
- **`GET /profit`**: Net profit calculations over time.

---

## 12. Income & Expense (`/stores/{storeId}`)

### Incomes
- **`GET /incomes`**: List non-POS incomes.
- **`POST /incomes`**: Log income. Body: `{"amount":100, "category_id":""}`
- **`PUT /incomes/{id}`**: Update income entry.
- **`DELETE /incomes/{id}`**: Delete income entry.

### Expenses & Recurring
- **`GET /expenses`**: List expenses.
- **`POST /expenses`**: Log expense.
- **`PUT /expenses/{id}`**: Update expense.
- **`DELETE /expenses/{id}`**: Delete expense.
- **`PATCH /expenses/{id}/status`**: Change expense lifecycle status.
- **`GET /recurring-expenses`**: List subscriptions/recurring rules.
- **`POST /recurring-expenses`**: Create recurring expense job.
- **`PUT /recurring-expenses/{id}`**: Update recurring job.
- **`DELETE /recurring-expenses/{id}`**: Delete recurring job.

---

## 13. Cash Flow Module (`/stores/{storeId}/reports/cash-flow`)

- **`GET /`**: Summarized Cash Flow dashboard stats (Total Cash In / Out, Net Cash).
- **`GET /detail`**: Granular list of every financial record affecting cash float for a specific day.

---

## 14. Activity Log (`/stores/{storeId}/activity-logs`)

- **`GET /`**: Full system audit trail. Params: `user_id`, `module`, `action_type`, `start_date`, `end_date`. Returns action metadata and references to modified rows.

---

## 15. Restaurant Mode (`/stores/{storeId}`)

*Note: These only work if store type is 'restaurant'.*

### Tables
- **`GET /tables`**: List all tables and occupancy status.
- **`POST /tables`**: Add table. Body: `{"name":"Table 1", "capacity":4}`
- **`PUT /tables/{tableId}`**: Update details.
- **`PUT /tables/{tableId}/status`**: Update table state directly.
- **`DELETE /tables/{tableId}`**: Remove table.

### Menu Items
- **`GET /menu-items`**: List restaurant menu offerings.
- **`POST /menu-items`**: Create menu. Body combines product tracking with menu visibility.
- **`PUT /menu-items/{menuItemId}`**: Update menu item.
- **`DELETE /menu-items/{menuItemId}`**: Remove menu item.

### Kitchen Display System (KDS)
- **`GET /kds/tickets`**: List active kitchen tickets (draft transaction items).
- **`PUT /kds/items/{itemId}`**: Move item status (e.g. PENDING -> PREPARING -> COMPLETED). Body: `{"status": "PREPARING"}`
