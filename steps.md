# Phase 1 — Build Steps

Suggested build order: 1 → 2 → 3 → 4 → 5 → 6 → 7 → 8 → 9

Steps 1–5 give you a working API you can test independently.
Steps 6–7 are the background scraping pipeline (goroutine-based, no message broker).
Steps 8–9 complete the user-facing experience.

---

## 1. Project Scaffolding

- [x] Init Go workspace / module structure
- [x] Set up `Makefile` targets (`run/api`, `db/migrations/up`, etc.)
- [x] Set up config — env loader
- [x] `docker-compose.yml` with Postgres (deferred)

---

## 2. Database Layer

- [x] Write goose migrations (users ✅, tokens ✅, products ✅ watchlist ✅, price_history ✅, notifications ✅)
- [x] Set up `sqlc` config and generate typed queries
- [x] Wire up connection pool, ping on startup

---

## 3. Auth (API Service)

- [x] `POST /v1/register` — hash password, insert user
- [x] `POST /v1/login` — verify password, issue JWT access + refresh tokens
- [x] `GET /v1/refresh` — validate refresh token, rotate both tokens
- [x] `PUT /v1/activation` - email confirmation (send emails)
- [x] Auth middleware — parse + verify JWT, inject user ID into context

---

## 4. Product Catalog (API Service)

- [x] `POST /v1/products` — accept Nike product URL, insert into catalog (mock scrape for now)
- [x] `GET /v1/products` — full-text search on name, supports `?category=`, `?in_stock=`, `?min_price=`, `?max_price=` filters + pagination
- [x] `GET /v1/products/:id` — get product details

---

## 5. Watchlist (API Service)

- [x] `POST /v1/watchlist` — add product to watchlist (enforce 5-slot free tier limit)
- [x] `GET /v1/watchlist` — list user's watched products
- [x] `PATCH /v1/watchlist/:id` — update alert preferences
- [x] `DELETE /v1/watchlist/:id` — remove entry

---

## 6. Scheduler + Scraper (Background Goroutine)

Runs inside the API process, started on app startup via `robfig/cron`.
Single interval of 5 minutes.

- [x] On startup, launch background goroutine with `robfig/cron`
- [x] Every 5 minutes: query DB for watched products with `last_scraped_at` older than 5 minutes (`GetStaleProducts` query — joins with watchlist so unwatched products are skipped)
- [x] For each product: hit Nike JSON API using `external_id`, parse response
- [x] Compare fetched data against current DB row — detect `PRICE_DROP`, `RESTOCK`
- [x] Insert new row into `price_history`, update `products.current_price`, `in_stock`, `last_scraped_at`

---

## 7. Notifier (Background Goroutine)

Triggered from within the scraper goroutine after a change is detected.

- [x] Query watchlist for all users watching the affected product
- [x] Filter by user alert preferences (`alert_sale`, `alert_restock`, sizes)
- [x] Send email
- [x] Insert row into `notifications` table

---

## 8. Notification History (API Service)

- [x] `GET /v1/notifications` — paginated list with `?unread=true` filter
- [x] `PATCH /v1/notifications/:id/read` — mark single as read
- [x] `PATCH /v1/notifications/read-all` — mark all as read

---

## API Testing — curl Commands

> Base URL: `http://localhost:8080`
> After login, store the token: `TOKEN=<access_token>` and `REFRESH=<refresh_token>`

---

### Healthcheck

```bash
curl -i http://localhost:8080/v1/healthcheck
```

---

### Auth


**Login**
```bash
curl -i -X POST http://localhost:8080/v1/login \
  -H "Content-Type: application/json" \
  -d 

# Store tokens:
# TOKEN=<access_token>
# REFRESH=<refresh_token>
```

### Products

**Add a product to the catalog** (requires auth)
```bash
curl -i -X POST http:// 
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  
```

**List / search products**
```bash
# All products (default limit 20)
curl -i "http://localhost:8080/v1/products" \
  -H "Authorization: Bearer $TOKEN"

# With full-text search
curl -i "http://localhost:8080/v1/products?q=air+max" \
  -H "Authorization: Bearer $TOKEN"

# Filter by category + in_stock
curl -i "http://localhost:8080/v1/products?category=FOOTWEAR&in_stock=true" \
  -H "Authorization: Bearer $TOKEN"

# Price range + pagination
curl -i "http://localhost:8080/v1/products?min_price=50&max_price=150&page=2&limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

**Get a single product**
```bash
curl -i "http://localhost:8080/v1/products/<product_uuid>" \
  -H "Authorization: Bearer $TOKEN"
```

---

### Watchlist

**Add product to watchlist**
```bash
curl -i -X POST http://localhost:8080/v1/watchlist \
  
   \
  -d '{"product_id":"<product_uuid>","alert_sale":true,"alert_restock":true}'
```

**Get your watchlist**
```bash
curl -i http://localhost:8080/v1/watchlist \
  -H "Authorization: Bearer $TOKEN"
```

**Update watchlist entry alert preferences**
```bash
curl -i -X PATCH "http://localhost:8080/v1/watchlist/<entry_uuid>" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"alert_sale":false,"alert_restock":true}'
```

**Remove from watchlist**
```bash
curl -i -X DELETE "http://localhost:8080/v1/watchlist/<entry_uuid>" \
  -H "Authorization: Bearer $TOKEN"
```

---

### Notifications

**Get all notifications**
```bash
curl -i "http://localhost:8080/v1/notifications" \
  -H "Authorization: Bearer $TOKEN"
```

**Get only unread notifications**
```bash
curl -i "http://localhost:8080/v1/notifications?unread=true" \
  -H "Authorization: Bearer $TOKEN"
```

**Paginate notifications**
```bash
curl -i "http://localhost:8080/v1/notifications?page=2&page_size=10" \
  -H "Authorization: Bearer $TOKEN"
```

**Mark a single notification as read**
```bash
curl -i -X PATCH "http://localhost:8080/v1/notifications/<notification_uuid>/read" \
  -H "Authorization: Bearer $TOKEN"
```

**Mark all notifications as read**
```bash
curl -i -X PATCH http://localhost:8080/v1/notifications/read-all \
  -H "Authorization: Bearer $TOKEN"
```

---

## 9. Next.js Frontend

- [ ] Auth pages: register + login (call API, store tokens)
- [ ] Watchlist dashboard: list, add, remove, edit preferences
- [ ] Basic notification feed (unread count + list)
