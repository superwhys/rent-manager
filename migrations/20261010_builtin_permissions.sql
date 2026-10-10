-- 将旧数据库的权限 ID 绑定迁移为后端固定权限编码，保留权限组、角色和楼栋授权。
-- 仅适用于含 permissions 表及 permission_group_permissions.permission_id 的旧结构。
-- 停止服务和其他写入并备份后执行；新数据库由新版 AutoMigrate 建表，无需本迁移。
-- mysql --default-character-set=utf8mb4 -u <用户名> -p <数据库名> < 本文件
-- MySQL DDL 隐式提交，不能事务回滚；DDL 阶段中断时从备份恢复后重新执行，不要加 --force。
-- 未知编码、孤立绑定或异常外键会在修改业务表之前拒绝，请先人工确认旧配置。
-- 仅迁移启用且未删除的权限绑定，停用/已删除权限的绑定不会重新生效。
-- 迁移按旧表真实 ID 查找编码，不要求权限 ID 固定；不关闭 foreign_key_checks。
-- 预检失败后再次执行前需 DROP PROCEDURE migrate_20261010_builtin_permissions。

DELIMITER $$

CREATE PROCEDURE migrate_20261010_builtin_permissions()
BEGIN
    DECLARE invalid_rows BIGINT DEFAULT 0;
    DECLARE fk_name VARCHAR(64);

    IF DATABASE() IS NULL THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '请先选择需要迁移的数据库。';
    END IF;
    SELECT COUNT(*) INTO invalid_rows
    FROM information_schema.TABLES
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME IN ('permissions', 'permission_group_permissions');
    IF invalid_rows <> 2 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '本迁移需要旧权限表和绑定表；新数据库无需执行。';
    END IF;
    SELECT COUNT(*) INTO invalid_rows
    FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'permission_group_permissions'
      AND COLUMN_NAME = 'permission_id';
    IF invalid_rows <> 1 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '未发现旧权限 ID 列，请人工确认数据库迁移状态。';
    END IF;
    SELECT COUNT(*) INTO invalid_rows
    FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'permission_group_permissions'
      AND COLUMN_NAME = 'permission_code';
    IF invalid_rows <> 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '权限编码列已存在，请检查中断的迁移并从备份恢复。';
    END IF;

    SELECT COUNT(*) INTO invalid_rows
    FROM permission_group_permissions b
    LEFT JOIN permissions p ON p.id = b.permission_id
    LEFT JOIN permission_groups g ON g.id = b.permission_group_id
    WHERE p.id IS NULL OR g.id IS NULL;
    IF invalid_rows > 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '存在孤立权限绑定，请先人工确认旧授权配置。';
    END IF;
    SELECT COUNT(*) INTO invalid_rows
    FROM permission_group_permissions b JOIN permissions p ON p.id = b.permission_id
    WHERE p.code IS NULL OR BINARY p.code NOT IN (
        'unit.view', 'unit.update', 'room.view', 'room.manage', 'room_type.view', 'meter_item.view',
        'meter_reading.view', 'meter_reading.manage', 'rent_bill.view', 'rent_bill.manage'
    );
    IF invalid_rows > 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '存在未知权限编码绑定，请先人工确认旧授权配置。';
    END IF;
    SELECT COUNT(*) INTO invalid_rows
    FROM (
        SELECT b.permission_group_id, p.code COLLATE utf8mb4_bin AS permission_code
        FROM permission_group_permissions b JOIN permissions p ON p.id = b.permission_id
        WHERE p.is_active = 1 AND p.deleted_at IS NULL
        GROUP BY b.permission_group_id, p.code COLLATE utf8mb4_bin HAVING COUNT(*) > 1
    ) duplicated;
    IF invalid_rows > 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '存在重复权限编码绑定，请先人工确认旧授权配置。';
    END IF;
    SELECT COUNT(*) INTO invalid_rows
    FROM information_schema.KEY_COLUMN_USAGE
    WHERE REFERENCED_TABLE_SCHEMA = DATABASE() AND REFERENCED_TABLE_NAME = 'permissions'
      AND NOT (TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'permission_group_permissions'
               AND COLUMN_NAME = 'permission_id' AND REFERENCED_COLUMN_NAME = 'id');
    IF invalid_rows > 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '旧权限表存在其他外键引用，请先人工制定迁移方案。';
    END IF;
    SELECT COUNT(*) INTO invalid_rows
    FROM information_schema.KEY_COLUMN_USAGE
    WHERE REFERENCED_TABLE_SCHEMA = DATABASE() AND REFERENCED_TABLE_NAME = 'permission_group_permissions';
    IF invalid_rows > 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '旧权限绑定表被其他外键引用，请先人工制定迁移方案。';
    END IF;
    SELECT COUNT(*) INTO invalid_rows
    FROM information_schema.KEY_COLUMN_USAGE
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'permission_group_permissions'
      AND COLUMN_NAME = 'permission_id' AND REFERENCED_TABLE_NAME IS NOT NULL
      AND NOT (REFERENCED_TABLE_SCHEMA = DATABASE() AND REFERENCED_TABLE_NAME = 'permissions'
               AND REFERENCED_COLUMN_NAME = 'id');
    IF invalid_rows > 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '旧权限 ID 列存在异常外键，请先人工制定迁移方案。';
    END IF;

    -- 从此处开始修改业务表，DDL 失败不能自动回滚。
    ALTER TABLE permission_group_permissions
        ADD COLUMN permission_code VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NULL;
    UPDATE permission_group_permissions b JOIN permissions p ON p.id = b.permission_id
    SET b.permission_code = p.code
    WHERE p.is_active = 1 AND p.deleted_at IS NULL;
    DELETE FROM permission_group_permissions WHERE permission_code IS NULL;

    -- 使用数据库实际外键名称；权限组外键保持原样。
    SELECT MIN(CONSTRAINT_NAME) INTO fk_name
    FROM information_schema.KEY_COLUMN_USAGE
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'permission_group_permissions'
      AND REFERENCED_TABLE_SCHEMA = DATABASE() AND REFERENCED_TABLE_NAME = 'permissions';
    WHILE fk_name IS NOT NULL DO
        SET @builtin_permission_ddl = CONCAT(
            'ALTER TABLE permission_group_permissions DROP FOREIGN KEY `', REPLACE(fk_name, '`', '``'), '`'
        );
        PREPARE builtin_permission_stmt FROM @builtin_permission_ddl;
        EXECUTE builtin_permission_stmt;
        DEALLOCATE PREPARE builtin_permission_stmt;
        SELECT MIN(CONSTRAINT_NAME) INTO fk_name
        FROM information_schema.KEY_COLUMN_USAGE
        WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'permission_group_permissions'
          AND REFERENCED_TABLE_SCHEMA = DATABASE() AND REFERENCED_TABLE_NAME = 'permissions';
    END WHILE;

    ALTER TABLE permission_group_permissions
        DROP PRIMARY KEY,
        DROP COLUMN permission_id,
        MODIFY COLUMN permission_code VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
        ADD PRIMARY KEY (permission_group_id, permission_code);
    DROP TABLE permissions;
END$$

CALL migrate_20261010_builtin_permissions()$$
DROP PROCEDURE migrate_20261010_builtin_permissions$$

DELIMITER ;
