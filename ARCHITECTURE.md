# PT AgroSupply ERP — System Architecture

> Dokumen arsitektur teknis untuk MVP ERP PT AgroSupply.
>
> Status: **MVP Phase 1–6 Selesai & Operasional**  
> Backend: Go (Modular Monolith) + Neon PostgreSQL (`pgxpool`)  
> Frontend: Next.js 14 (App Router) + TypeScript + Custom Design System ([DESIGN.md](DESIGN.md))  
> Pengujian: End-to-End Workflow Test (`tests/`) & Bruno API Collection (`bruno/`)

## 1. Tujuan Sistem

PT AgroSupply adalah distributor bahan baku, misalnya batok kelapa, yang membeli barang dari supplier dan memasoknya kepada pelanggan.

MVP ini berfokus pada proses procurement dan inventory hulu-ke-hilir agar perusahaan dapat:
- Mengelola data master produk dan rekanan pemasok (supplier).
- Mencatat kebutuhan pengadaan melalui Purchase Request (PR) dengan workflow pengajuan & persetujuan.
- Menerbitkan dan menyetujui Purchase Order (PO) ke supplier aktif dengan proteksi business rules.
- Mencatat penerimaan fisik barang (Goods Receipt), termasuk inspeksi QC (diterima vs ditolak) dan penerimaan parsial (multi-truk).
- Melacak pergerakan dan saldo stok secara audit-proof melalui Inventory Ledger (buku besar mutasi stok).
- Mengoperasikan sistem melalui antarmuka web modern (Next.js) yang terhubung langsung ke REST API.

Fitur sales lengkap, akuntansi, pembayaran, multi-level approval kompleks, dan integrasi eksternal ditunda sampai kebutuhan inti berjalan stabil.

## 2. Prinsip Arsitektur

1. **Modular Monolith** — satu aplikasi backend Go dengan modul domain yang dipisahkan secara logis (`product`, `supplier`, `procurement`, `inventory`).
2. **Business rules di backend** — aturan bisnis, validasi kuantitas, dan perubahan status dijaga ketat di backend, bukan hanya di frontend.
3. **Database sebagai sumber data transaksi** — data penting disimpan secara persisten di Neon PostgreSQL dengan relasi foreign key dan constraint integritas.
4. **Audit-Proof Inventory ledger** — pergerakan stok dicatat sebagai mutasi masuk (`IN`) dan keluar (`OUT`), bukan sekadar mengubah angka kolom `stock` secara sembarangan.
5. **Explicit workflow** — PR, PO, dan Goods Receipt memiliki status transisi yang terdefinisi dan divalidasi (`DRAFT` $\rightarrow$ `SUBMITTED` $\rightarrow$ `APPROVED` $\rightarrow$ `PARTIALLY_RECEIVED` $\rightarrow$ `COMPLETED`).
6. **Atomic Database Transactions** — pencatatan Goods Receipt dan penambahan mutasi ledger dilakukan dalam satu transaksi database (`pgx.Tx`) untuk mencegah desinkronisasi.
7. **Production Stabilization & Security** — logging tersanitasi (tanpa membocorkan secret/password), panic recovery middleware untuk uptime, dan proteksi CORS untuk frontend.
8. **Build incrementally** — implementasi bertahap dari Foundation, Master Data, Procurement, Inventory, Stabilization, hingga Frontend.

## 3. Technology Stack

| Layer | Teknologi | Implementasi & Tanggung Jawab |
|---|---|---|
| Frontend | Next.js 14 + React 18 + TypeScript | UI Dashboard, Master Data, Form Transaksi PR/PO/GR, & visualisasi Ledger |
| Styling / Design System | Vanilla CSS (CSS Custom Properties) | Glassmorphism modern, responsif, tanpa dependensi framework CSS eksternal ([DESIGN.md](DESIGN.md)) |
| Backend | Go (Golang) 1.22+ | HTTP REST API, business rules, validasi, dan transaksi atomik |
| HTTP routing & Middleware | `net/http` standard library | Routing RESTful, CORS handler, Panic Recovery, dan Sanitized Logger |
| PostgreSQL driver | `github.com/jackc/pgx/v5/pgxpool` | Connection pool berkinerja tinggi, concurrency-safe, dan transaksional |
| Database | Neon PostgreSQL (Serverless) | Penyimpanan cloud PostgreSQL terkelola dengan koneksi terenkripsi (`sslmode=require`) |
| Schema migration | Built-in Custom Migration Runner | Eksekusi otomatis migrasi SQL terurut (`migrations/*.sql`) saat server boot |
| Testing & Verification | Go `testing` + Bruno Collection | Pengujian alur bisnis hulu-ke-hilir (`e2e_workflow_test.go`) & pengujian API manual |
| Version control | Git & GitHub | Pelacakan versi kode dan deployment |

