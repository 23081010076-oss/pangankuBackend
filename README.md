# PanganKu Backend

Sistem Informasi Ketahanan Pangan untuk Diskominfo Kabupaten Lamongan.

## Tech Stack

- **Backend**: Golang 1.21, Gin, GORM, MySQL 8, Redis
- **Security**: JWT, Argon2id, AES-256-GCM
- **Algorithms**: Dynamic Programming (price prediction), Dijkstra (distribution)

## Quick Start

### Prerequisites

- Go 1.21+
- MySQL 8+
- Redis 7+

### Setup

1. Clone repository:

```bash
git clone <repository-url>
cd panganku_backend
```

2. Copy dan edit environment variables:

```bash
cp .env.example .env
# Edit .env dengan konfigurasi Anda
```

3. Install dependencies:

```bash
go mod download
```

4. Jalankan migrasi database:

```bash
go run cmd/server/main.go
# Migrasi otomatis berjalan saat startup
```

5. Run server:

```bash
go run cmd/server/main.go
```

Server berjalan di `http://localhost:8080`

## Docker

Build dan jalankan dengan Docker Compose:

```bash
docker-compose up -d
```

Ini akan menjalankan:

- API server di port 8080
- MySQL di port 3306
- Redis di port 6379

## API Endpoints

### Authentication

- `POST /api/v1/auth/register` - Daftar akun baru
- `POST /api/v1/auth/login` - Login
- `POST /api/v1/auth/logout` - Logout (requires JWT)
- `POST /api/v1/auth/refresh` - Refresh access token
- `GET /api/v1/auth/me` - Get user info (requires JWT)

### Harga Komoditas

- `GET /api/v1/harga` - List harga dengan filter & pagination
- `GET /api/v1/harga/latest` - Harga terkini semua komoditas
- `GET /api/v1/harga/trend/:komoditas_id` - Trend harga (7d/30d/90d)
- `GET /api/v1/harga/forecast` - Prediksi harga 7 hari ke depan
- `POST /api/v1/harga` - Tambah data harga (admin/petugas only)

### Upload

- `POST /api/v1/upload/foto` - Upload foto (JPG/PNG/WebP, max 5MB)

## Testing

Run unit tests:

```bash
go test ./... -v
```

With coverage:

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

Run benchmarks:

```bash
go test ./... -bench=. -benchmem
```

## Security Features

- ✅ Password hashing dengan Argon2id
- ✅ JWT dengan refresh token (Redis)
- ✅ Rate limiting per IP
- ✅ AES-256-GCM encryption untuk data sensitif
- ✅ Input validation & sanitization
- ✅ File upload validation (magic bytes)
- ✅ RBAC (Role-Based Access Control)
- ✅ Security headers (HSTS, X-Frame-Options, dll)
- ✅ Audit logging
- ✅ Brute force protection (max 5 attempts)

## Project Structure

```
panganku_backend/
├── cmd/
│   └── server/         # Entry point
├── internal/
│   ├── algorithms/     # Algoritma prediksi & distribusi
│   ├── config/         # Database & Redis config
│   ├── handlers/       # HTTP handlers
│   ├── middleware/     # Custom middleware
│   ├── models/         # GORM models
│   └── security/       # Security utilities
├── uploads/            # Uploaded files
├── .env.example
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── README.md
```

## Contributing

1. Fork repository
2. Buat branch feature (`git checkout -b feature/AmazingFeature`)
3. Commit changes (`git commit -m 'Add some AmazingFeature'`)
4. Push ke branch (`git push origin feature/AmazingFeature`)
5. Buat Pull Request

## License

Copyright © 2026 Diskominfo Kabupaten Lamongan
