DELIMITER $$

DROP PROCEDURE IF EXISTS sp_topup_wallet$$

CREATE PROCEDURE sp_topup_wallet(
    IN p_wallet_id VARCHAR(20),
    IN p_bank_id BIGINT UNSIGNED,
    IN p_nominal DECIMAL(15,2)
)
BEGIN
    DECLARE v_wallet_exists INT DEFAULT 0;
    DECLARE v_bank_active INT DEFAULT 0;
    DECLARE v_kode_transaksi VARCHAR(50);

    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;

    IF p_wallet_id IS NULL OR p_wallet_id = '' THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Wallet wajib diisi';
    END IF;

    IF p_nominal IS NULL OR p_nominal <= 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Nominal top up harus lebih dari 0';
    END IF;

    SELECT COUNT(*)
    INTO v_wallet_exists
    FROM wallet
    WHERE id_wallet = p_wallet_id;

    IF v_wallet_exists = 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Wallet tidak ditemukan';
    END IF;

    SELECT COUNT(*)
    INTO v_bank_active
    FROM bank
    WHERE id_bank = p_bank_id
      AND is_active = TRUE;

    IF v_bank_active = 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Bank tidak ditemukan atau tidak aktif';
    END IF;

    START TRANSACTION;

    UPDATE wallet
    SET saldo = saldo + p_nominal
    WHERE id_wallet = p_wallet_id;

    SET v_kode_transaksi = CONCAT(
        'TRX-', DATE_FORMAT(NOW(), '%Y%m%d'), '-',
        UPPER(SUBSTRING(REPLACE(UUID(), '-', ''), 1, 8))
    );

    INSERT INTO transaksi (
        kode_transaksi,
        jenis_transaksi,
        nominal,
        status,
        bank_id_bank,
        merchant_id,
        wallet_id_wallet,
        keterangan
    ) VALUES (
        v_kode_transaksi,
        'TOPUP',
        p_nominal,
        'SUCCESS',
        p_bank_id,
        NULL,
        p_wallet_id,
        'Top up saldo dari bank mitra'
    );

    COMMIT;
END$$

DROP PROCEDURE IF EXISTS sp_bayar_merchant$$

CREATE PROCEDURE sp_bayar_merchant(
    IN p_wallet_id VARCHAR(20),
    IN p_merchant_id VARCHAR(20),
    IN p_nominal DECIMAL(15,2)
)
BEGIN
    DECLARE v_wallet_exists INT DEFAULT 0;
    DECLARE v_merchant_active INT DEFAULT 0;
    DECLARE v_saldo DECIMAL(15,2) DEFAULT 0.00;
    DECLARE v_kode_transaksi VARCHAR(50);

    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;

    IF p_wallet_id IS NULL OR p_wallet_id = '' THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Wallet wajib diisi';
    END IF;

    IF p_merchant_id IS NULL OR p_merchant_id = '' THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Merchant wajib diisi';
    END IF;

    IF p_nominal IS NULL OR p_nominal <= 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Nominal pembayaran harus lebih dari 0';
    END IF;

    SELECT COUNT(*)
    INTO v_wallet_exists
    FROM wallet
    WHERE id_wallet = p_wallet_id;

    IF v_wallet_exists = 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Wallet tidak ditemukan';
    END IF;

    SELECT COUNT(*)
    INTO v_merchant_active
    FROM merchant
    WHERE id_merchant = p_merchant_id
      AND status = 'ACTIVE';

    IF v_merchant_active = 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Merchant tidak ditemukan atau tidak aktif';
    END IF;

    START TRANSACTION;

    SELECT saldo
    INTO v_saldo
    FROM wallet
    WHERE id_wallet = p_wallet_id
    FOR UPDATE;

    SET v_kode_transaksi = CONCAT(
        'TRX-', DATE_FORMAT(NOW(), '%Y%m%d'), '-',
        UPPER(SUBSTRING(REPLACE(UUID(), '-', ''), 1, 8))
    );

    IF v_saldo < p_nominal THEN
        INSERT INTO transaksi (
            kode_transaksi,
            jenis_transaksi,
            nominal,
            status,
            bank_id_bank,
            merchant_id,
            wallet_id_wallet,
            keterangan
        ) VALUES (
            v_kode_transaksi,
            'PAYMENT',
            p_nominal,
            'FAILED',
            NULL,
            p_merchant_id,
            p_wallet_id,
            'Pembayaran gagal: saldo tidak mencukupi'
        );

        COMMIT;
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Saldo tidak mencukupi';
    END IF;

    UPDATE wallet
    SET saldo = saldo - p_nominal
    WHERE id_wallet = p_wallet_id;

    UPDATE merchant
    SET saldo_merchant = saldo_merchant + p_nominal
    WHERE id_merchant = p_merchant_id;

    INSERT INTO transaksi (
        kode_transaksi,
        jenis_transaksi,
        nominal,
        status,
        bank_id_bank,
        merchant_id,
        wallet_id_wallet,
        keterangan
    ) VALUES (
        v_kode_transaksi,
        'PAYMENT',
        p_nominal,
        'SUCCESS',
        NULL,
        p_merchant_id,
        p_wallet_id,
        'Pembayaran merchant kampus'
    );

    COMMIT;
END$$

DELIMITER ;

