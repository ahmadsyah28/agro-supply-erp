-- Memastikan kolom-kolom tabel suppliers lengkap jika sebelumnya sudah ada tabel lama di database
ALTER TABLE suppliers ADD COLUMN IF NOT EXISTS contact_name VARCHAR(255);
ALTER TABLE suppliers ADD COLUMN IF NOT EXISTS email VARCHAR(255);
ALTER TABLE suppliers ADD COLUMN IF NOT EXISTS phone VARCHAR(64);
ALTER TABLE suppliers ADD COLUMN IF NOT EXISTS address TEXT;
ALTER TABLE suppliers ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE suppliers ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE suppliers ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- Lepas constraint NOT NULL dan UNIQUE pada kolom legacy 'code' jika tabel suppliers lama memilikinya
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'suppliers' AND column_name = 'code'
    ) THEN
        ALTER TABLE suppliers DROP CONSTRAINT IF EXISTS suppliers_code_key;
        ALTER TABLE suppliers ALTER COLUMN code DROP NOT NULL;
        ALTER TABLE suppliers ALTER COLUMN code DROP DEFAULT;
    END IF;
END $$;

-- Memastikan kolom-kolom tabel products juga lengkap
ALTER TABLE products ADD COLUMN IF NOT EXISTS sku VARCHAR(64);
ALTER TABLE products ADD COLUMN IF NOT EXISTS name VARCHAR(255);
ALTER TABLE products ADD COLUMN IF NOT EXISTS unit VARCHAR(32);
ALTER TABLE products ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE products ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE products ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP;
