INSERT INTO bank (id_bank, nama_bank, kode_bank, biaya_admin, is_active)
VALUES
    (1, 'BNI', 'BNI', 1000.00, TRUE),
    (2, 'BRI', 'BRI', 1000.00, TRUE),
    (3, 'Mandiri', 'MDR', 1500.00, TRUE),
    (4, 'BCA', 'BCA', 1500.00, TRUE),
    (5, 'BTN', 'BTN', 1000.00, TRUE),
    (6, 'BSI', 'BSI', 1000.00, TRUE)
ON DUPLICATE KEY UPDATE
    nama_bank = VALUES(nama_bank),
    kode_bank = VALUES(kode_bank),
    biaya_admin = VALUES(biaya_admin),
    is_active = VALUES(is_active);

INSERT INTO merchant (id_merchant, nama_merchant, kategori, saldo_merchant, status)
VALUES
    ('MRC-0001', 'Kantin Teknik', 'Makanan & Minuman', 0.00, 'ACTIVE'),
    ('MRC-0002', 'Fotokopi Kampus', 'Percetakan', 0.00, 'ACTIVE'),
    ('MRC-0003', 'Parkir Kampus', 'Transportasi', 0.00, 'ACTIVE'),
    ('MRC-0004', 'Koperasi Mahasiswa', 'Retail', 0.00, 'ACTIVE'),
    ('MRC-0005', 'Event Kampus', 'Acara', 0.00, 'ACTIVE')
ON DUPLICATE KEY UPDATE
    nama_merchant = VALUES(nama_merchant),
    kategori = VALUES(kategori),
    status = VALUES(status);

INSERT INTO mahasiswa (id_mahasiswa, nrp, nama_mahasiswa, email, role, status)
VALUES
    (1, '5025201001', 'Alya Putri', 'alya@itspay.test', 'mahasiswa', 'ACTIVE'),
    (2, '0000000000', 'Admin ITSPay', 'admin@itspay.test', 'admin', 'ACTIVE')
ON DUPLICATE KEY UPDATE
    nrp = VALUES(nrp),
    nama_mahasiswa = VALUES(nama_mahasiswa),
    email = VALUES(email),
    role = VALUES(role),
    status = VALUES(status);

INSERT IGNORE INTO wallet (id_wallet, mahasiswa_id, jenis_wallet, saldo)
VALUES
    ('WLT-0001', 1, 'REGULAR', 0.00),
    ('WLT-0002', 2, 'ADMIN', 0.00);

INSERT INTO mahasiswa_auth (mahasiswa_id, password_hash)
VALUES
    (1, '$2a$10$Z4J3DAqyJDrvE8qHl1xWvOAdOM.R5DySr0rLF0Ql3oASaOypUciny'),
    (2, '$2a$10$Z4J3DAqyJDrvE8qHl1xWvOAdOM.R5DySr0rLF0Ql3oASaOypUciny')
ON DUPLICATE KEY UPDATE
    password_hash = VALUES(password_hash);

INSERT INTO rekening_mahasiswa (mahasiswa_id, bank_id_bank, no_rekening, nama_pemilik, is_active)
VALUES
    (1, 1, '8800123456', 'Alya Putri', TRUE)
ON DUPLICATE KEY UPDATE
    nama_pemilik = VALUES(nama_pemilik),
    is_active = VALUES(is_active);
