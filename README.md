# Go Gallery API

A RESTful image gallery backend built with Go, PostgreSQL, and Cloudflare R2.

## Features

- **Image gallery** with public listing and viewing
- **Admin authentication** via JWT (access + refresh tokens)
- **Image upload** with automatic resizing to main (2560px) and thumbnail (400x400)
- **Cloudflare R2** object storage for images
- **PostgreSQL** for metadata persistence
- **Display ordering** for images

## Tech Stack

- Go 1.26
- PostgreSQL
- Cloudflare R2 (S3-compatible)
- JWT (HS256)
- bcrypt password hashing
- AWS SDK v2 for R2

## Project Structure

```
.
├── main.go                  # Entry point, server setup, middleware
├── internal/
│   ├── auth/                # JWT token generation & validation
│   ├── database/            # Database models & queries (sqlc-generated)
│   ├── httpapi/
│   │   ├── admin.go        # Admin handlers (login, image CRUD)
│   │   ├── middleware.go   # JWT auth middleware
│   │   └── public.go       # Public handlers (list/view images)
│   └── storage/
│       └── r2.go           # R2 client, image resizing & upload
├── sql/
│   ├── schema/              # Database migrations (Goose)
│   └── queries/            # SQL queries for sqlc
├── example.env             # Environment variable template
└── sqlc.yml                # sqlc configuration
```

## Setup

### 1. Configure environment

Copy `example.env` to `.env` and fill in your values:

```bash
cp example.env .env
```

Required variables:

| Variable | Description |
|---|---|
| `DB_URL` | PostgreSQL connection string |
| `TOKEN_SECRET` | JWT signing secret (min 32 chars) |
| `ADMIN_EMAIL` | Initial admin account email |
| `ADMIN_PASSWORD` | Initial admin account password |
| `CF_R2_ACCOUNT_ID` | Cloudflare R2 account ID |
| `CF_R2_ACCESS_KEY_ID` | R2 access key ID |
| `CF_R2_SECRET_ACCESS_KEY` | R2 secret access key |
| `CF_R2_BUCKET_NAME` | R2 bucket name |
| `CF_R2_PUBLIC_URL` | Public URL prefix for serving images |

### 2. Run database migrations

```bash
# Install goose if needed
go install github.com/pressly/goose/v3/cmd/goose@latest

# Run migrations
goose -dir sql/schema postgres "$DB_URL" up
```

### 3. Generate database code

```bash
sqlc generate
```

### 4. Run the server

```bash
go run main.go
```

The server starts on port 8080 by default.

## API Endpoints

### Public

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/images` | List all images |
| `GET` | `/api/images?id=<id>` | Get a single image by ID |

### Admin (requires `Authorization: Bearer <access_token>`)

| Method | Path | Description |
|---|---|---|
| `POST` | `/admin/login` | Admin login, returns access + refresh tokens |
| `POST` | `/admin/images` | Upload a new image (multipart, field: `file`) |
| `GET` | `/admin/images` | List all images with metadata |
| `PUT` | `/admin/images` | Update image metadata |
| `DELETE` | `/admin/images?id=<id>` | Delete an image |
| `PUT` | `/admin/images/reorder` | Reorder image display order |
| `POST` | `/admin/refresh` | Refresh access token |

### Request/Response Examples

**Admin login**

```bash
curl -X POST http://localhost:8080/admin/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@gallery.local","password":"your-password"}'
```

**Upload image**

```bash
curl -X POST http://localhost:8080/admin/images \
  -H "Authorization: Bearer <access_token>" \
  -F "file=@/path/to/image.jpg" \
  -F "name=My Image" \
  -F "description=Optional description"
```

**List public images**

```bash
curl http://localhost:8080/api/images
```

## Database Schema

### `users`

| Column | Type | Notes |
|---|---|---|
| `id` | UUID | Primary key |
| `email` | TEXT | Unique |
| `password_hash` | TEXT | bcrypt hash |
| `created_at` | TIMESTAMP | |
| `updated_at` | TIMESTAMP | |

### `images`

| Column | Type | Notes |
|---|---|---|
| `id` | UUID | Primary key |
| `image_name` | TEXT | |
| `image_url` | TEXT | Full-size R2 key |
| `thumbnail_url` | TEXT | Thumbnail R2 key |
| `user_id` | UUID | FK → users |
| `image_description` | TEXT | Nullable |
| `display_order` | INT | For manual ordering |
| `created_at` | TIMESTAMP | |
| `updated_at` | TIMESTAMP | |

## Token Notes

- Access tokens expire after `TOKEN_DEFAULT_DURATION` (default 24h)
- Refresh tokens expire after `REFRESH_TOKEN_TIMEOUT` (default 7 days)
- Refresh tokens are single-use — each refresh issues a new one

## License

MIT
