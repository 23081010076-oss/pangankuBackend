# Setup WhatsApp Cloud API (PanganKu)

Panduan ini untuk mengaktifkan input data dari WhatsApp ke backend PanganKu.

## 1. Siapkan Meta Developer

1. Buka Meta for Developers dan buat App.
2. Tambahkan produk WhatsApp.
3. Ambil nilai berikut dari dashboard WhatsApp API:
   - Access Token
   - Phone Number ID
   - App Secret (di Basic Settings)

## 2. Isi Environment Variable

Tambahkan ke file .env backend:

WHATSAPP_VERIFY_TOKEN=token_verifikasi_webhook_anda
WHATSAPP_APP_SECRET=app_secret_meta_anda
WHATSAPP_PHONE_NUMBER_ID=phone_number_id_anda
WHATSAPP_ACCESS_TOKEN=access_token_anda

Catatan:
- WHATSAPP_VERIFY_TOKEN Anda tentukan sendiri, harus sama dengan konfigurasi webhook di Meta.
- Untuk production gunakan permanent access token, bukan token sementara.

## 3. Jalankan Backend

Jalankan API:

go run cmd/server/main.go

Endpoint yang dipakai:
- GET /api/v1/whatsapp/webhook (verifikasi)
- POST /api/v1/whatsapp/webhook (event incoming message)
- GET /api/v1/whatsapp/help (cek format dan readiness)

## 4. Expose URL ke Internet

Karena webhook Meta harus URL publik, gunakan tunnel saat development.
Contoh dengan ngrok:

ngrok http 8080

Lalu gunakan URL hasil tunnel, misalnya:
https://abc123.ngrok-free.app/api/v1/whatsapp/webhook

## 5. Konfigurasi Webhook di Meta

1. Callback URL: URL publik + /api/v1/whatsapp/webhook
2. Verify Token: isi sama dengan WHATSAPP_VERIFY_TOKEN
3. Subscribe field messages

## 6. Format Pesan dari Petani/Pedagang

Input laporan darurat:
LAPOR#kecamatan_id#jenis_masalah#deskripsi#prioritas

Input harga komoditas:
HARGA#komoditas_id#kecamatan_id#harga_per_kg#YYYY-MM-DD

Contoh:
- LAPOR#d8f2b4a9-7abc-4c07-9d27-6a1a78f77d55#Kekurangan Beras#Stok menipis 2 hari terakhir#4
- HARGA#8c154040-3bb2-4971-ac9c-1b985e7d6d4f#d8f2b4a9-7abc-4c07-9d27-6a1a78f77d55#13500#2026-03-17

## 7. Syarat Data User

Nomor pengirim WhatsApp akan dipetakan ke field phone user.
Pastikan:
- User aktif (is_active=true)
- Format nomor konsisten, disarankan 62xxxxxxxxxx

## 8. Uji Cepat

1. Cek readiness:
   GET /api/v1/whatsapp/help
2. Kirim pesan WA sesuai format.
3. Verifikasi data masuk ke tabel laporan_darurat atau harga_pasars.

## 9. Troubleshooting

- Error verify token: token di Meta tidak sama dengan WHATSAPP_VERIFY_TOKEN.
- Webhook 401 signature: cek WHATSAPP_APP_SECRET.
- Nomor tidak dikenali: update field phone user agar sesuai nomor pengirim.
- Tidak ada balasan WA: cek WHATSAPP_ACCESS_TOKEN dan WHATSAPP_PHONE_NUMBER_ID.
