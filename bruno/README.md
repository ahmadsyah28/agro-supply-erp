# Bruno API Collection - PT AgroSupply ERP

Koleksi pengujian API menggunakan **[Bruno](https://www.usebruno.com/)** (Open Source API Client alternatif Postman/Insomnia).

## Struktur Koleksi
- **00-Health**: Liveness & Readiness probe database
- **01-Suppliers**: CRUD Supplier / Pemasok
- **02-Products**: CRUD Master Data Produk & Satuan
- **03-Procurement-PR**: Purchase Request (Buat, List, Detail, Submit, Approve, Reject)
- **04-Procurement-PO**: Purchase Order (Buat PO, List, Detail, Approve, Cancel)
- **05-Inventory**: Penerimaan Barang (Goods Receipt), Cek Saldo Stok Gudang, Buku Besar Kartu Stok (Ledger)

## Cara Menggunakan

### Cara 1: Menggunakan Bruno Desktop App (GUI)
1. Unduh dan pasang aplikasi Bruno Desktop dari **[usebruno.com](https://www.usebruno.com/)**.
2. Buka aplikasi Bruno.
3. Di halaman awal, pilih **"Open Collection"**.
4. Arahkan dan pilih folder:
   ```
   f:\erp\agro-supply-erp\bruno
   ```
5. Pada dropdown environment (pojok kanan atas di Bruno), pilih **`local`** (baseUrl mengarah ke `http://localhost:8080`).
6. Jalankan request mana saja atau klik kanan folder dan pilih **"Run"** untuk mengeksekusi semua request.

---

### Cara 2: Menjalankan via Terminal (CLI)
Pastikan server backend Go sedang berjalan (`go run ./cmd/api`), lalu buka terminal baru dan jalankan:

```powershell
cd f:\erp\agro-supply-erp\bruno
npx @usebruno/cli run --env local
```
Semua endpoint dan assertion akan diuji secara otomatis dan hasilnya akan ditampilkan dalam bentuk ringkasan tabel di terminal.
