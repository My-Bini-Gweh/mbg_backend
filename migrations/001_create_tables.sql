ALTER DATABASE CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS mahasiswa (
    id_mahasiswa BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    nrp VARCHAR(32) NOT NULL UNIQUE,
    nama_mahasiswa VARCHAR(120) NOT NULL,
    email VARCHAR(120) NOT NULL UNIQUE,
    role ENUM('mahasiswa', 'admin') NOT NULL DEFAULT 'mahasiswa',
    status ENUM('ACTIVE', 'INACTIVE', 'SUSPENDED') NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS mahasiswa_auth (
    id_auth BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    mahasiswa_id BIGINT UNSIGNED NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    pin_hash VARCHAR(255) NULL,
    last_login_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_mahasiswa_auth_mahasiswa
        FOREIGN KEY (mahasiswa_id) REFERENCES mahasiswa(id_mahasiswa)
        ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS wallet (
    id_wallet VARCHAR(20) PRIMARY KEY,
    mahasiswa_id BIGINT UNSIGNED NOT NULL UNIQUE,
    jenis_wallet ENUM('REGULAR', 'ADMIN') NOT NULL DEFAULT 'REGULAR',
    saldo DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_wallet_mahasiswa
        FOREIGN KEY (mahasiswa_id) REFERENCES mahasiswa(id_mahasiswa)
        ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT chk_wallet_saldo_non_negative CHECK (saldo >= 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS bank (
    id_bank BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    nama_bank VARCHAR(80) NOT NULL,
    kode_bank VARCHAR(20) NOT NULL UNIQUE,
    biaya_admin DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS rekening_mahasiswa (
    id_rekening BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    mahasiswa_id BIGINT UNSIGNED NOT NULL,
    bank_id_bank BIGINT UNSIGNED NOT NULL,
    no_rekening VARCHAR(40) NOT NULL,
    nama_pemilik VARCHAR(120) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_rekening_mahasiswa
        FOREIGN KEY (mahasiswa_id) REFERENCES mahasiswa(id_mahasiswa)
        ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_rekening_bank
        FOREIGN KEY (bank_id_bank) REFERENCES bank(id_bank)
        ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT uq_rekening_bank_no UNIQUE (bank_id_bank, no_rekening)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS merchant (
    id_merchant VARCHAR(20) PRIMARY KEY,
    nama_merchant VARCHAR(120) NOT NULL,
    kategori VARCHAR(80) NOT NULL,
    saldo_merchant DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    status ENUM('ACTIVE', 'INACTIVE') NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT chk_merchant_saldo_non_negative CHECK (saldo_merchant >= 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS transaksi (
    id_transaksi BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    kode_transaksi VARCHAR(50) NOT NULL UNIQUE,
    jenis_transaksi ENUM('TOPUP', 'PAYMENT') NOT NULL,
    nominal DECIMAL(15,2) NOT NULL,
    status ENUM('SUCCESS', 'PENDING', 'FAILED') NOT NULL DEFAULT 'SUCCESS',
    waktu TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    bank_id_bank BIGINT UNSIGNED NULL,
    merchant_id VARCHAR(20) NULL,
    wallet_id_wallet VARCHAR(20) NOT NULL,
    keterangan VARCHAR(255) NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_transaksi_bank
        FOREIGN KEY (bank_id_bank) REFERENCES bank(id_bank)
        ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT fk_transaksi_merchant
        FOREIGN KEY (merchant_id) REFERENCES merchant(id_merchant)
        ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT fk_transaksi_wallet
        FOREIGN KEY (wallet_id_wallet) REFERENCES wallet(id_wallet)
        ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT chk_transaksi_nominal_positive CHECK (nominal > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS audit_logs (
    id_audit BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    transaksi_id BIGINT UNSIGNED NULL,
    action VARCHAR(50) NOT NULL,
    description VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_audit_transaksi
        FOREIGN KEY (transaksi_id) REFERENCES transaksi(id_transaksi)
        ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
