CREATE INDEX idx_mahasiswa_email ON mahasiswa(email);
CREATE INDEX idx_mahasiswa_nrp ON mahasiswa(nrp);
CREATE INDEX idx_wallet_mahasiswa ON wallet(mahasiswa_id);
CREATE INDEX idx_bank_active ON bank(is_active);
CREATE INDEX idx_merchant_status ON merchant(status);
CREATE INDEX idx_transaksi_wallet_waktu ON transaksi(wallet_id_wallet, waktu);
CREATE INDEX idx_transaksi_status_waktu ON transaksi(status, waktu);
CREATE INDEX idx_transaksi_jenis_waktu ON transaksi(jenis_transaksi, waktu);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);

