# Go Gallery API

A RESTful image gallery backend built with Go, PostgreSQL, and Cloudflare R2.

## Motivation

This was created to give a simple implementation of a photo gallery that can be easily deployed either raw or containerized, making use of Cloudflare database and storage capabilities to offload those requirements to a managed and trusted cloud provider. Ultimately, I was after a simple to manage and deploy system that can be run on low-cost hardware without the need to purchase or manage the storage for the database and assets. 

## Quick Start

### 0. Prerequisites

To run this you will need a Cloudflare account (the free tier has all you need to get going). You will need to configure the following:

- R2 Bucket, where the image assets will be stored.
- PostgreSQL Database, where the site config and metadata is stored.

Make sure that you have configured access tokens for these, as they will be required for getting the gallery up and running.

If you plan on taking this into production, I suggest you have a domain that is configured in Cloudflare and use the Cloudflare connector setup to handle presenting it to the web, as well as using their Zero Trust tools to manage access to the administration pages. I also suggest that you set the R2 and database assets to use your domain as opposed to the default settings (mainly because it looks neater to people who inspect your site's HTML).

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

## Usage

### Features

- **Image gallery** with public listing and viewing
- **Admin authentication** via JWT (access + refresh tokens)
- **Image upload** with automatic resizing to main (2560px) and thumbnail (400x400)
- **Cloudflare R2** object storage for images
- **PostgreSQL** for metadata persistence
- **Display ordering** for images

### API Endpoints

#### Public

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/images` | List all images |
| `GET` | `/api/images?id=<id>` | Get a single image by ID |

#### Admin (requires `Authorization` header)

| Method | Path | Description |
|---|---|---|
| `POST` | `/admin/login` | Admin login, returns access + refresh tokens |
| `POST` | `/admin/images` | Upload a new image (multipart, field: `file`) |
| `GET` | `/admin/images` | List all images with metadata |
| `PUT` | `/admin/images` | Update image metadata |
| `DELETE` | `/admin/images?id=<id>` | Delete an image |
| `PUT` | `/admin/images/reorder` | Reorder image display order |
| `POST` | `/admin/refresh` | Refresh access token |

#### Request Examples

**Admin login**
```bash
curl -X POST http://localhost:8080/admin/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@gallery.local","password":"your-password"}'
```

**Upload image**
```bash
curl -X POST http://localhost:8080/admin/images \
  -H "Authorization: Bearer <token>" \
  -F "file=@/path/to/image.jpg" \
  -F "name=My Image" \
  -F "description=Optional description"
```

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository.
2. Create a new feature branch (`git checkout -b feature/AmazingFeature`).
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`).
4. Push to the branch (`git push origin feature/AmazingFeature`).
5. Open a Pull Request.

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

## Database Schema

### `users`
- `id`: UUID (PK)
- `email`: TEXT (Unique)
- `password_hash`: TEXT (bcrypt)

### `images`
- `id`: UUID (PK)
- `image_name`: TEXT
- `image_url`: TEXT (Main R2 key)
- `thumbnail_url`: TEXT (Thumbnail R2 key)
- `user_id`: UUID (FK)
- `display_order`: INT

## License

MIT
