# PT AgroSupply ERP — System Architecture

> Dokumen arsitektur teknis untuk MVP ERP PT AgroSupply.
>
> Status: Initial design  
> Fokus saat ini: Backend Go + Neon PostgreSQL  
> Frontend: Dikembangkan setelah API inti siap

## 1. Tujuan Sistem

PT AgroSupply adalah distributor bahan baku, misalnya batok kelapa, yang membeli barang dari supplier dan memasoknya kepada pelanggan.

MVP ini berfokus pada proses procurement dan inventory agar perusahaan dapat:
- Mengelola data produk dan supplier.
- Mencatat kebutuhan pengadaan melalui Purchase Request (PR).
- Membuat dan menyetujui Purchase Order (PO).
- Mencatat penerimaan barang, termasuk penerimaan parsial.
- Melacak pergerakan dan saldo stok.
- Menjaga riwayat transaksi dan perubahan penting.

Fitur sales lengkap, akuntansi, pembayaran, multi-level approval kompleks, dan integrasi eksternal ditunda sampai kebutuhan inti berjalan.

## 2. Prinsip Arsitektur

1. **Modular Monolith** — satu aplikasi backend, dengan modul domain yang dipisahkan secara logis.
2. **Business rules di backend** — aturan bisnis tidak hanya bergantung pada frontend.
3. **Database sebagai sumber data transaksi** — data penting disimpan secara persisten di PostgreSQL.
4. **Inventory ledger** — pergerakan stok dicatat sebagai transaksi, bukan hanya mengubah angka stok tanpa riwayat.
5. **Explicit workflow** — PR, PO, dan Goods Receipt memiliki status dan transisi yang divalidasi.
6. **Build incrementally** — implementasikan dan tes satu bagian sebelum menambah modul berikutnya.
7. **Avoid premature complexity** — belum menggunakan microservices, message broker, atau CQRS untuk MVP.

## 3. Technology Stack

| Layer | Teknologi | Tanggung jawab |
|---|---|---|
| Frontend (tahap berikutnya) | Next.js + TypeScript | UI dan interaksi pengguna |
| Backend | Go | HTTP API, business rules, dan orchestration |
| HTTP routing | `net/http` | Routing dan penanganan HTTP |
| PostgreSQL driver | `github.com/jackc/pgx/v5/pgxpool` | Koneksi dan query database |
| Database | Neon PostgreSQL | Penyimpanan data cloud |
| Schema migration | `golang-migrate` atau tool sejenis | Perubahan schema yang terkontrol |
| API testing | Postman atau Bruno | Pengujian endpoint |
| Version control | Git | Riwayat perubahan kode |

Gunakan versi dependency yang kompatibel dan simpan dependency melalui `go.mod` serta `go.sum`.

## 4. High-Level Architecture

```text
User / API Client
       |
       v
Next.js Frontend (later)
       |
       | HTTP / JSON
       v
Go REST API
       |
       +--> HTTP Handlers
       |       |
       |       v
       +--> Application / Business Logic
       |       |
       |       v
       +--> Data Access (SQL / pgx)
               |
               v
        Neon PostgreSQL
```

Tanggung jawab setiap bagian:

- **Handler**: membaca request, memvalidasi format input dasar, memanggil application logic, dan membentuk HTTP response.
- **Application/business logic**: menerapkan aturan bisnis, mengatur workflow, dan menentukan operasi yang harus dilakukan.
- **Data access**: menjalankan query SQL dan memetakan hasil database ke struktur Go.
- **PostgreSQL**: menyimpan data, menjaga relasi, constraint, dan transaksi database.

Jangan menaruh seluruh business rules di handler. Untuk MVP, layer tidak perlu dibuat terlalu banyak; pisahkan ketika tanggung jawabnya memang berbeda.

## 5. Struktur Folder yang Disarankan