## 4. High-Level Architecture

```text
User / Web Browser
       │
       ▼
Next.js Frontend (web: App Router, React 18, TypeScript)
       │
       │ HTTP / JSON (via web/lib/api.ts)
       ▼
Go REST API (cmd/api:8080)
       │
  [Middleware Pipeline: CORS, Logger Sanitizer, Panic Recovery]
       │
       ├──► HTTP Handlers (product, supplier, procurement, inventory, health)
       │       │
       │       ▼
       ├──► Application / Business Services (Rules, Validations, Status Transitions)
       │       │
       │       ▼
       └──► Repository / Data Access (pgxpool SQL Queries & Atomic Transactions)
               │
               ▼
       Neon PostgreSQL (Cloud Database)
```

Tanggung jawab setiap layer:
- **Frontend (Next.js)**: Menyajikan antarmuka visual interaktif, mengelola form input, menampilkan kartu KPI, filter status, dan memanggil REST API backend via `web/lib/api.ts`.
- **Middleware**: Menangani cross-origin resource sharing (CORS), menangkap dan memulihkan unhandled panics, serta mencatat request log tanpa mengekspos kredensial sensitif.
- **Handler**: Mem-parsing format JSON/query param, memvalidasi input dasar, memanggil service layer, dan mengembalikan respon JSON standar (`status: "success"` / `"error"`).
- **Service / Business Logic**: Menerapkan aturan bisnis inti (misal: PO hanya boleh ke supplier aktif, GR tidak boleh melebihi sisa PO, pemisahan accepted vs rejected QC).
- **Repository**: Melakukan operasi I/O ke database menggunakan prepared query dan transaksi `pgx.Tx`.
- **Neon PostgreSQL**: Menjaga integritas data melalui tabel relasional, foreign keys, indeks, dan constraint `CHECK`.

## 5. Struktur Direktori Proyek Aktual

```text
agro-supply-erp/
├── cmd/
│   └── api/
│       └── main.go                 # Entrypoint server, migrasi, dan router wiring
├── internal/
│   ├── config/                     # Konfigurasi aplikasi dari environment variable (.env)
│   ├── database/                   # Koneksi pgxpool dan built-in auto-migration runner
│   ├── health/                     # Handler liveness (/health/live) dan readiness (/health/ready)
│   ├── inventory/                  # Modul Goods Receipt, QC, dan Stock Ledger
│   │   ├── handler.go
│   │   ├── model.go
│   │   ├── repository.go
│   │   └── service.go
│   ├── middleware/                 # CORS, Sanitized Logger, dan Panic Recovery
│   ├── procurement/                # Modul Purchase Request & Purchase Order
│   │   ├── handler.go
│   │   ├── model.go
│   │   ├── repository.go
│   │   └── service.go
│   ├── product/                    # Modul Master Data Product
│   │   ├── handler.go
│   │   ├── model.go
│   │   ├── repository.go
│   │   └── service.go
│   ├── response/                   # Standarisasi JSON payload, error formatting, dan pagination
│   └── supplier/                   # Modul Master Data Supplier
│       ├── handler.go
│       ├── model.go
│       ├── repository.go
│       └── service.go
├── migrations/                     # File migrasi SQL terurut
│   ├── 000001_create_products_table.up.sql
│   ├── 000002_create_suppliers_table.up.sql
│   ├── 000003_create_procurement_tables.up.sql
│   ├── 000004_create_inventory_tables.up.sql
│   └── 000005_ensure_suppliers_columns.up.sql
├── tests/
│   └── e2e_workflow_test.go        # Pengujian integrasi bisnis hulu-ke-hilir
├── bruno/                          # Bruno API Testing Collection
├── web/                            # Frontend Next.js 14
│   ├── app/
│   │   ├── globals.css             # Desain sistem & styling kustom
│   │   ├── layout.tsx              # Root layout & font loader
│   │   └── page.tsx                # Halaman utama aplikasi (tab navigasi & view)
│   ├── lib/
│   │   └── api.ts                  # Client fetch wrapper untuk REST API Go
│   ├── package.json
│   └── tsconfig.json
├── .env.example
├── .gitignore
├── ARCHITECTURE.md
├── DESIGN.md                       # Dokumentasi spesifikasi visual & design tokens
├── go.mod
├── go.sum
└── README.md
```

