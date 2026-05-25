DELIMITER $$

DROP FUNCTION IF EXISTS fn_hitung_total_topup$$

CREATE FUNCTION fn_hitung_total_topup(p_mahasiswa_id BIGINT UNSIGNED)
RETURNS DECIMAL(15,2)
READS SQL DATA
BEGIN
    DECLARE v_total DECIMAL(15,2);

    SELECT COALESCE(SUM(t.nominal), 0.00)
    INTO v_total
    FROM transaksi t
    INNER JOIN wallet w ON w.id_wallet = t.wallet_id_wallet
    WHERE w.mahasiswa_id = p_mahasiswa_id
      AND t.jenis_transaksi = 'TOPUP'
      AND t.status = 'SUCCESS';

    RETURN v_total;
END$$

DELIMITER ;

