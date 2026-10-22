-- Goods Receipts (Penerimaan Barang)
CREATE TABLE IF NOT EXISTS goods_receipts (
    id BIGSERIAL PRIMARY KEY,
    gr_number VARCHAR(64) NOT NULL UNIQUE,
    purchase_order_id BIGINT NOT NULL REFERENCES purchase_orders(id),
    delivery_order_number VARCHAR(128), -- Nomor Surat Jalan dari Supplier
    receipt_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_gr_gr_number ON goods_receipts(gr_number);
CREATE INDEX IF NOT EXISTS idx_gr_po_id ON goods_receipts(purchase_order_id);

-- Goods Receipt Items (Rincian Fisik & Kualitas Barang)
CREATE TABLE IF NOT EXISTS goods_receipt_items (
    id BIGSERIAL PRIMARY KEY,
    goods_receipt_id BIGINT NOT NULL REFERENCES goods_receipts(id) ON DELETE CASCADE,
    purchase_order_item_id BIGINT NOT NULL REFERENCES purchase_order_items(id),
    product_id BIGINT NOT NULL REFERENCES products(id),
    received_quantity NUMERIC(15, 3) NOT NULL CHECK (received_quantity >= 0),
    accepted_quantity NUMERIC(15, 3) NOT NULL CHECK (accepted_quantity >= 0),
    rejected_quantity NUMERIC(15, 3) NOT NULL DEFAULT 0 CHECK (rejected_quantity >= 0),
    rejection_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_gr_items_gr_id ON goods_receipt_items(goods_receipt_id);
CREATE INDEX IF NOT EXISTS idx_gr_items_product_id ON goods_receipt_items(product_id);

-- Inventory Transactions (Buku Besar / Ledger Pergerakan Stok)
CREATE TABLE IF NOT EXISTS inventory_transactions (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id),
    transaction_type VARCHAR(32) NOT NULL, -- IN, OUT, ADJUSTMENT
    reference_type VARCHAR(64) NOT NULL,   -- GOODS_RECEIPT, DELIVERY_ORDER, STOCK_OPNAME
    reference_id BIGINT NOT NULL,          -- ID dari dokumen sumber (misal goods_receipt_id)
    reference_number VARCHAR(64) NOT NULL, -- Nomor dokumen sumber (misal GR-202610-0001)
    quantity NUMERIC(15, 3) NOT NULL CHECK (quantity > 0),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_inv_trans_product_id ON inventory_transactions(product_id);
CREATE INDEX IF NOT EXISTS idx_inv_trans_ref ON inventory_transactions(reference_type, reference_id);
CREATE INDEX IF NOT EXISTS idx_inv_trans_created_at ON inventory_transactions(created_at);
