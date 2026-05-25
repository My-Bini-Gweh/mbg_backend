DELIMITER $$

DROP TRIGGER IF EXISTS trg_mahasiswa_after_insert_create_wallet$$

CREATE TRIGGER trg_mahasiswa_after_insert_create_wallet
AFTER INSERT ON mahasiswa
FOR EACH ROW
BEGIN
    INSERT INTO wallet (id_wallet, mahasiswa_id, jenis_wallet, saldo)
    VALUES (
        CONCAT('WLT-', LPAD(NEW.id_mahasiswa, 4, '0')),
        NEW.id_mahasiswa,
        'REGULAR',
        0.00
    );
END$$

DROP TRIGGER IF EXISTS trg_wallet_before_insert_prevent_negative$$

CREATE TRIGGER trg_wallet_before_insert_prevent_negative
BEFORE INSERT ON wallet
FOR EACH ROW
BEGIN
    IF NEW.saldo < 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Saldo wallet tidak boleh negatif';
    END IF;
END$$

DROP TRIGGER IF EXISTS trg_wallet_before_update_prevent_negative$$

CREATE TRIGGER trg_wallet_before_update_prevent_negative
BEFORE UPDATE ON wallet
FOR EACH ROW
BEGIN
    IF NEW.saldo < 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Saldo wallet tidak boleh negatif';
    END IF;
END$$

DROP TRIGGER IF EXISTS trg_transaksi_after_insert_audit_log$$

CREATE TRIGGER trg_transaksi_after_insert_audit_log
AFTER INSERT ON transaksi
FOR EACH ROW
BEGIN
    DECLARE v_action VARCHAR(50);
    DECLARE v_description VARCHAR(255);

    SET v_action = CASE
        WHEN NEW.jenis_transaksi = 'TOPUP' AND NEW.status = 'SUCCESS' THEN 'TOPUP_SUCCESS'
        WHEN NEW.jenis_transaksi = 'PAYMENT' AND NEW.status = 'SUCCESS' THEN 'PAYMENT_SUCCESS'
        WHEN NEW.jenis_transaksi = 'PAYMENT' AND NEW.status = 'FAILED' THEN 'PAYMENT_FAILED'
        ELSE CONCAT(NEW.jenis_transaksi, '_', NEW.status)
    END;

    SET v_description = CONCAT(
        'Transaksi ', NEW.kode_transaksi,
        ' ', LOWER(NEW.status),
        ' dengan nominal ', CAST(NEW.nominal AS CHAR)
    );

    INSERT INTO audit_logs (transaksi_id, action, description)
    VALUES (NEW.id_transaksi, v_action, v_description);
END$$

DELIMITER ;

