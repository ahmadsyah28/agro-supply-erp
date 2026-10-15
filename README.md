# PT AgroSupply ERP

Dokumen arsitektur dan spesifikasi teknis lengkap mengacu pada [ARCHITECTURE.md](ARCHITECTURE.md).

## Persyaratan Sistem

- Go 1.22+ (teruji pada Go 1.27)
- PostgreSQL (Neon PostgreSQL didukung dengan `sslmode=require`)

## Struktur Proyek

```text
agro-supply-erp/
├── cmd/
│   └── api/
│       └── main.go           # Entrypoint HTTP server
├── internal/
│   ├── config/               # Konfigurasi runtime & pembacaan environment
│   ├── database/             # Koneksi pgxpool & migration runner
│   ├── health/               # Endpoint liveness & readiness (/health/live, /health/ready)
│   ├── middleware/           # Logging & CORS middleware
│   └── response/             # Standard JSON response & error formatting
├── migrations/               # Skrip migrasi SQL terurut
│   ├── 000001_create_products_table.up.sql
│   ├── 000001_create_products_table.down.sql
│   └── fs.go                 # Embedded migration filesystem
├── .env.example              # Template environment
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

## Konfigurasi

Salin file `.env.example` menjadi `.env` jika belum ada:

```bash
cp .env.example .env
```

Isi variabel `DATABASE_URL` dengan string koneksi Neon PostgreSQL Anda:

```dotenv
APP_ENV=development
HTTP_ADDR=:8080
DATABASE_URL=postgresql://USER:PASSWORD@HOST/DATABASE?sslmode=require
```

## Menjalankan Aplikasi

Jalankan server API:

```bash
go run ./cmd/api
```

Saat server berjalan dengan `DATABASE_URL` yang valid, migrasi database yang belum berjalan akan otomatis dieksekusi secara terurut dalam transaksi database atomik.

## Menguji Endpoint Health

1. **Liveness Check** (apakah proses aplikasi hidup):
   ```bash
   curl http://localhost:8080/health/live
   ```
   Response:
   ```json
   {
     "status": "success",
     "data": {
       "status": "alive",
       "timestamp": "..."
     }
   }
   ```

2. **Readiness Check** (apakah koneksi database PostgreSQL aktif):
   ```bash
   curl http://localhost:8080/health/ready
   ```

## Menjalankan Pengujian (Testing)

Jalankan unit tests:

```bash
go test -v ./...
```