## 6. Modul dan Batas Tanggung Jawab

### 6.1 Product
Mengelola master data bahan baku:
- SKU unik produk (misal: `BATOK-001`).
- Nama produk dan satuan unit (`kg`, `ton`, dll.).
- Status aktif/nonaktif (`is_active`).
- *Soft-deletion*: Produk dinonaktifkan alih-alih dihapus jika sudah memiliki referensi transaksi.

### 6.2 Supplier
Mengelola rekanan pemasok:
- Nama perusahaan/badan usaha.
- Nama kontak (PIC), nomor telepon, email, dan alamat fisik.
- Status aktif supplier (`is_active`).
- *Business Rule*: Hanya supplier yang berstatus aktif yang diizinkan untuk diterbitkan Purchase Order (PO).

### 6.3 Procurement
Mengelola siklus pengadaan barang:
- **Purchase Request (PR)**:
  - Header & Item rincian kebutuhan kuantitas.
  - Siklus status: `DRAFT` $\rightarrow$ `SUBMITTED` $\rightarrow$ `APPROVED` / `REJECTED`.
- **Purchase Order (PO)**:
  - Diterbitkan dari PR yang telah disetujui atau dibuat langsung ke supplier aktif.
  - Menyimpan kuantitas pesanan, harga satuan, dan kalkulasi total biaya.
  - Siklus status: `DRAFT` $\rightarrow$ `APPROVED` $\rightarrow$ `PARTIALLY_RECEIVED` $\rightarrow$ `COMPLETED` / `CANCELLED`.
  - PO terkunci dari perubahan saat sudah berstatus `APPROVED`.

### 6.4 Inventory
Mengelola penerimaan fisik dan saldo stok:
- **Goods Receipt (GR)**:
  - Mencatat nomor surat jalan supplier (`delivery_order_number`), tanggal datang, dan catatan penerimaan.
  - Rincian item penerimaan fisik: `received_quantity`.
- **Quality Control (QC)**:
  - Memisahkan barang menjadi `accepted_quantity` (layak pakai) dan `rejected_quantity` (rusak/kadar air tinggi) beserta alasan penolakan (`rejection_reason`).
- **Penerimaan Parsial (Multi-Truk)**:
  - Akumulasi kuantitas yang diterima terhadap kuantitas PO terbuka.
  - Transisi otomatis status PO menjadi `PARTIALLY_RECEIVED` jika masih ada sisa, atau `COMPLETED` jika pesanan sudah terpenuhi penuh.
- **Stock Ledger (Buku Besar Mutasi)**:
  - Hanya barang yang lolos QC (`accepted_quantity > 0`) yang dicatat ke buku besar mutasi bertipe `IN`.
  - Saldo stok dihitung secara dinamis dari akumulasi transaksi (`SUM(quantity)` mutasi masuk dikurangi keluar).

### 6.5 Stabilization & Response Formatter
- Mengisolasi formatting JSON respon sukses dan respon error agar konsisten di seluruh endpoint.
- Menyediakan metadata pagination (`page`, `limit`, `total_items`, `total_pages`).
- Sanitasi log: Memfilter nilai sensitif seperti password dan database credentials sebelum ditulis ke stdout.
- Panic Recovery: Menjaga server tetap hidup jika terjadi panic pada salah satu goroutine HTTP handler.

## 7. Alur Bisnis Inti (End-to-End Workflow)

