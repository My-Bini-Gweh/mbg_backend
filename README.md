# ITSPay Backend

Backend demo untuk Final Project Manajemen Basis Data. Fokus utama project ini adalah menunjukkan proses finansial yang dikendalikan database melalui stored procedure, trigger, function, view, index, dan transaction control.

## Fitur Utama

- REST API dengan Go, Gin, MySQL, dan JWT.
- Register mahasiswa otomatis membuat wallet lewat trigger.
- Top up memanggil `sp_topup_wallet`.
- Pembayaran merchant memanggil `sp_bayar_merchant`.
- Validasi saldo tidak negatif di trigger database.
- Audit log otomatis setelah transaksi dibuat.
- Riwayat mahasiswa dari `v_riwayat_transaksi_mahasiswa`.
- Laporan admin dari `v_laporan_transaksi_harian`.

## Setup

Masuk ke folder backend:

```bash
cd mbg_backend
```

Install dependency:

```bash
go mod tidy
```

Buat database:

```sql
CREATE DATABASE itspay CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

Copy environment:

```bash
cp .env.example .env
```

Isi `.env` sesuai MySQL lokal:

```env
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=
DB_NAME=itspay
JWT_SECRET=change-this-secret
APP_PORT=8080
CORS_ALLOWED_ORIGIN=http://localhost:3000
```

Jalankan migration berurutan memakai MySQL CLI:

```bash
mysql -u root -p itspay < migrations/001_create_tables.sql
mysql -u root -p itspay < migrations/002_create_function.sql
mysql -u root -p itspay < migrations/003_create_procedures.sql
mysql -u root -p itspay < migrations/004_create_triggers.sql
mysql -u root -p itspay < migrations/005_create_views.sql
mysql -u root -p itspay < migrations/006_create_indexes.sql
mysql -u root -p itspay < migrations/007_seed_demo_data.sql
```

Jalankan server:

```bash
go run main.go
```

Server berjalan di:

```text
http://localhost:8080
```

## Akun Demo

Mahasiswa:

```text
email: alya@itspay.test
password: password
wallet: WLT-0001
```

Admin:

```text
email: admin@itspay.test
password: password
```

## Demo Flow

Login mahasiswa:

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alya@itspay.test","password":"password"}'
```

Top up saldo:

```bash
curl -X POST http://localhost:8080/api/topups \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN_MAHASISWA>" \
  -d '{"wallet_id":"WLT-0001","bank_id":1,"nominal":100000}'
```

Bayar merchant:

```bash
curl -X POST http://localhost:8080/api/payments \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN_MAHASISWA>" \
  -d '{"wallet_id":"WLT-0001","merchant_id":"MRC-0001","nominal":15000}'
```

Cek riwayat mahasiswa:

```bash
curl http://localhost:8080/api/mahasiswa/transactions \
  -H "Authorization: Bearer <TOKEN_MAHASISWA>"
```

Login admin lalu cek audit log dan laporan harian:

```bash
curl http://localhost:8080/api/admin/audit-logs \
  -H "Authorization: Bearer <TOKEN_ADMIN>"

curl http://localhost:8080/api/admin/reports/daily \
  -H "Authorization: Bearer <TOKEN_ADMIN>"
```

## Endpoint

Authentication:

- `POST /api/auth/register`
- `POST /api/auth/login`
- `GET /api/auth/me`

Mahasiswa:

- `GET /api/mahasiswa/profile`
- `GET /api/mahasiswa/wallet`
- `GET /api/mahasiswa/transactions`

Financial:

- `POST /api/topups`
- `POST /api/payments`

Catalog:

- `GET /api/banks`
- `GET /api/merchants`
- `GET /api/merchants/:id`

Admin:

- `GET /api/admin/transactions`
- `GET /api/admin/audit-logs`
- `GET /api/admin/reports/daily`

## Catatan Presentasi

Bagian yang perlu ditonjolkan saat demo:

- Backend tidak mengubah saldo manual untuk top up/payment.
- Proses finansial terjadi di `sp_topup_wallet` dan `sp_bayar_merchant`.
- `sp_bayar_merchant` memakai transaksi database dan `SELECT ... FOR UPDATE`.
- Saldo tidak bisa negatif karena validasi database.
- Setiap transaksi otomatis masuk `audit_logs` lewat trigger.
- Report harian dan riwayat mahasiswa dibaca dari view.
