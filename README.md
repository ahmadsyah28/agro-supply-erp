# PT AgroSupply ERP — Backend API

Backend REST API untuk ERP distributor bahan baku PT AgroSupply, dibangun menggunakan **Go (Golang)** murni dengan arsitektur **Modular Monolith**, database **PostgreSQL (Neon)** via `pgxpool`, dan migrasi SQL terintegrasi.

Dokumen arsitektur lengkap: [ARCHITECTURE.md](ARCHITECTURE.md).

---

## 🚀 Fitur Utama & Modul

1. **Master Data Products:** Manajemen katalog produk bahan baku (SKU unik, nama, satuan unit, status aktif).
2. **Master Data Suppliers:** Manajemen data rekanan pemasok (nama PT/CV, kontak PIC, email, telepon, alamat).
3. **Procurement (PR & PO):**
   - **Purchase Request (PR):** Pengajuan kebutuhan barang internal dengan workflow `DRAFT` $\rightarrow$ `SUBMITTED` $\rightarrow$ `APPROVED` / `REJECTED`.
   - **Purchase Order (PO):** Penerbitan pesanan resmi ke supplier aktif dengan proteksi business rules (`DRAFT` $\rightarrow$ `APPROVED` $\rightarrow$ `PARTIALLY_RECEIVED` $\rightarrow$ `COMPLETED` / `CANCELLED`).
4. **Inventory & Stock Ledger:**
   - **Goods Receipt (Penerimaan Barang):** Penerimaan fisik dari surat jalan supplier dengan pemisahan hasil Quality Control (`accepted_quantity` vs `rejected_quantity`).
   - **Penerimaan Parsial:** Rekonsiliasi otomatis status PO antar pengiriman truk.
   - **Audit-Proof Stock Ledger:** Saldo stok dihitung dari akumulasi mutasi masuk (`IN`) dan keluar (`OUT`) tanpa manipulasi angka manual.
5. **Stabilization & Reliability:**
   - Panic Recovery middleware (mencegah server crash akibat unhandled error).
   - Sanitasi logging (tidak pernah mencatat password, token, atau secret ke log).
   - Standardisasi format respon JSON & pagination metadata.
   - CORS enabled untuk integrasi frontend.

---

## 🛠️ Persyaratan Sistem

- **Go:** Versi 1.22 ke atas (teruji pada Go 1.27)
- **Database:** PostgreSQL (Neon PostgreSQL didukung penuh dengan `sslmode=require`)
- **Git**

---

## ⚙️ Panduan Menjalankan Aplikasi

### 1. Konfigurasi Lingkungan (`.env`)
Salin template konfigurasi dan sesuaikan string koneksi database Neon Anda:

```bash
cp .env.example .env
```

Isi file `.env`:
```dotenv
APP_ENV=development
HTTP_ADDR=:8080
DATABASE_URL=postgresql://USER:PASSWORD@HOST/DATABASE?sslmode=require
```

### 2. Menjalankan Server
```bash
go run ./cmd/api
```
*Catatan:* Server otomatis menjalankan file migrasi database SQL yang belum diaplikasikan saat pertama kali terhubung ke database.

### 3. Menjalankan Pengujian (Unit & E2E Tests)
```bash
go test -v ./...
```

---

## 📚 Dokumentasi Endpoint REST API

Semua request dan response menggunakan format baku `application/json; charset=utf-8`.

### Struktur Respon Standar

**Respon Berhasil:**
```json
{
  "status": "success",
  "data": { ... },
  "meta": {
    "page": 1,
    "limit": 20,
    "total_items": 100,
    "total_pages": 5
  }
}
```

**Respon Gagal / Validasi:**
```json
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Pesan galat yang informatif",
    "details": null
  }
}
```

---

### Daftar Endpoint Lengkap

#### 1. Health Checks
- `GET /health/live` — Liveness probe (mengecek apakah proses web server berjalan).
- `GET /health/ready` — Readiness probe (mengecek kesiapan koneksi ke database PostgreSQL).

#### 2. Master Data Products
- `POST /api/v1/products` — Tambah produk baru
  ```json
  { "sku": "PROD-BATOK-001", "name": "Batok Kelapa Grade A", "unit": "kg" }
  ```
- `GET /api/v1/products` — Ambil daftar produk
- `GET /api/v1/products/{id}` — Detail produk
- `PUT /api/v1/products/{id}` — Update produk (`name`, `unit`, `is_active`)
- `DELETE /api/v1/products/{id}` — Nonaktifkan produk (*soft delete*)

#### 3. Master Data Suppliers
- `POST /api/v1/suppliers` — Tambah supplier baru
  ```json
  {
    "name": "CV Agro Lestari",
    "contact_name": "Budi",
    "email": "budi@agrolestari.com",
    "phone": "08123456789",
    "address": "Lampung"
  }
  ```