```text
Kebutuhan Bahan Baku
        │
        ▼
Purchase Request (PR) [DRAFT]
        │
        ▼ submit
Purchase Request [SUBMITTED]
        │
        ▼ approve
Purchase Request [APPROVED]
        │
        ▼ create PO
Purchase Order (PO) [DRAFT] ──► Validasi: Supplier Aktif & Harga Satuan Valid
        │
        ▼ approve
Purchase Order [APPROVED] (Resmi dikirim ke Rekanan Pemasok)
        │
        ▼ Supplier Mengirim Truk / Pengiriman Fisik
Goods Receipt (Penerimaan Barang)
        │
        ├──► Quality Control (QC Inspection)
        │       ├── Diterima (Accepted)  ──► Catat Mutasi IN ke Inventory Ledger
        │       └── Ditolak (Rejected)   ──► Dicatat terpisah beserta alasan penolakan
        │
        ▼ Rekonsiliasi Kuantitas Terbuka PO
Total Diterima < Total PO?
        ├── YA  ──► PO otomatis berstatus [PARTIALLY_RECEIVED] (Truk berikutnya ditunggu)
        └── TDK ──► PO otomatis berstatus [COMPLETED] (Pesanan Selesai)
```

## 8. Skema Database & Migrasi SQL

Skema dikelola oleh sistem migrasi bawaan (`migrations/*.up.sql`):

### Tabel & Relasi Utama

```text
┌─────────────────┐       ┌────────────────────────┐
│    products     │       │       suppliers        │
├─────────────────┤       ├────────────────────────┤
│ id (PK)         │       │ id (PK)                │
│ sku (UNIQUE)    │       │ name, contact, phone   │
│ name, unit      │       │ email, address         │
│ is_active       │       │ is_active              │
└────────┬────────┘       └───────────┬────────────┘
         │                            │
         │         ┌──────────────────┴──────────────────┐
         │         │                                     │
         ▼         ▼                                     ▼
┌────────────────────────┐                    ┌────────────────────────┐
│   purchase_requests    │                    │    purchase_orders     │
├────────────────────────┤                    ├────────────────────────┤
│ id (PK)                │◄───────────────────┤ id (PK)                │
│ pr_number (UNIQUE)     │ purchase_request_id│ po_number (UNIQUE)     │
│ status, notes          │                    │ supplier_id (FK)       │
└────────┬───────────────┘                    │ status, notes          │
         │                                    └───────────┬────────────┘
         ▼                                                │
┌────────────────────────┐                                │
│ purchase_request_items │                                ▼
├────────────────────────┤                    ┌────────────────────────┐
│ id (PK)                │                    │  purchase_order_items  │
│ purchase_request_id(FK)│                    ├────────────────────────┤
│ product_id (FK)        │                    │ id (PK)                │
│ quantity, notes        │                    │ purchase_order_id (FK) │
└────────────────────────┘                    │ product_id (FK)        │
                                              │ quantity, unit_price   │
                                              └───────────┬────────────┘
                                                          │
         ┌────────────────────────────────────────────────┘
         │
         ▼
┌────────────────────────┐       ┌────────────────────────┐
│     goods_receipts     │       │  goods_receipt_items   │
├────────────────────────┤       ├────────────────────────┤
│ id (PK)                │◄──────┤ id (PK)                │
│ gr_number (UNIQUE)     │       │ goods_receipt_id (FK)  │
│ purchase_order_id (FK) │       │ po_item_id (FK)        │
│ delivery_order_number  │       │ product_id (FK)        │
│ received_date, notes   │       │ received_quantity      │
└────────────────────────┘       │ accepted_quantity      │
                                 │ rejected_quantity      │
                                 │ rejection_reason       │
                                 └───────────┬────────────┘
                                             │
                                             ▼
                                 ┌────────────────────────┐
                                 │ inventory_transactions │
                                 ├────────────────────────┤
                                 │ id (PK)                │
                                 │ product_id (FK)        │
                                 │ transaction_type (IN)  │
                                 │ quantity               │
                                 │ reference_type (GR)    │
                                 │ reference_id (FK GR)   │
                                 │ notes, created_at      │
                                 └────────────────────────┘
```