```text
agro-supply-erp/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── database/
│   │   └── postgres.go
│   ├── product/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── model.go
│   ├── supplier/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── model.go
│   ├── procurement/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── model.go
│   └── inventory/
│       ├── handler.go
│       ├── service.go
│       ├── repository.go
│       └── model.go
├── migrations/
├── tests/
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

Catatan:
- Struktur di atas adalah target bertahap, bukan berarti semua file harus dibuat pada hari pertama.
- `cmd/api/main.go` menjadi entry point.
- `internal/` berisi kode aplikasi yang tidak dimaksudkan sebagai public package untuk aplikasi Go lain.
- `migrations/` berisi perubahan schema database yang diberi nomor dan dijalankan secara terkontrol.
- Hindari membuat package kosong sebelum ada implementasi yang membutuhkannya.

## 6. Modul dan Batas Tanggung Jawab

### 6.1 Product
Mengelola master data produk, misalnya:
- SKU / kode produk
- Nama produk
- Unit dasar, misalnya kg atau ton
- Status aktif/nonaktif

### 6.2 Supplier
Mengelola data supplier:
- Nama perusahaan
- Kontak dan alamat
- Status supplier
- Informasi kapasitas, kualitas, lokasi, dan lead time jika diperlukan

Supplier yang belum aktif tidak boleh dipakai untuk membuat PO sesuai business rule yang ditetapkan.

### 6.3 Procurement
Mengelola:
- Purchase Request (PR)
- Item PR
- Approval PR
- Purchase Order (PO)
- Item PO
- Status dan tindak lanjut kekurangan pengiriman

### 6.4 Inventory
Mengelola:
- Goods Receipt
- Pemeriksaan kuantitas dan kualitas
- Penerimaan parsial
- Stock ledger / inventory transactions
- Saldo stok per produk dan gudang (jika gudang sudah dimodelkan)

### 6.5 Auth (tahap berikutnya)
Mengelola autentikasi, identitas pengguna, dan otorisasi. Jangan menganggap menyembunyikan tombol di frontend sebagai kontrol akses yang cukup; backend harus memvalidasi izin.

## 7. Alur Bisnis Inti

```text
Kebutuhan barang
      |
      v
Purchase Request (PR)
      |
      v
Review / Approval
      |
      v
Pemilihan Supplier
      |
      v
Purchase Order (PO)
      |
      v
Supplier mengirim barang
      |
      v
Goods Receipt + pemeriksaan kualitas
      |
      +--> Barang diterima --> catat inventory IN
      |
      +--> Barang ditolak --> catat hasil inspeksi dan tindak lanjut
      |
      v
