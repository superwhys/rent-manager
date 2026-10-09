-- 仅适用于当前旧版结构，且旧 accounts、账号所属表及 meter_readings 必须全部为空。
-- 旧整数 ID 不能直接转换为 authkit 账号，本迁移不复制账号、不建立映射、不清空数据。
-- 停止服务和其他写入后，使用 mysql --default-character-set=utf8mb4 <数据库名> < 本文件执行。
-- MySQL DDL 会隐式提交；先备份，不能依赖事务回滚。新数据库无需执行本迁移。
-- authkit 的表由新版服务调用 authkit AutoMigrate 创建，本文件不维护其内部结构。
-- 预检失败不会修改业务表；再次执行前需 DROP PROCEDURE migrate_20261009_authkit_accounts。

DELIMITER $$

CREATE PROCEDURE migrate_20261009_authkit_accounts()
BEGIN
    DECLARE existing_rows BIGINT DEFAULT 0;

    SELECT
        (SELECT COUNT(*) FROM accounts) +
        (SELECT COUNT(*) FROM account_settings) +
        (SELECT COUNT(*) FROM units) +
        (SELECT COUNT(*) FROM roles) +
        (SELECT COUNT(*) FROM meter_items) +
        (SELECT COUNT(*) FROM room_types) +
        (SELECT COUNT(*) FROM unit_account_roles) +
        (SELECT COUNT(*) FROM meter_readings)
    INTO existing_rows;

    IF existing_rows > 0 THEN
        SIGNAL SQLSTATE '45000'
            SET MESSAGE_TEXT = '旧账号或所属业务表存在数据，请使用新数据库；本迁移不执行账号映射。';
    END IF;

    ALTER TABLE account_settings
        DROP FOREIGN KEY account_settings_accounts_account_settings,
        MODIFY COLUMN account_id VARCHAR(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL;

    ALTER TABLE units
        DROP FOREIGN KEY units_accounts_units,
        MODIFY COLUMN account_id VARCHAR(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL;

    ALTER TABLE roles
        DROP FOREIGN KEY roles_accounts_roles,
        MODIFY COLUMN account_id VARCHAR(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL;

    ALTER TABLE meter_items
        DROP FOREIGN KEY meter_items_accounts_meter_items,
        MODIFY COLUMN account_id VARCHAR(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL;

    ALTER TABLE room_types
        DROP FOREIGN KEY room_types_accounts_room_types,
        MODIFY COLUMN account_id VARCHAR(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL;

    ALTER TABLE unit_account_roles
        DROP FOREIGN KEY unit_account_roles_accounts_unit_account_roles,
        DROP FOREIGN KEY unit_account_roles_accounts_granted_unit_account_roles,
        MODIFY COLUMN account_id VARCHAR(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
        MODIFY COLUMN granted_by_account_id VARCHAR(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL;

    ALTER TABLE meter_readings
        MODIFY COLUMN reader_id VARCHAR(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NULL;

    DROP TABLE accounts;
END$$

DELIMITER ;

CALL migrate_20261009_authkit_accounts();
DROP PROCEDURE migrate_20261009_authkit_accounts;