### Daftar Migrasi yang Telah Dijalankan:
1. `000001_create_products_table.up.sql`: Pembuatan tabel `products` dengan index SKU unik.
2. `000002_create_suppliers_table.up.sql`: Pembuatan tabel `suppliers` dan status aktif.
3. `000003_create_procurement_tables.up.sql`: Pembuatan tabel `purchase_requests`, `purchase_request_items`, `purchase_orders`, dan `purchase_order_items`.
4. `000004_create_inventory_tables.up.sql`: Pembuatan tabel `goods_receipts`, `goods_receipt_items`, dan `inventory_transactions`.
5. `000005_ensure_suppliers_columns.up.sql`: Verifikasi idempotent kolom dan constraint integritas data supplier.

## 9. Inventory Ledger (Buku Besar Stok)

Sistem mengadopsi prinsip akuntansi buku besar (ledger):
- Kolom saldo statis pada tabel produk dihindari karena rentan mengalami kondisi balapan (race condition) dan ketidaksesuaian data.
- Setiap penerimaan barang fisik (`accepted_quantity`) menambahkan baris mutasi `IN` pada tabel `inventory_transactions`.
- Pengeluaran bahan baku di masa mendatang (penjualan/pemakaian) akan menambahkan baris mutasi `OUT`.
- Saldo stok dihitung melalui agregasi SQL:
  ```sql
  SELECT product_id,
         COALESCE(SUM(CASE WHEN transaction_type = 'IN' THEN quantity ELSE 0 END), 0) AS total_in,
         COALESCE(SUM(CASE WHEN transaction_type = 'OUT' THEN quantity ELSE 0 END), 0) AS total_out,
         COALESCE(SUM(CASE WHEN transaction_type = 'IN' THEN quantity ELSE -quantity END), 0) AS current_stock
  FROM inventory_transactions
  GROUP BY product_id;
  ```

## 10. Daftar Endpoint REST API Lengkap

Prefix API: `/api/v1`

### 10.1 Health Checks
| Method | Endpoint | Deskripsi |
|---|---|---|
| `GET` | `/health/live` | Liveness probe (memverifikasi server Go aktif) |
| `GET` | `/health/ready` | Readiness probe (memverifikasi kesiapan koneksi database Neon) |

### 10.2 Products
| Method | Endpoint | Deskripsi |
|---|---|---|
| `POST` | `/api/v1/products` | Tambah produk baru (SKU unik, nama, satuan) |
| `GET` | `/api/v1/products` | Ambil daftar produk (mendukung pagination) |
| `GET` | `/api/v1/products/{id}` | Detail produk berdasarkan ID |
| `PUT` | `/api/v1/products/{id}` | Perbarui nama, unit, atau status produk |
| `DELETE` | `/api/v1/products/{id}` | Nonaktifkan produk (*soft-delete*) |

### 10.3 Suppliers
| Method | Endpoint | Deskripsi |
|---|---|---|
| `POST` | `/api/v1/suppliers` | Tambah supplier baru |
| `GET` | `/api/v1/suppliers` | Ambil daftar supplier |
| `GET` | `/api/v1/suppliers?active=true` | Ambil daftar supplier yang berstatus aktif saja |
| `GET` | `/api/v1/suppliers/{id}` | Detail supplier berdasarkan ID |
| `PUT` | `/api/v1/suppliers/{id}` | Perbarui data kontak/status supplier |
| `DELETE` | `/api/v1/suppliers/{id}` | Nonaktifkan supplier |

### 10.4 Procurement — Purchase Request (PR)
| Method | Endpoint | Deskripsi |
|---|---|---|
| `POST` | `/api/v1/purchase-requests` | Buat PR baru status `DRAFT` beserta rincian item |
| `GET` | `/api/v1/purchase-requests` | Ambil daftar PR |
| `GET` | `/api/v1/purchase-requests/{id}` | Detail PR dan item-itemnya |
| `POST` | `/api/v1/purchase-requests/{id}/submit` | Ajukan PR (`DRAFT` $\rightarrow$ `SUBMITTED`) |
| `POST` | `/api/v1/purchase-requests/{id}/approve` | Setujui PR (`SUBMITTED` $\rightarrow$ `APPROVED`) |
| `POST` | `/api/v1/purchase-requests/{id}/reject` | Tolak PR (`SUBMITTED` $\rightarrow$ `REJECTED`) |