Rekonsiliasi kuantitas PO dan penerimaan
```

Contoh penerimaan parsial:
- PO: 7.000 kg
- Barang diterima secara fisik: 6.000 kg
- Kekurangan: 1.000 kg
- PO tetap menyimpan kuantitas awal 7.000 kg.
- Goods Receipt mencatat jumlah aktual 6.000 kg.
- PO dapat berstatus `partially_received` selama masih ada kuantitas terbuka.
- Stok hanya bertambah berdasarkan jumlah yang diterima dan dinyatakan layak digunakan.
- Kekurangan ditindaklanjuti berdasarkan kebijakan perusahaan.

## 8. Rancangan Data Awal

Schema ini adalah gambaran awal. Detail constraint dan tipe data akan ditentukan saat membuat migration.

### Master data
- `products`
- `suppliers`
- `warehouses` (dapat ditambahkan saat multi-gudang dibutuhkan)
- `users` dan `roles` (ketika autentikasi dan otorisasi mulai dibuat)

### Procurement
- `purchase_requests`
- `purchase_request_items`
- `purchase_orders`
- `purchase_order_items`
- `goods_receipts`
- `goods_receipt_items`

### Inventory
- `inventory_transactions`
- Opsional: tabel saldo stok yang dioptimalkan setelah kebutuhan query diketahui

Relasi utama:
- Satu PR memiliki banyak PR items.
- Satu PO memiliki banyak PO items.
- Satu PO dapat memiliki beberapa Goods Receipt.
- Satu Goods Receipt memiliki banyak Goods Receipt items.
- Inventory transaction harus memiliki referensi yang dapat ditelusuri ke transaksi sumbernya.

### Aturan data penting
- Gunakan primary key yang konsisten.
- Gunakan foreign key untuk menjaga referensi antarentitas.
- Simpan kuantitas dengan tipe numerik yang sesuai, bukan floating-point biner.
- Simpan harga sebagai `NUMERIC`/decimal, bukan `float`.
- Simpan timestamp dengan tipe yang sesuai dan gunakan UTC pada backend bila memungkinkan.
- Tambahkan `created_at` dan `updated_at` pada tabel yang memerlukannya.
- Jangan menyimpan password plaintext.
- Hindari menghapus transaksi bisnis yang sudah diposting; gunakan pembatalan atau reversal yang tercatat sesuai kebutuhan.

## 9. Inventory Ledger

Jangan hanya menyimpan `products.stock` sebagai satu-satunya sumber kebenaran.

Setiap perubahan stok harus menghasilkan transaksi, misalnya:

| Referensi | Jenis | Kuantitas |
|---|---|---:|
| GR-001 | IN | +6.000 kg |
| DO-001 | OUT | -1.000 kg |

Saldo stok secara konsep adalah jumlah seluruh transaksi masuk dan keluar untuk produk serta lokasi yang sama, dengan mempertimbangkan transaksi yang valid.

Untuk mencegah inkonsistensi:
- Catat Goods Receipt dan inventory transaction dalam satu transaksi database.
- Gunakan constraint dan validasi untuk mencegah data tidak valid.
- Jika kelak menyimpan saldo stok terhitung untuk performa, perbarui saldo dan ledger secara atomik.
- Rancang idempotency atau proteksi duplikasi sebelum endpoint posting transaksi digunakan secara luas.

## 10. API Awal

Gunakan prefix `/api/v1`.

### Health
- `GET /health/live`
- `GET /health/ready`

### Products
- `POST /api/v1/products`
- `GET /api/v1/products`
- `GET /api/v1/products/{id}`
- `PUT /api/v1/products/{id}`
- `DELETE /api/v1/products/{id}` atau nonaktifkan produk jika sudah direferensikan transaksi

### Suppliers
- `POST /api/v1/suppliers`
- `GET /api/v1/suppliers`
- `GET /api/v1/suppliers/{id}`
- `PUT /api/v1/suppliers/{id}`

### Procurement dan inventory (tahap berikutnya)
- `POST /api/v1/purchase-requests`
- `POST /api/v1/purchase-requests/{id}/submit`
- `POST /api/v1/purchase-requests/{id}/approve`
- `POST /api/v1/purchase-orders`
- `POST /api/v1/purchase-orders/{id}/approve`
- `POST /api/v1/goods-receipts`
- `GET /api/v1/inventory`

Endpoint di atas adalah rancangan awal, bukan berarti semua harus dibuat sekaligus.

## 11. Aturan HTTP API

- Gunakan JSON untuk request dan response.
- Gunakan HTTP status yang sesuai: `200`, `201`, `204`, `400`, `401`, `403`, `404`, `409`, dan `500`.
- Validasi input di backend.
- Jangan mengirim detail internal database atau secret dalam error response.
- Gunakan format error response yang konsisten.
- Gunakan pagination untuk list yang dapat bertambah besar.
- Pisahkan error validasi, konflik business rule, dan kegagalan internal.
- Jangan menganggap request sukses sebelum transaksi database berhasil di-commit.

## 12. Konfigurasi dan Keamanan

Konfigurasi runtime diambil dari environment variables, misalnya:
- `DATABASE_URL`
- `APP_ENV`
- `HTTP_ADDR`

Ketentuan:
- Simpan secret di environment, bukan di source code.
- Jangan commit file `.env`; commit `.env.example` tanpa secret asli.
- Tambahkan `.env` ke `.gitignore`.
- Gunakan koneksi TLS sesuai konfigurasi Neon.
- Gunakan database role dengan hak minimum yang dibutuhkan.
- Jangan log connection string, password, token, atau data sensitif.
- Untuk deployment production, gunakan pengelolaan secret yang sesuai platform.

Contoh `.env.example`:

```dotenv
APP_ENV=development
HTTP_ADDR=:8080
DATABASE_URL=postgresql://USER:PASSWORD@HOST/DATABASE?sslmode=require
```

Ganti placeholder dengan nilai yang sesuai secara lokal. Jangan menggunakan contoh tersebut sebagai kredensial nyata.

## 13. Testing Strategy

Tahap awal:
1. Unit test untuk validasi dan business rules.
2. Integration test untuk query PostgreSQL.
3. API test untuk status code dan response.
4. Workflow test untuk PR → PO → Goods Receipt.
5. Negative test untuk kondisi seperti supplier nonaktif, PO tidak disetujui, penerimaan melebihi kuantitas terbuka, dan penerimaan duplikat.

Contoh skenario penerimaan:
- PO 7.000 kg, menerima 6.000 kg: PO menjadi `partially_received`.
- Sisa 1.000 kg diterima kemudian: kuantitas penerimaan terakumulasi dengan benar.
- Kuantitas penerimaan tidak boleh otomatis mengubah kuantitas awal PO.
- Barang yang ditolak tidak otomatis menambah stok yang layak digunakan.

## 14. Roadmap Implementasi

### Phase 1 — Foundation
- [x] Buat project Neon PostgreSQL.
- [x] Buat module Go dan struktur awal.
- [x] Koneksi Go ke Neon melalui `pgxpool`.
- [x] Implementasikan health endpoints.
- [x] Siapkan migration tooling dan konfigurasi environment.


### Phase 2 — Master Data
- [x] Migration `products`.
- [x] CRUD products.
- [x] Migration `suppliers`.
- [x] CRUD suppliers.
- [x] Validasi dan API tests.

### Phase 3 — Procurement
- [ ] PR dan PR items.
- [ ] Workflow submit/approve sederhana.
- [ ] PO dan PO items.
- [ ] Validasi status dan supplier aktif.

### Phase 4 — Inventory
- [ ] Goods Receipt dan item.
- [ ] Penerimaan parsial.
- [ ] Inventory ledger.
- [ ] Transaksi database yang atomik.
- [ ] Pengujian penerimaan dan stok.

### Phase 5 — Stabilization
- [ ] Error response konsisten.
- [ ] Pagination dan filtering dasar.
- [ ] Logging tanpa secret.
- [ ] Testing workflow end-to-end.
- [ ] README untuk menjalankan project.

### Phase 6 — Frontend
- [ ] Setup Next.js + TypeScript.
- [ ] Halaman master products dan suppliers.
- [ ] Halaman PR dan PO.
- [ ] Halaman penerimaan barang dan stok.
- [ ] Integrasi frontend dengan API Go.

## 15. Definition of Done

Sebuah fitur dianggap selesai jika:
- Business rule yang relevan telah ditentukan.
- Input divalidasi oleh backend.
- Database operation berjalan dengan benar.
- Error response sesuai dan tidak membocorkan detail internal.
- Ada test untuk skenario normal dan skenario gagal.
- Perubahan schema memiliki migration.
- Kode dapat dijalankan dari instruksi README.

## 16. Keputusan yang Sengaja Ditunda

Untuk menjaga MVP tetap sederhana, hal-hal berikut belum menjadi bagian tahap pertama:
- Microservices.
- Message broker/event streaming.
- CQRS.
- Multi-level approval yang kompleks.
- Accounting dan general ledger lengkap.
- Integrasi pembayaran.
- Multi-tenant architecture.
- Optimasi performa sebelum ada pengukuran.

Keputusan ini dapat ditinjau kembali jika kebutuhan bisnis dan skala sistem memang membutuhkannya.

---

## Next action

Mulai dari **Phase 1 — Foundation**:
1. Buat project Neon dan simpan `DATABASE_URL` secara aman.
2. Pastikan Go dapat terhubung ke Neon.
3. Pastikan `/health/live` dan `/health/ready` berfungsi.
4. Tambahkan migration tooling.
5. Buat tabel `products` sebagai schema bisnis pertama.
