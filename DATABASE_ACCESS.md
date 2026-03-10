# Informasi Database & Akun PanganKu

## 🗄️ Akses Database PostgreSQL

Database sudah running di Docker. Berikut cara aksesnya:

### Koneksi Database

```bash
Host: localhost
Port: 5433
Database: panganku_db
Username: panganku
Password: secretPanganKu2026!
```

### Akses via Command Line

```bash
# Masuk ke container PostgreSQL
docker exec -it panganku_backend-postgres-1 psql -U panganku -d panganku_db

# Atau langsung dari host (jika psql terinstall)
psql -h localhost -p 5433 -U panganku -d panganku_db
```

### Akses via GUI (DBeaver, pgAdmin, DataGrip)

1. Buat koneksi baru
2. Pilih PostgreSQL
3. Masukkan kredensial di atas
4. Test connection dan Save

### Query Contoh

```sql
-- Lihat semua tabel
\dt

-- Lihat semua user
SELECT id, name, email, role, is_active FROM users;

-- Lihat semua komoditas
SELECT * FROM komoditas ORDER BY nama;

-- Lihat data harga terbaru
SELECT h.*, k.nama as komoditas, kec.nama as kecamatan
FROM harga_pasar h
JOIN komoditas k ON k.id = h.komoditas_id
JOIN kecamatan kec ON kec.id = h.kecamatan_id
ORDER BY h.tanggal DESC
LIMIT 10;
```

## 👤 Akun Admin Default

Backend sudah otomatis membuat akun admin saat pertama kali dijalankan:

### Kredensial Login

```
Email: admin@panganku.id
Password: Admin123!
Role: admin
```

### Cara Login

#### Via API (Postman/cURL)

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@panganku.id",
    "password": "Admin123!"
  }'
```

Response:

```json
{
  "access_token": "eyJhbGciOiJIUzI1Ni...",
  "refresh_token": "dGhpcyBpcyByYW5kb20...",
  "user": {
    "id": "uuid-here",
    "name": "Administrator",
    "email": "admin@panganku.id",
    "role": "admin"
  }
}
```

#### Via Flutter App

1. Jalankan Flutter app
2. Di halaman login, masukkan:
   - Email: `admin@panganku.id`
   - Password: `Admin123!`
3. Klik "Masuk"

## 🔧 Redis

Redis juga sudah running untuk cache dan session:

```bash
Host: localhost
Port: 6380
Password: (none)
```

Akses Redis CLI:

```bash
docker exec -it panganku_backend-redis-1 redis-cli

# Lihat semua keys
KEYS *

# Lihat data cache harga
GET harga:latest

# Monitor real-time commands
MONITOR
```

## 📊 Data yang Sudah Di-seed

### Komoditas (9 items)

- Beras
- Jagung
- Kedelai
- Cabai Merah
- Bawang Merah
- Gula Pasir
- Minyak Goreng
- Daging Ayam
- Telur Ayam

### Kecamatan di Lamongan (10 items)

- Babat
- Lamongan
- Sekaran
- Deket
- Tikung
- Sarirejo
- Pucuk
- Karanggeneng
- Kedungpring
- Paciran

## 🚀 Endpoint API Tersedia

Base URL: `http://localhost:8080/api/v1`

### Auth Endpoints

- `POST /auth/register` - Registrasi user baru
- `POST /auth/login` - Login dan dapatkan token
- `POST /auth/logout` - Logout (butuh token)
- `POST /auth/refresh` - Refresh access token
- `GET /auth/me` - Get current user info (butuh token)

### Harga Endpoints (butuh token)

- `GET /harga` - List harga dengan filter
- `GET /harga/latest` - Harga terbaru per komoditas
- `GET /harga/trend/:komoditas_id` - Trend harga 30 hari
- `GET /harga/forecast?komoditas_id=xxx` - Prediksi harga 7 hari
- `POST /harga` - Tambah data harga (admin/petugas only)

### Upload Endpoint

- `POST /upload/foto` - Upload foto laporan (butuh token)

## 🔐 Testing dengan Postman

### 1. Login

```
POST http://localhost:8080/api/v1/auth/login
Body (JSON):
{
  "email": "admin@panganku.id",
  "password": "Admin123!"
}
```

### 2. Simpan Access Token

Copy `access_token` dari response

### 3. Test Endpoint Lain

```
GET http://localhost:8080/api/v1/auth/me
Headers:
Authorization: Bearer <your-access-token>
```

## 🔄 Restart Services

Jika perlu restart:

```bash
cd panganku_backend

# Stop semua container
docker-compose down

# Start lagi
docker-compose up -d

# Lihat logs real-time
docker-compose logs -f api
```

## 📝 Buat User Baru

Setelah login sebagai admin, buat user baru via API:

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Petugas Babat",
    "email": "petugas.babat@panganku.id",
    "password": "Petugas123!",
    "phone": "081234567891"
  }'
```

Role default: `publik`. Untuk ubah role jadi `petugas` atau `admin`, update langsung di database.

## ⚠️ Troubleshooting

### Backend tidak bisa akses database

```bash
# Cek status container
docker-compose ps

# Cek logs jika ada error
docker-compose logs api
docker-compose logs postgres
```

### Lupa password admin

```bash
# Reset password via SQL
docker exec -it panganku_backend-postgres-1 psql -U panganku -d panganku_db

UPDATE users SET password = '<hash-password-baru>' WHERE email = 'admin@panganku.id';
```

Atau rebuild backend agar seed ulang (hapus dulu user admin).