### 10.5 Procurement — Purchase Order (PO)
| Method | Endpoint | Deskripsi |
|---|---|---|
| `POST` | `/api/v1/purchase-orders` | Buat PO ke supplier aktif |
| `GET` | `/api/v1/purchase-orders` | Ambil daftar PO |
| `GET` | `/api/v1/purchase-orders/{id}` | Detail PO, supplier, dan rincian harga |
| `POST` | `/api/v1/purchase-orders/{id}/approve` | Setujui PO untuk dikirim ke supplier (`APPROVED`) |
| `POST` | `/api/v1/purchase-orders/{id}/cancel` | Batalkan PO (`CANCELLED`) |

### 10.6 Inventory & Goods Receipt
| Method | Endpoint | Deskripsi |
|---|---|---|
| `POST` | `/api/v1/goods-receipts` | Catat penerimaan fisik, QC, dan posting mutasi stok |
| `GET` | `/api/v1/goods-receipts` | Ambil riwayat penerimaan barang |
| `GET` | `/api/v1/goods-receipts/{id}` | Detail penerimaan barang dan hasil QC |
| `GET` | `/api/v1/inventory/stock` | Saldo stok terkini setiap produk (`total_in`, `total_out`, `current_stock`) |
| `GET` | `/api/v1/inventory/ledger` | Riwayat buku besar mutasi stok (opsional filter `?product_id=1`) |

## 11. Format Respon & Standarisasi Error

Semua response menggunakan header `Content-Type: application/json; charset=utf-8`.

### Format Respon Berhasil
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

### Format Respon Gagal
```json
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Supplier tidak aktif atau tidak ditemukan",
    "details": null
  }
}
```

Kode status HTTP yang digunakan:
- `200 OK`: Pengambilan data atau pembaruan status berhasil.
- `201 Created`: Pembuatan entitas berhasil (`products`, `suppliers`, `PR`, `PO`, `GR`).
- `400 Bad Request`: Format JSON tidak valid atau validasi input gagal.
- `404 Not Found`: Entitas tidak ditemukan.
- `409 Conflict`: Pelanggaran constraint (misal: duplikasi SKU unik).
- `422 Unprocessable Entity`: Pelanggaran aturan bisnis (misal: PO ke supplier nonaktif, kuantitas penerimaan melebihi sisa PO).
- `500 Internal Server Error`: Kegagalan internal database atau server (detail internal tidak dibocorkan).

## 12. Konfigurasi dan Keamanan

1. **Environment Variables**:
   - `DATABASE_URL`: String koneksi Neon PostgreSQL dengan parameter `sslmode=require`.
   - `HTTP_ADDR`: Port listen web server (default `:8080`).
   - `APP_ENV`: Environment running (`development` / `production`).
2. **Kerahasiaan Kredensial**:
   - File `.env` masuk ke `.gitignore`. Template tersedia di `.env.example`.
   - Middleware logging melakukan *redaction* terhadap connection string dan token sensitif.
3. **CORS (Cross-Origin Resource Sharing)**:
   - Dikonfigurasi untuk mendukung komunikasi aman antara antarmuka Next.js dan backend Go.

## 13. Testing Strategy

1. **Unit Testing & Business Rules**: Menguji validasi format data, perhitungan kuantitas, dan error handling.
2. **End-to-End Workflow Test (`tests/e2e_workflow_test.go`)**:
   - Simulasi alur lengkap: Pembuatan produk $\rightarrow$ pembuatan supplier $\rightarrow$ pengajuan PR $\rightarrow$ approval PR $\rightarrow$ pembuatan PO $\rightarrow$ approval PO $\rightarrow$ penerimaan parsial Truk 1 $\rightarrow$ penerimaan penyelesaian Truk 2 $\rightarrow$ validasi saldo akhir buku besar stok.
3. **API Collection (`bruno/`)**:
   - Koleksi request siap pakai untuk eksplorasi dan verifikasi manual setiap endpoint REST API.

## 14. Roadmap Implementasi

