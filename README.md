# KickAlert — Nike Product Alert SaaS

> Get notified the moment your favourite Nike products go on sale or restock.

---

## Table of Contents

- [Motivation](#motivation)
- [Quick Start](#quick-start)
- [Usage](#usage)
- [Overview](#overview)
- [Architecture](#architecture)
- [How the Scraper Works](#how-the-scraper-works)
- [Database Schema](#database-schema)
- [Project Structure](#project-structure)
- [Tech Stack](#tech-stack)
- [Freemium Tier Design](#freemium-tier-design)
- [API Reference](#api-reference)
- [Environment Variables](#environment-variables)
- [Local Development](#local-development)
- [Build Phases](#build-phases)
- [Contributing](#contributing)

---

## Motivation

Sneaker drops and Nike sales sell out in minutes. Manually refreshing product pages is tedious and unreliable. KickAlert was built to automate that — scraping Nike every 5 minutes and sending an email the moment a price drops or a sold-out product comes back in stock, so you never miss a drop again.

---

## Quick Start

**Prerequisites:** Go 1.22+, PostgreSQL, an [Apify](https://apify.com) account (for the Nike scraper actor), and an SMTP server.

```bash
# 1. Clone the repo
git clone https://github.com/your-username/kick-alert.git
cd kick-alert

# 2. Copy the example env file and fill in your values
cp .env.example .env

# 3. Run database migrations
make db/migrations/up

# 4. Start the API (background scheduler included)
make run/api
```

The API will be available at `http://localhost:4000`.

---

## Usage

### Register and activate an account

```bash
# Register
curl -X POST http://localhost:4000/v1/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Jane Doe","email":"jane@example.com","password":"pa$$word123"}'

# Activate (token is sent to your email)
curl -X PUT http://localhost:4000/v1/activation \
  -H "Content-Type: application/json" \
  -d '{"token":"<activation_token>"}'
```

### Log in and watch a product

```bash
# Log in — returns access and refresh tokens
curl -X POST http://localhost:4000/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email":"jane@example.com","password":"pa$$word123"}'

# Add a Nike product to the catalog by URL
curl -X POST http://localhost:4000/v1/products \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{"product_url":"https://www.nike.com/t/air-max-90-shoes/..."}'

# Add the product to your watchlist
curl -X POST http://localhost:4000/v1/watchlist \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{"product_id":"<product_uuid>","alert_sale":true,"alert_restock":true}'
```

From here the background scheduler takes over — you will receive an email whenever a price drop or restock is detected.

---

## Overview

KickAlert is a SaaS that monitors Nike for price drops and restocks — then notifies subscribed users via email. Supports footwear, apparel, and equipment.

**Core user flow:**
1. User signs up and activates their account via email
2. User submits a Nike product URL; the API scrapes it via Apify and adds it to the catalog
3. User adds the product to their watchlist with optional alert preferences
4. A background scheduler re-scrapes every 5 minutes
5. When a price drop or restock is detected, an email notification is dispatched and logged

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        FRONTEND                                  │
│                    Next.js (App Router)                          │
│        Auth · Dashboard · Watchlist · Notification Feed          │
└──────────────────────────┬──────────────────────────────────────┘
                           │ HTTPS / REST
┌──────────────────────────▼──────────────────────────────────────┐
│                      API SERVICE (Go)                            │
│         /auth  /products  /watchlist  /notifications             │
│                     JWT · Rate Limiting                          │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │              Background Goroutines (same process)        │    │
│  │                                                          │    │
│  │   robfig/cron → Scheduler → Apify Scraper → Notifier    │    │
│  └─────────────────────────────────────────────────────────┘    │
└──────────────────────────┬──────────────────────────────────────┘
                           │ reads / writes
                           ▼
              ┌────────────────────────┐
              │       PostgreSQL        │
              │                        │
              │  users                 │
              │  tokens                │
              │  products              │
              │  watchlist             │
              │  price_history         │
              │  notifications         │
              └────────────────────────┘
```

Everything runs in a single Go binary. The scheduler, scraper, and notifier are background goroutines — no message broker needed.

---

## How the Scraper Works

The scraper runs as a goroutine launched on app startup using `robfig/cron`.

```
Every 5 minutes:
  1. Query DB for all products
  2. For each product: call Apify Nike actor using the product's external_id
  3. Compare fetched price/stock against current DB row
  4. If price changed:
     a. Insert row into price_history
     b. Update products.current_price, in_stock, last_scraped_at
     c. Query watchlist for users watching this product
     d. Filter by user alert preferences (alert_sale, alert_restock)
     e. Send email notification via SMTP
     f. Insert row into notifications table
```

Scrape interval: **every 5 minutes** for all tiers. Tier differences are enforced at the watchlist level.

---

## Database Schema

```sql
-- Core users
CREATE TABLE users (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email           citext UNIQUE NOT NULL,
  password_hash   TEXT NOT NULL,
  name            TEXT,
  activated       BOOLEAN DEFAULT false,
  notify_email    BOOLEAN DEFAULT true,
  notify_push     BOOLEAN DEFAULT false,
  tier            TEXT DEFAULT 'free',   -- 'free' | 'pro'
  created_at      TIMESTAMPTZ DEFAULT now(),
  updated_at      TIMESTAMPTZ DEFAULT now()
);

-- Auth tokens (activation + refresh)
CREATE TABLE tokens (
  hash       TEXT PRIMARY KEY,
  user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  expiry     TIMESTAMPTZ NOT NULL,
  scope      TEXT NOT NULL   -- 'activation' | 'refresh'
);

-- Normalised Nike product catalog (shared across all users)
CREATE TABLE products (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  slug            TEXT UNIQUE NOT NULL,
  name            TEXT NOT NULL,
  sku             TEXT NOT NULL,
  external_id     TEXT UNIQUE NOT NULL,   -- Nike cloudProductId
  category        TEXT NOT NULL DEFAULT 'FOOTWEAR',
  url             TEXT NOT NULL,
  image_url       TEXT,
  current_price   TEXT,
  currency        TEXT DEFAULT 'USD',
  in_stock        BOOLEAN,
  last_scraped_at TIMESTAMPTZ,
  created_at      TIMESTAMPTZ DEFAULT now()
);

-- User watch preferences per product
CREATE TABLE watchlist (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  product_id    UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  alert_sale    BOOLEAN DEFAULT true,
  alert_restock BOOLEAN DEFAULT true,
  created_at    TIMESTAMPTZ DEFAULT now(),
  UNIQUE(user_id, product_id)
);

-- Full price + stock history (append-only)
CREATE TABLE price_history (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  price      TEXT,
  in_stock   BOOLEAN,
  scraped_at TIMESTAMPTZ DEFAULT now()
);

-- Sent notifications log
CREATE TABLE notifications (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  product_id   UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  watchlist_id UUID NOT NULL REFERENCES watchlist(id) ON DELETE CASCADE,
  type         notification_type NOT NULL,   -- 'PRICE_DROP' | 'RESTOCK'
  old_price    TEXT,
  new_price    TEXT,
  read         BOOLEAN DEFAULT false,
  created_at   TIMESTAMPTZ DEFAULT now()
);
```

---

## Project Structure

```
kick-alert/
├── cmd/api/
│   ├── main.go            # Entry point, config loading, starts server
│   ├── server.go          # HTTP server with graceful shutdown
│   ├── routes.go          # Gin router and middleware setup
│   ├── middleware.go       # JWT auth + per-IP rate limiting
│   ├── errors.go          # Unified error response helpers
│   ├── healthcheck.go     # GET /v1/healthcheck
│   ├── auth.go            # Login, token refresh, activation handlers
│   ├── users.go           # Register handler
│   ├── products.go        # Product catalog handlers + Apify scrape on add
│   ├── watchlist.go       # Watchlist CRUD handlers
│   ├── notifications.go   # Notification retrieval + read handlers
│   ├── scheduler.go       # Background scraper + notifier goroutine
│   └── apify.go           # Apify Nike actor client
├── internal/
│   ├── auth/
│   │   └── auth.go        # JWT creation/validation, Argon2id hashing, token generation
│   ├── database/          # sqlc-generated type-safe query code
│   │   ├── db.go
│   │   ├── models.go
│   │   ├── users.sql.go
│   │   ├── tokens.sql.go
│   │   ├── products.sql.go
│   │   ├── watchlist.sql.go
│   │   ├── notifications.sql.go
│   │   └── price_history.sql.go
│   └── mailer/
│       ├── mailer.go      # SMTP email client with retry logic
│       └── templates/
│           ├── user_welcome.tmpl
│           └── price_alert.tmpl
├── sql/
│   ├── schema/            # goose migrations (6 files)
│   └── queries/           # sqlc source queries
├── sqlc.yaml
├── Makefile
└── .env
```

---

## Tech Stack

| Layer | Choice | Reason |
|---|---|---|
| Language | **Go** | Cheap goroutines, great for background workers |
| Frontend | **Next.js** (App Router) | SSR, great DX |
| Database | **PostgreSQL** | Relational integrity, battle-tested |
| DB Queries | **sqlc** | Type-safe SQL, no ORM magic |
| Migrations | **goose** | File-based, CI-friendly |
| HTTP Router | **Gin** | Lightweight, composable middleware |
| Job Scheduling | **robfig/cron** | Battle-tested Go cron library |
| Email | **go-mail + SMTP** | Standard SMTP, works with any provider |
| Scraping | **Apify** | Managed Nike scraper actor |
| Auth | **JWT + Argon2id** | Stateless access tokens + secure password hashing |
| Config | **env vars + godotenv** | 12-factor app compliant |

---

## Freemium Tier Design

| Feature | Free | Pro |
|---|---|---|
| Watchlist slots | 5 products | Unlimited |
| Scrape frequency | Every 5 min | Every 5 min |
| Alert channels | Email | Email |
| Price history | Full | Full |

---

## API Reference

All authenticated endpoints require `Authorization: Bearer <access_token>`.

### Auth

```
POST  /v1/register          Create a new account
                            Body: { name, email, password }

POST  /v1/login             Get access + refresh tokens
                            Body: { email, password }

GET   /v1/refresh           Rotate tokens using refresh token
                            Header: Authorization: Bearer <refresh_token>

PUT   /v1/activation        Activate account
                            Body: { token }
```

### Products (authenticated)

```
POST  /v1/products          Scrape and add a product by Nike URL
                            Body: { product_url }

GET   /v1/products          Search catalog
                            Query: ?q=&category=&in_stock=&min_price=&max_price=&page=&limit=

GET   /v1/products/:id      Get product details
```

### Watchlist (authenticated)

```
POST   /v1/watchlist        Add a product to watchlist
                            Body: { product_id, alert_sale, alert_restock }

GET    /v1/watchlist        Get all watched products with current price/stock

PATCH  /v1/watchlist/:id    Update alert preferences
                            Body: { alert_sale, alert_restock }

DELETE /v1/watchlist/:id    Remove from watchlist
```

### Notifications (authenticated)

```
GET    /v1/notifications            Get notification history
                                    Query: ?page=&limit=&unread=true

PATCH  /v1/notifications/:id/read   Mark a notification as read

PATCH  /v1/notifications/read-all   Mark all notifications as read
```

### Healthcheck

```
GET   /v1/healthcheck       Returns status, environment, and version
```

---

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `PORT` | Yes | Server port (e.g. `4000`) |
| `ENV` | Yes | `development` or `production` |
| `KICK_ALERT_DB_DSN` | Yes | PostgreSQL connection string |
| `DB_MAX_OPEN_CONNS` | Yes | Max open DB connections |
| `DB_MAX_IDLE_CONNS` | Yes | Max idle DB connections |
| `DB_MAX_IDLE_TIME` | Yes | Connection idle timeout in minutes |
| `JWT_SECRET` | Yes | Secret key for signing JWTs |
| `FRONTEND_ACTIVATION_URL` | Yes | Base URL for activation email link |
| `SMTP_HOST` | Yes | SMTP server hostname |
| `SMTP_PORT` | Yes | SMTP server port |
| `SMTP_USERNAME` | Yes | SMTP username |
| `SMTP_PASSWORD` | Yes | SMTP password |
| `SMTP_SENDER` | Yes | Sender address (e.g. `KickAlert <no-reply@example.com>`) |
| `APIFY_TOKEN` | Yes | Apify API token for the Nike scraper actor |
| `LIMITER_ENABLED` | No | Enable rate limiting (default: `false`) |
| `LIMITER_RPS` | No | Requests per second per IP (default: `2`) |
| `LIMITER_BURST` | No | Burst size (default: `4`) |

---

## Local Development

```bash
# Run migrations
make db/migrations/up

# Start API (includes background scheduler)
make run/api
```

---

## Build Phases

### Phase 1 — MVP

- [x] PostgreSQL schema + goose migrations
- [x] User registration with email activation
- [x] JWT auth with access + refresh tokens
- [x] Product catalog with Apify scraping
- [x] Watchlist CRUD with tier-based limits
- [x] Background scheduler + price change detection
- [x] Email notifications via SMTP
- [x] Notifications endpoint with read/unread state
- [ ] Next.js: auth pages + watchlist dashboard

### Phase 3 — Growth & Monetisation

- [ ] Stripe integration (free → pro upgrade)
- [ ] Webhook delivery (Discord, Telegram, custom URL)

---

## Contributing

Contributions are welcome! Here is how to get started:

1. **Fork** the repository and create a feature branch off `master`.
2. **Set up** your local environment following the [Quick Start](#quick-start) guide.
3. **Make your changes** — keep commits focused and atomic.
4. **Run the tests** before opening a PR:
   ```bash
   go test ./...
   ```
5. **Open a pull request** against `master` with a clear description of what you changed and why.

Please open an issue first for any significant change so we can discuss the approach before you invest time in implementation.