- `GET /api/v1/suppliers` — Ambil daftar supplier
- `GET /api/v1/suppliers?active=true` — Ambil hanya supplier yang berstatus aktif
- `GET /api/v1/suppliers/{id}` — Detail supplier
- `PUT /api/v1/suppliers/{id}` — Update supplier
- `DELETE /api/v1/suppliers/{id}` — Nonaktifkan supplier

#### 4. Procurement: Purchase Request (PR)
- `POST /api/v1/purchase-requests` — Buat PR baru
  ```json
  {
    "notes": "Kebutuhan bahan baku produksi arang",
    "items": [
      { "product_id": 1, "quantity": 5000, "notes": "Kadar air maks 12%" }
    ]
  }
  ```
- `GET /api/v1/purchase-requests` — Daftar PR
- `GET /api/v1/purchase-requests/{id}` — Detail PR beserta item-itemnya
- `POST /api/v1/purchase-requests/{id}/submit` — Ajukan PR (`DRAFT` $\rightarrow$ `SUBMITTED`)
- `POST /api/v1/purchase-requests/{id}/approve` — Setujui PR (`SUBMITTED` $\rightarrow$ `APPROVED`)
- `POST /api/v1/purchase-requests/{id}/reject` — Tolak PR (`SUBMITTED` $\rightarrow$ `REJECTED`)

#### 5. Procurement: Purchase Order (PO)
- `POST /api/v1/purchase-orders` — Buat PO baru (Hanya ke supplier aktif!)
  ```json
  {
    "purchase_request_id": 1,
    "supplier_id": 1,
    "notes": "Pengiriman bertahap 2 truk",
    "items": [
      { "product_id": 1, "quantity": 5000, "unit_price": 2500 }
    ]
  }
  ```
- `GET /api/v1/purchase-orders` — Daftar PO
- `GET /api/v1/purchase-orders/{id}` — Detail PO, nama supplier, dan rincian harga
- `POST /api/v1/purchase-orders/{id}/approve` — Setujui PO untuk dikirim ke supplier (`APPROVED`)
- `POST /api/v1/purchase-orders/{id}/cancel` — Batalkan PO (`CANCELLED`)

#### 6. Inventory & Stock Ledger
- `POST /api/v1/goods-receipts` — Catat penerimaan barang fisik & hasil QC
  ```json
  {
    "purchase_order_id": 1,
    "delivery_order_number": "SJ-001/ALM/2026",
    "notes": "Penerimaan Truk 1",
    "items": [
      {
        "purchase_order_item_id": 1,
        "product_id": 1,
        "received_quantity": 3000,
        "accepted_quantity": 2900,
        "rejected_quantity": 100,
        "rejection_reason": "100 kg basah kehujanan"
      }
    ]
  }
  ```
- `GET /api/v1/goods-receipts` — Daftar seluruh riwayat Goods Receipt
- `GET /api/v1/goods-receipts/{id}` — Detail Goods Receipt & inspeksi QC
- `GET /api/v1/inventory/stock` — Saldo stok terkini setiap produk (`total_in`, `total_out`, `current_stock`)
- `GET /api/v1/inventory/ledger` — Riwayat buku besar mutasi stok (opsional filter `?product_id=1`)

---

## 🏛️ Struktur Direktori Proyek

```text
agro-supply-erp/
├── cmd/
│   └── api/
│       └── main.go                 # Entrypoint server & router wiring
├── internal/
│   ├── config/                     # Pengelola environment variables (.env)
│   ├── database/                   # pgxpool PostgreSQL & auto-migration runner
│   ├── health/                     # Liveness & Readiness handlers
│   ├── inventory/                  # Modul Goods Receipt & Inventory Ledger
│   ├── middleware/                 # CORS, Sanitized Logger, & Panic Recovery
│   ├── procurement/                # Modul Purchase Request & Purchase Order
│   ├── product/                    # Modul Master Data Product
│   ├── response/                   # Standard JSON & Pagination Formatter
│   └── supplier/                   # Modul Master Data Supplier
├── migrations/                     # File migrasi database SQL terurut
│   ├── 000001_create_products_table.up.sql
│   ├── 000002_create_suppliers_table.up.sql
│   ├── 000003_create_procurement_tables.up.sql
│   └── 000004_create_inventory_tables.up.sql
├── tests/
│   └── e2e_workflow_test.go        # Pengujian integrasi bisnis hulu-ke-hilir
├── .env.example
├── .gitignore
├── ARCHITECTURE.md
├── go.mod
├── go.sum
└── README.md
```