### Phase 1 — Foundation (Selesai)
- [x] Project Neon PostgreSQL terkonfigurasi.
- [x] Inisialisasi Go module dan struktur modular monolith.
- [x] Koneksi pool `pgxpool` dengan SSL support.
- [x] Endpoint liveness (`/health/live`) dan readiness (`/health/ready`).
- [x] Custom automated migration runner terintegrasi.

### Phase 2 — Master Data (Selesai)
- [x] Migrasi tabel `products` dan `suppliers`.
- [x] CRUD dan validasi SKU unik untuk produk.
- [x] CRUD dan validasi nomor kontak supplier.
- [x] Fitur soft-delete (`is_active`).

### Phase 3 — Procurement (Selesai)
- [x] Migrasi tabel `purchase_requests` dan `purchase_orders`.
- [x] Workflow PR: `DRAFT` $\rightarrow$ `SUBMITTED` $\rightarrow$ `APPROVED` / `REJECTED`.
- [x] Workflow PO: Pembuatan PO dengan validasi supplier aktif.
- [x] PO Approval & Cancellation.

### Phase 4 — Inventory & Quality Control (Selesai)
- [x] Migrasi tabel `goods_receipts` dan `inventory_transactions`.
- [x] Pencatatan Goods Receipt dengan nomor surat jalan rekanan.
- [x] Pemisahan inspeksi QC (`accepted_quantity` vs `rejected_quantity`).
- [x] Penerimaan bertahap / parsial (rekonsiliasi otomatis status PO).
- [x] Buku besar stok (*Stock Ledger*) transaksi atomik `IN`.

### Phase 5 — Stabilization (Selesai)
- [x] Standarisasi format envelope JSON respon dan pagination metadata.
- [x] Panic recovery middleware untuk mencegah web service crash.
- [x] Sanitasi log untuk melindungi data sensitif.
- [x] End-to-end integration test (`tests/e2e_workflow_test.go`).
- [x] Dokumentasi lengkap operasional pada `README.md`.

### Phase 6 — Frontend Web Application (Selesai)
- [x] Setup Next.js 14 (App Router) + TypeScript.
- [x] Implementasi Custom Design System Glassmorphism ([DESIGN.md](DESIGN.md)).
- [x] Dashboard metrik operasional dan ringkasan stok cepat.
- [x] Antarmuka Master Data: Manajemen Produk dan Supplier (tambah, edit, nonaktifkan).
- [x] Antarmuka Procurement: Pembuatan dan persetujuan PR & PO.
- [x] Antarmuka Inventory: Input Goods Receipt, pencatatan QC, dan visualisasi Stock Ledger.
- [x] Client API wrapper terintegrasi dengan backend Go (`web/lib/api.ts`).

---

## 15. Rencana Pengembangan Selanjutnya (Next Phases)

Setelah MVP operasional hulu-ke-hilir selesai, pengembangan lanjutan akan difokuskan pada:

### Phase 7 — Autentikasi & Role-Based Access Control (RBAC)
- [ ] Registrasi & Login pengguna (JWT / HttpOnly Cookie).
- [ ] Pemisahan peran akses:
  - `Admin`: Akses penuh konfigurasi sistem dan user.
  - `Purchasing`: Membuat PR, membuat PO, mengelola master supplier.
  - `Manager / Approver`: Menyetujui atau menolak PR dan PO.
  - `Warehouse / QC`: Mencatat Goods Receipt, inspeksi mutu barang, dan melihat stok.

### Phase 8 — Outbound Logistics & Multi-Warehouse
- [ ] Master data multi-gudang (`warehouses`) dan alokasi lokasi penyimpanan.
- [ ] Delivery Order (DO) / Surat Jalan Pengiriman ke Pelanggan.
- [ ] Pencatatan mutasi pengeluaran stok (`OUT`) pada Stock Ledger.
- [ ] Reservasi stok untuk pesanan penjualan yang belum dikirim.

### Phase 9 — Laporan & Export Data
- [ ] Ekspor dokumen cetak PDF untuk Purchase Order (format surat resmi perusahaan).
- [ ] Ekspor laporan mutasi buku besar stok dan penerimaan barang ke Excel / CSV.
- [ ] Notifikasi batas minimum stok (*low-stock reorder warning*).

