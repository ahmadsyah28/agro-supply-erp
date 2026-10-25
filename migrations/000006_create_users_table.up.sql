-- Create users table for Authentication & RBAC (Phase 7)
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('ADMIN', 'PURCHASING', 'MANAGER', 'WAREHOUSE')),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- Seed default demo users for the 4 primary ERP roles (Password: password123)
-- Hash: $2a$10$wOuekgVZintN1dOq5mWjzen2TlP/aZ5tL9IGb66Qm0v2kQXC63Deq
INSERT INTO users (username, email, password_hash, full_name, role, is_active)
VALUES
    ('admin', 'admin@agrosupply.local', '$2a$10$wOuekgVZintN1dOq5mWjzen2TlP/aZ5tL9IGb66Qm0v2kQXC63Deq', 'System Administrator', 'ADMIN', TRUE),
    ('purchasing', 'purchasing@agrosupply.local', '$2a$10$wOuekgVZintN1dOq5mWjzen2TlP/aZ5tL9IGb66Qm0v2kQXC63Deq', 'Budi Purchasing', 'PURCHASING', TRUE),
    ('manager', 'manager@agrosupply.local', '$2a$10$wOuekgVZintN1dOq5mWjzen2TlP/aZ5tL9IGb66Qm0v2kQXC63Deq', 'Siti Manager', 'MANAGER', TRUE),
    ('warehouse', 'warehouse@agrosupply.local', '$2a$10$wOuekgVZintN1dOq5mWjzen2TlP/aZ5tL9IGb66Qm0v2kQXC63Deq', 'Joko Gudang & QC', 'WAREHOUSE', TRUE)
ON CONFLICT (username) DO NOTHING;
