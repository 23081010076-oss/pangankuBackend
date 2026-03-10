# 🎉 BACKEND & DATABASE SIAP DIGUNAKAN!

## ✅ Status Services

Semua services sudah running:

```
✓ Backend API:  http://localhost:8080 (HEALTHY)
✓ PostgreSQL:   localhost:5433
✓ Redis Cache:  localhost:6380
```

---

## 👤 AKUN ADMIN DEFAULT

Login dengan kredensial berikut:

```
📧 Email:     admin@panganku.id
🔑 Password:  Admin123!
👔 Role:      admin
```

### Cara Login:

**Option 1: Via Flutter App (Recommended)**

1. Jalankan Flutter app
2. Masukkan email dan password di atas
3. Klik "Masuk"

**Option 2: Via API/Postman**

```bash
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json

{
  "email": "admin@panganku.id",
  "password": "Admin123!"
}
```

Response sukses:

```json
{
  "access_token": "eyJhbGciOiJIUzI1Ni...",
  "refresh_token": "random-token-here",
  "user": {
    "id": "uuid",
    "name": "Administrator",
    "email": "admin@panganku.id",
    "role": "admin"
  }
}
```

---

## 🗄️ AKSES DATABASE

### PostgreSQL Connection

```
Host:     localhost
Port:     5433
Database: panganku_db
Username: panganku
Password: secretPanganKu2026!
```

### Cara Akses:

**1. Via psql (Command Line)**

```bash
# Dari Docker container
docker exec -it panganku_backend-postgres-1 psql -U panganku -d panganku_db

# Dari host (jika psql installed)
psql -h localhost -p 5433 -U panganku -d panganku_db
```

**2. Via DBeaver / pgAdmin / DataGrip**

- Buat New Connection → PostgreSQL
- Input kredensial di atas
- Test & Save

**3. Via VS Code Extension**
Install extension **PostgreSQL** by Chris Kolkman, lalu tambah koneksi dengan kredensial di atas.

### Query Cepat:

```sql
-- Lihat semua tabel
\dt

-- Lihat user yang ada
SELECT id, name, email, role FROM users;

-- Lihat komoditas yang di-seed
SELECT * FROM komoditas;

-- Lihat kecamatan
SELECT * FROM kecamatan ORDER BY nama;
```

---

## 📊 DATA YANG SUDAH TERSEDIA

### ✅ Users

- **1 Admin**: admin@panganku.id (Password: Admin123!)

### ✅ Komoditas (9 items)

- Beras
- Jagung
- Kedelai
- Cabai Merah
- Bawang Merah
- Gula Pasir
- Minyak Goreng
- Daging Ayam
- Telur Ayam

### ✅ Kecamatan Lamongan (10 items)

- Babat, Lamongan, Sekaran, Deket, Tikung
- Sarirejo, Pucuk, Karanggeneng, Kedungpring, Paciran

---

## 🚀 TESTING API

### Health Check

```bash
GET http://localhost:8080/health
```

### Login Admin

```bash
POST http://localhost:8080/api/v1/auth/login
Body: {"email":"admin@panganku.id","password":"Admin123!"}
```

### Get Current User Info

```bash
GET http://localhost:8080/api/v1/auth/me
Headers: Authorization: Bearer <your-access-token>
```

### List Komoditas

```bash
GET http://localhost:8080/api/v1/komoditas
Headers: Authorization: Bearer <your-access-token>
```

---

## 🔧 REDIS CACHE

```
Host: localhost
Port: 6380
```

Akses Redis CLI:

```bash
docker exec -it panganku_backend-redis-1 redis-cli

# Lihat cache yang ada
KEYS *

# Get specific cache
GET harga:latest
```

---

## 📱 JALANKAN FLUTTER APP

```bash
cd d:\ketahanPanganMobile\panganku_mobile

# Via Chrome (Web)
flutter run -d chrome

# Via Android Emulator (jika ada)
flutter run -d emulator-5554

# Via Windows Desktop
flutter run -d windows
```

Login di app dengan:

- Email: `admin@panganku.id`
- Password: `Admin123!`

---

## 🔄 RESTART / STOP SERVICES

### Restart

```bash
cd d:\ketahanPanganMobile\panganku_backend
docker-compose restart
```

### Stop

```bash
docker-compose down
```

### Start Again

```bash
docker-compose up -d
```

### View Logs

```bash
docker-compose logs -f api
docker-compose logs -f postgres
```

---

## 🎯 NEXT STEPS

1. ✅ Backend Running
2. ✅ Database Ready
3. ✅ Admin Created
4. ⏭️ Jalankan Flutter App
5. ⏭️ Login sebagai Admin
6. ⏭️ Mulai input data harga komoditas
7. ⏭️ Buat user petugas per kecamatan

---

## 📚 DOKUMENTASI LENGKAP

Lihat file berikut untuk detail lebih lanjut:

- `DATABASE_ACCESS.md` - Panduan lengkap akses database
- `README.md` - Dokumentasi backend
- `panganku_mobile/README.md` - Dokumentasi Flutter

---

## ⚠️ TROUBLESHOOTING

### Backend tidak bisa akses database?

```bash
docker-compose logs api
docker-compose ps
```

### Lupa password admin?

Login ke PostgreSQL dan query:

```sql
SELECT email, role FROM users WHERE role = 'admin';
```

Password default: `Admin123!`

### Port sudah terpakai?

Ubah port di `docker-compose.yml`:

- PostgreSQL: `5433:5432` (host:container)
- Redis: `6380:6379`
- API: `8080:8080`

---

**🎉 Selamat! Backend PanganKu sudah siap digunakan!**
