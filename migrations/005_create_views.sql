CREATE OR REPLACE VIEW v_riwayat_transaksi_mahasiswa AS
SELECT
    m.id_mahasiswa,
    m.nrp,
    m.nama_mahasiswa,
    w.id_wallet,
    t.id_transaksi,
    t.kode_transaksi,
    t.jenis_transaksi,
    t.nominal,
    t.status,
    t.waktu,
    t.bank_id_bank,
    COALESCE(t.merchant_id, '') AS merchant_id,
    t.wallet_id_wallet,
    COALESCE(t.keterangan, '') AS keterangan,
    COALESCE(b.nama_bank, '') AS bank_name,
    COALESCE(mc.nama_merchant, '') AS merchant_name
FROM transaksi t
INNER JOIN wallet w ON w.id_wallet = t.wallet_id_wallet
INNER JOIN mahasiswa m ON m.id_mahasiswa = w.mahasiswa_id
LEFT JOIN bank b ON b.id_bank = t.bank_id_bank
LEFT JOIN merchant mc ON mc.id_merchant = t.merchant_id;

CREATE OR REPLACE VIEW v_laporan_transaksi_harian AS
SELECT
    DATE(waktu) AS tanggal,
    COUNT(*) AS total_transaksi,
    COALESCE(SUM(CASE WHEN status = 'SUCCESS' THEN nominal ELSE 0 END), 0.00) AS total_nominal,
    COALESCE(SUM(CASE WHEN jenis_transaksi = 'TOPUP' AND status = 'SUCCESS' THEN nominal ELSE 0 END), 0.00) AS total_topup,
    COALESCE(SUM(CASE WHEN jenis_transaksi = 'PAYMENT' AND status = 'SUCCESS' THEN nominal ELSE 0 END), 0.00) AS total_payment,
    SUM(CASE WHEN status = 'SUCCESS' THEN 1 ELSE 0 END) AS total_transaksi_success
FROM transaksi
GROUP BY DATE(waktu);

