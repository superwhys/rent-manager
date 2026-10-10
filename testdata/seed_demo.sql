-- Rent Manager 演示数据（当前业务模型 + authkit v1.9.0，MySQL 8.0+）。
-- 先启动一次当前版本服务完成 AutoMigrate，再选择目标测试数据库执行本文件。
-- mysql --default-character-set=utf8mb4 -u <用户名> -p <数据库名> < testdata/seed_demo.sql
-- 需要目标库 INSERT/SELECT 和 CREATE ROUTINE/ALTER ROUTINE 权限。
-- 一次性导入；演示账号/邮箱/权限组已存在时会报错并停止，不覆盖已有数据。
-- 所有数据写入在同一事务内；本脚本不关闭外键、不清表、不修改已有账号。
-- 5 个演示账号的共同密码：RentDemo2026!（仅用于测试环境）。
-- 邮箱：owner / manager / reader / finance / other-owner @rent-demo.test。
-- owner 是业务楼栋所有者；如需全局权限组写操作，另将其 ID 加入
-- config 中 admin.SuperAdminAccountIDs（SQL 不会改变服务配置）。
-- 金额单位为分；unit_price 为分/计量单位；is_active=0 已租，1 空房。

SET NAMES utf8mb4;
SET time_zone = '+08:00';

DELIMITER $$

DROP PROCEDURE IF EXISTS seed_20261010_rent_manager_demo$$
CREATE PROCEDURE seed_20261010_rent_manager_demo()
BEGIN
    DECLARE conflict_count BIGINT DEFAULT 0;
    DECLARE seed_now DATETIME DEFAULT NOW();
    DECLARE auth_now DATETIME(6) DEFAULT UTC_TIMESTAMP(6);
    DECLARE month_start DATE;
    DECLARE owner_id VARCHAR(64);
    DECLARE other_owner_id VARCHAR(64);

    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;

    IF DATABASE() IS NULL THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '请先选择已完成当前版本迁移的测试数据库。';
    END IF;
    SET month_start = CAST(DATE_FORMAT(CURRENT_DATE(), '%Y-%m-01') AS DATE);

    DROP TEMPORARY TABLE IF EXISTS tmp_rent_demo_accounts;
    CREATE TEMPORARY TABLE tmp_rent_demo_accounts (
        seed_key INT PRIMARY KEY, account_id VARCHAR(64) COLLATE utf8mb4_bin,
        email VARCHAR(254), username VARCHAR(64), password_hash VARCHAR(255)
    );
    INSERT INTO tmp_rent_demo_accounts VALUES
        (1, SHA2('rent-manager-demo-v1:owner', 256), 'owner@rent-demo.test', '演示主账号',
         '$argon2id$v=19$m=19456,t=2,p=1$IVj5VXut+0Wey8shx7itLw$Ihm2LXckrm0/E30L1kOIMo9Fu2lHqg3y0L/zdgvSwxE'),
        (2, SHA2('rent-manager-demo-v1:manager', 256), 'manager@rent-demo.test', '演示管家',
         '$argon2id$v=19$m=19456,t=2,p=1$i8O3Ga1mzz6ayu/OLzqS9g$+fys5O+lbwHwHblcpCXolh6CSvCnSq7NO0Q4wwTn/oI'),
        (3, SHA2('rent-manager-demo-v1:reader', 256), 'reader@rent-demo.test', '演示抄表员',
         '$argon2id$v=19$m=19456,t=2,p=1$jGDQrU5trBoEV5PNyvVd8Q$xvLFOIt7nEOhjXE/wmKDaRjLCWZmH8wct64tAEIZjgw'),
        (4, SHA2('rent-manager-demo-v1:finance', 256), 'finance@rent-demo.test', '演示财务',
         '$argon2id$v=19$m=19456,t=2,p=1$V4R1Ak1iGiCulOYR8+4sDA$gzebj2gKglxABPltML8jagAW57n78XEaJ2cycoHXkio'),
        (5, SHA2('rent-manager-demo-v1:other-owner', 256), 'other-owner@rent-demo.test', '演示独立业主',
         '$argon2id$v=19$m=19456,t=2,p=1$Phrs0j52jugotgz88yRDEw$vUpmTlDcOWyji9VAGkbAbIf/nVSA3waRXniNnm845mk');

    SELECT account_id INTO owner_id FROM tmp_rent_demo_accounts WHERE seed_key = 1;
    SELECT account_id INTO other_owner_id FROM tmp_rent_demo_accounts WHERE seed_key = 5;
    SELECT COUNT(*) INTO conflict_count
    FROM auth_accounts a JOIN tmp_rent_demo_accounts d ON a.id = d.account_id;
    IF conflict_count > 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '演示账号已存在，未写入数据；本脚本只导入一次。';
    END IF;
    SELECT COUNT(*) INTO conflict_count
    FROM auth_bindings b JOIN tmp_rent_demo_accounts d ON b.identifier = d.email;
    IF conflict_count > 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '演示邮箱已有认证绑定，未写入数据，请使用其他测试库。';
    END IF;
    SELECT
        (SELECT COUNT(*) FROM units WHERE account_id IN (owner_id, other_owner_id)) +
        (SELECT COUNT(*) FROM roles WHERE account_id IN (owner_id, other_owner_id)) +
        (SELECT COUNT(*) FROM room_types WHERE account_id IN (owner_id, other_owner_id)) +
        (SELECT COUNT(*) FROM meter_items WHERE account_id IN (owner_id, other_owner_id)) +
        (SELECT COUNT(*) FROM account_settings s JOIN tmp_rent_demo_accounts d ON s.account_id = d.account_id) +
        (SELECT COUNT(*) FROM permission_groups WHERE code LIKE 'rent_demo.%')
    INTO conflict_count;
    IF conflict_count > 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '存在冲突的演示业务数据或权限组，未写入数据。';
    END IF;

    START TRANSACTION;

    -- 账号采用双绑定，密码登录可直接使用，账号响应可正常显示邮箱。
    INSERT INTO auth_accounts (id, username, created_at, updated_at)
    SELECT account_id, username, auth_now, auth_now FROM tmp_rent_demo_accounts;
    INSERT INTO auth_bindings (method, identifier, account_id, password_hash, created_at, updated_at)
    SELECT 'email', email, account_id, NULL, auth_now, auth_now FROM tmp_rent_demo_accounts;
    INSERT INTO auth_bindings (method, identifier, account_id, password_hash, created_at, updated_at)
    SELECT 'password', email, account_id, password_hash, auth_now, auth_now FROM tmp_rent_demo_accounts;

    -- 权限目录由后端固定定义；这里只列出演示权限组需要绑定的编码。
    DROP TEMPORARY TABLE IF EXISTS tmp_rent_demo_permissions;
    CREATE TEMPORARY TABLE tmp_rent_demo_permissions (
        code VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin PRIMARY KEY
    );
    INSERT INTO tmp_rent_demo_permissions VALUES
        ('unit.view'), ('unit.update'), ('room.view'), ('room.manage'),
        ('room_type.view'), ('meter_item.view'),
        ('meter_reading.view'), ('meter_reading.manage'), ('rent_bill.view'), ('rent_bill.manage');

    INSERT INTO permission_groups (code, name, description, sort_order, is_active, created_at, updated_at) VALUES
        ('rent_demo.operations', '演示·楼栋运营', '所有当前楼栋业务权限，共享定义仍由业主管理', 100, 1, seed_now, seed_now),
        ('rent_demo.readings', '演示·抄表协作', '查看楼栋和房间，录入抄表，不可查看账单', 110, 1, seed_now, seed_now),
        ('rent_demo.finance', '演示·财务协作', '查看楼栋和房间，管理账单，不可修改房间或抄表', 120, 1, seed_now, seed_now),
        ('rent_demo.readonly', '演示·只读查看', '查看业务资料，不可修改', 130, 1, seed_now, seed_now);
    INSERT INTO permission_group_permissions (permission_group_id, permission_code)
    SELECT g.id, p.code FROM permission_groups g CROSS JOIN tmp_rent_demo_permissions p
    WHERE g.code = 'rent_demo.operations'
       OR (g.code = 'rent_demo.readings' AND p.code IN
           ('unit.view', 'room.view', 'room_type.view', 'meter_item.view', 'meter_reading.view', 'meter_reading.manage'))
       OR (g.code = 'rent_demo.finance' AND p.code IN
           ('unit.view', 'room.view', 'room_type.view', 'meter_item.view', 'rent_bill.view', 'rent_bill.manage'))
       OR (g.code = 'rent_demo.readonly' AND p.code IN
           ('unit.view', 'room.view', 'room_type.view', 'meter_item.view', 'meter_reading.view', 'rent_bill.view'));

    DROP TEMPORARY TABLE IF EXISTS tmp_rent_demo_roles;
    CREATE TEMPORARY TABLE tmp_rent_demo_roles (
        seed_key INT PRIMARY KEY, name VARCHAR(255), group_code VARCHAR(255), is_active BOOLEAN, role_id BIGINT
    );
    INSERT INTO tmp_rent_demo_roles VALUES
        (1, '演示·管家', 'rent_demo.operations', 1, NULL),
        (2, '演示·抄表员', 'rent_demo.readings', 1, NULL),
        (3, '演示·财务', 'rent_demo.finance', 1, NULL),
        (4, '演示·只读', 'rent_demo.readonly', 1, NULL),
        (5, '演示·仅入口', NULL, 1, NULL),
        (6, '演示·已停用', 'rent_demo.operations', 0, NULL);
    INSERT INTO roles (name, is_default_owner, is_active, account_id, created_at, updated_at)
    SELECT name, 0, is_active, owner_id, seed_now, seed_now FROM tmp_rent_demo_roles;
    UPDATE tmp_rent_demo_roles d JOIN roles r ON r.account_id = owner_id AND r.name = d.name SET d.role_id = r.id;
    INSERT INTO role_permission_groups (role_id, permission_group_id)
    SELECT r.role_id, g.id FROM tmp_rent_demo_roles r JOIN permission_groups g ON g.code = r.group_code;

    -- 两个所有者各自的共享定义；未绑定的燃气项目可用于测试下级选用完整目录。
    DROP TEMPORARY TABLE IF EXISTS tmp_rent_demo_room_types;
    CREATE TEMPORARY TABLE tmp_rent_demo_room_types (
        seed_key INT PRIMARY KEY, account_id VARCHAR(64) COLLATE utf8mb4_bin, name VARCHAR(12), room_type_id BIGINT
    );
    INSERT INTO tmp_rent_demo_room_types VALUES
        (1, owner_id, '演示单间', NULL), (2, owner_id, '演示一居', NULL), (3, owner_id, '演示两居', NULL),
        (4, other_owner_id, '演示单间', NULL), (5, other_owner_id, '演示一居', NULL);
    INSERT INTO room_types (name, account_id, created_at, updated_at)
    SELECT name, account_id, seed_now, seed_now FROM tmp_rent_demo_room_types;
    UPDATE tmp_rent_demo_room_types d JOIN room_types r ON r.account_id = d.account_id AND r.name = d.name
    SET d.room_type_id = r.id;

    DROP TEMPORARY TABLE IF EXISTS tmp_rent_demo_meter_items;
    CREATE TEMPORARY TABLE tmp_rent_demo_meter_items (
        seed_key INT PRIMARY KEY, account_id VARCHAR(64) COLLATE utf8mb4_bin, code VARCHAR(255),
        unit VARCHAR(255), unit_price DOUBLE, calc_type VARCHAR(16), meter_item_id BIGINT
    );
    INSERT INTO tmp_rent_demo_meter_items VALUES
        (1, owner_id, '演示水费', '吨', 450, 'DIFFERENCE', NULL),
        (2, owner_id, '演示电费', '度', 85, 'DIFFERENCE', NULL),
        (3, owner_id, '演示物业费', '月', 5000, 'DIRECT', NULL),
        (4, owner_id, '演示燃气费', '立方米', 320, 'DIFFERENCE', NULL),
        (5, other_owner_id, '演示水费', '吨', 500, 'DIFFERENCE', NULL),
        (6, other_owner_id, '演示电费', '度', 80, 'DIFFERENCE', NULL),
        (7, other_owner_id, '演示物业费', '月', 3000, 'DIRECT', NULL);
    INSERT INTO meter_items (code, unit, unit_price, is_system, calc_type, remark, account_id, created_at, updated_at)
    SELECT code, unit, unit_price, 0, calc_type, '演示数据：单价单位为分', account_id, seed_now, seed_now
    FROM tmp_rent_demo_meter_items;
    UPDATE tmp_rent_demo_meter_items d JOIN meter_items m ON m.account_id = d.account_id AND m.code = d.code
    SET d.meter_item_id = m.id;

    DROP TEMPORARY TABLE IF EXISTS tmp_rent_demo_units;
    CREATE TEMPORARY TABLE tmp_rent_demo_units (
        seed_key INT PRIMARY KEY, account_id VARCHAR(64) COLLATE utf8mb4_bin, name VARCHAR(255),
        address VARCHAR(255), base_price BIGINT, unit_id BIGINT
    );
    INSERT INTO tmp_rent_demo_units VALUES
        (1, owner_id, '演示·梧桐公寓', '南山区演示路 101 号', 200000, NULL),
        (2, owner_id, '演示·海棠公寓', '福田区演示路 202 号', 220000, NULL),
        (3, owner_id, '演示·银杏公寓', '宝安区演示路 303 号', 180000, NULL),
        (4, other_owner_id, '演示·独立业主公寓', '龙华区演示路 404 号', 160000, NULL);
    INSERT INTO units (name, address, city, province, json_extend, account_id, created_at, updated_at)
    SELECT name, address, '深圳市', '广东省', '{"meter_reading_day":1}', account_id, seed_now, seed_now
    FROM tmp_rent_demo_units;
    UPDATE tmp_rent_demo_units d JOIN units u ON u.account_id = d.account_id AND u.name = d.name SET d.unit_id = u.id;
    INSERT INTO unit_meter_items (unit_id, meter_item_id)
    SELECT u.unit_id, m.meter_item_id FROM tmp_rent_demo_units u
    JOIN tmp_rent_demo_meter_items m ON m.account_id = u.account_id WHERE m.seed_key <> 4;

    -- 每栋 6 间：101/102/201/202 已租，301/302 空房。租金在 1600~3200 元之间。
    DROP TEMPORARY TABLE IF EXISTS tmp_rent_demo_room_slots;
    CREATE TEMPORARY TABLE tmp_rent_demo_room_slots (
        slot_key INT PRIMARY KEY, room_no VARCHAR(255), floor INT, price_delta BIGINT, size INT, is_active BOOLEAN
    );
    INSERT INTO tmp_rent_demo_room_slots VALUES
        (1, '101', 1, 0, 22, 0), (2, '102', 1, 20000, 28, 0),
        (3, '201', 2, 80000, 42, 0), (4, '202', 2, 100000, 48, 0),
        (5, '301', 3, 10000, 25, 1), (6, '302', 3, 60000, 38, 1);
    INSERT INTO rooms (room_no, floor, price, size, is_active, room_type_id, unit_id, created_at, updated_at)
    SELECT s.room_no, s.floor, u.base_price + s.price_delta, s.size, s.is_active, t.room_type_id, u.unit_id, seed_now, seed_now
    FROM tmp_rent_demo_units u CROSS JOIN tmp_rent_demo_room_slots s
    JOIN tmp_rent_demo_room_types t ON t.seed_key = CASE
        WHEN u.seed_key = 4 THEN IF(s.slot_key IN (1, 2, 5), 4, 5)
        WHEN s.slot_key IN (1, 2, 5) THEN 1 WHEN s.slot_key IN (3, 6) THEN 2 ELSE 3 END;
    DROP TEMPORARY TABLE IF EXISTS tmp_rent_demo_rooms;
    CREATE TEMPORARY TABLE tmp_rent_demo_rooms AS
    SELECT (u.seed_key - 1) * 6 + s.slot_key AS seed_key, u.seed_key AS unit_key, s.slot_key,
           r.id AS room_id, r.unit_id, r.price, r.is_active
    FROM tmp_rent_demo_units u CROSS JOIN tmp_rent_demo_room_slots s
    JOIN rooms r ON r.unit_id = u.unit_id AND r.room_no = s.room_no;
    INSERT INTO room_meter_items (room_id, meter_item_id)
    SELECT r.room_id, m.meter_item_id FROM tmp_rent_demo_rooms r
    JOIN unit_meter_items m ON m.unit_id = r.unit_id;

    -- 按楼栋授予 owner 创建的角色；另一业主的楼栋未共享，用于验证账号隔离。
    DROP TEMPORARY TABLE IF EXISTS tmp_rent_demo_grants;
    CREATE TEMPORARY TABLE tmp_rent_demo_grants (unit_key INT, account_key INT, role_key INT, status VARCHAR(16));
    INSERT INTO tmp_rent_demo_grants VALUES
        (1, 2, 1, 'active'), (2, 2, 4, 'active'), (3, 2, 1, 'disabled'),
        (1, 3, 2, 'active'), (2, 3, 5, 'active'), (3, 3, 6, 'active'),
        (1, 4, 3, 'active'), (2, 4, 3, 'active');
    INSERT INTO unit_account_roles (unit_id, account_id, role_id, granted_by_account_id, status, created_at, updated_at)
    SELECT u.unit_id, a.account_id, r.role_id, owner_id, d.status, seed_now, seed_now
    FROM tmp_rent_demo_grants d JOIN tmp_rent_demo_units u ON u.seed_key = d.unit_key
    JOIN tmp_rent_demo_accounts a ON a.seed_key = d.account_key JOIN tmp_rent_demo_roles r ON r.seed_key = d.role_key;

    INSERT INTO account_settings (`key`, value, account_id, created_at, updated_at)
    SELECT 'demo.note', '演示数据，金额按分存储', account_id, seed_now, seed_now FROM tmp_rent_demo_accounts;
    INSERT INTO account_settings (`key`, value, account_id, created_at, updated_at)
    SELECT 'demo.scenario', username, account_id, seed_now, seed_now FROM tmp_rent_demo_accounts;

    -- 上上月/上月/本月的水电读数连续；上月梧桐102故意缺物业读数，其他房间均完整。
    DROP TEMPORARY TABLE IF EXISTS tmp_rent_demo_periods;
    CREATE TEMPORARY TABLE tmp_rent_demo_periods (period_key INT PRIMARY KEY, first_day DATE, period_month VARCHAR(7));
    INSERT INTO tmp_rent_demo_periods VALUES
        (0, DATE_SUB(month_start, INTERVAL 2 MONTH), DATE_FORMAT(DATE_SUB(month_start, INTERVAL 2 MONTH), '%Y-%m')),
        (1, DATE_SUB(month_start, INTERVAL 1 MONTH), DATE_FORMAT(DATE_SUB(month_start, INTERVAL 1 MONTH), '%Y-%m')),
        (2, month_start, DATE_FORMAT(month_start, '%Y-%m'));
    INSERT INTO meter_readings
        (room_id, meter_item_id, period_month, previous_reading, current_reading, usage_amount,
         reading_time, reader_id, source, remark, created_at, updated_at)
    SELECT r.room_id, m.id, p.period_month,
        CASE
            WHEN m.calc_type = 'DIRECT' THEN NULL
            WHEN m.unit = '吨' THEN 100 + r.seed_key * 2 + p.period_key * (8 + MOD(r.seed_key, 5))
                 + p.period_key * (p.period_key - 1) / 2
            ELSE 500 + r.seed_key * 10 + p.period_key * (60 + r.seed_key * 3)
                 + 5 * p.period_key * (p.period_key - 1) / 2 END,
        CASE
            WHEN m.calc_type = 'DIRECT' THEN 1
            WHEN m.unit = '吨' THEN 100 + r.seed_key * 2 + (p.period_key + 1) * (8 + MOD(r.seed_key, 5))
                 + p.period_key * (p.period_key + 1) / 2
            ELSE 500 + r.seed_key * 10 + (p.period_key + 1) * (60 + r.seed_key * 3)
                 + 5 * p.period_key * (p.period_key + 1) / 2 END,
        CASE WHEN m.calc_type = 'DIRECT' THEN 1 WHEN m.unit = '吨' THEN 8 + MOD(r.seed_key, 5) + p.period_key
             ELSE 60 + r.seed_key * 3 + p.period_key * 5 END,
        CAST(p.first_day AS DATETIME),
        (SELECT account_id FROM tmp_rent_demo_accounts WHERE seed_key = IF(r.unit_key = 1, 3, IF(r.unit_key = 4, 5, 1))),
        'MANUAL', '演示月度读数', seed_now, seed_now
    FROM tmp_rent_demo_rooms r CROSS JOIN tmp_rent_demo_periods p
    JOIN room_meter_items rm ON rm.room_id = r.room_id JOIN meter_items m ON m.id = rm.meter_item_id
    WHERE r.is_active = 0 AND NOT (p.period_key = 1 AND r.unit_key = 1 AND r.slot_key = 2 AND m.calc_type = 'DIRECT');

    -- 上月梧桐201读数完整但未出账，102未抄表且未出账；其余存在已付/未付/取消账单。
    -- 金额从真实演示读数逐项 ROUND 计算，避免账单与抄表数据不一致。
    INSERT INTO rent_bills
        (room_id, period_month, base_rent, meter_fee, total_amount, status, paid_at, remark, created_at, updated_at)
    SELECT r.room_id, p.period_month, r.price, SUM(ROUND(mr.usage_amount * m.unit_price)),
        r.price + SUM(ROUND(mr.usage_amount * m.unit_price)),
        CASE WHEN p.period_key = 0 OR r.slot_key = 1 THEN 'PAID'
             WHEN r.slot_key = 4 AND (p.period_key = 2 OR r.unit_key <> 1) THEN 'CANCELLED'
             WHEN p.period_key = 1 AND r.slot_key = 3 THEN 'PAID' ELSE 'UNPAID' END,
        CASE WHEN p.period_key = 0 OR r.slot_key = 1 OR (p.period_key = 1 AND r.slot_key = 3)
             THEN CAST(p.first_day AS DATETIME) ELSE NULL END,
        '演示月租账单：含水电及物业费', seed_now, seed_now
    FROM tmp_rent_demo_rooms r CROSS JOIN tmp_rent_demo_periods p
    JOIN room_meter_items rm ON rm.room_id = r.room_id JOIN meter_items m ON m.id = rm.meter_item_id
    JOIN meter_readings mr ON mr.room_id = r.room_id AND mr.meter_item_id = m.id AND mr.period_month = p.period_month
    WHERE r.is_active = 0 AND NOT (p.period_key = 1 AND r.unit_key = 1 AND r.slot_key IN (2, 3))
    GROUP BY r.room_id, p.period_month, r.price, p.period_key, r.slot_key, r.unit_key, p.first_day;

    COMMIT;

    -- 导入结果与可登录账号；密码仅在文件说明中提供。
    SELECT username AS `账号`, email AS `登录邮箱`, account_id AS `账号ID` FROM tmp_rent_demo_accounts ORDER BY seed_key;
    SELECT u.name AS `楼栋`, u.account_id AS `所有者ID`, u.unit_id AS `楼栋ID`,
           COUNT(r.room_id) AS `房间数`, SUM(r.is_active = 0) AS `已租`, SUM(r.is_active = 1) AS `空房`
    FROM tmp_rent_demo_units u JOIN tmp_rent_demo_rooms r ON r.unit_id = u.unit_id GROUP BY u.seed_key, u.name, u.account_id, u.unit_id;
    SELECT p.period_month AS `月份`, b.status AS `账单状态`, COUNT(*) AS `数量`, SUM(b.total_amount) / 100 AS `金额_元`
    FROM tmp_rent_demo_periods p JOIN rent_bills b ON b.period_month = p.period_month
    JOIN tmp_rent_demo_rooms r ON r.room_id = b.room_id GROUP BY p.period_month, b.status ORDER BY p.period_month, b.status;

    DROP TEMPORARY TABLE tmp_rent_demo_periods, tmp_rent_demo_grants, tmp_rent_demo_rooms,
        tmp_rent_demo_room_slots, tmp_rent_demo_units, tmp_rent_demo_meter_items, tmp_rent_demo_room_types,
        tmp_rent_demo_roles, tmp_rent_demo_permissions, tmp_rent_demo_accounts;
END$$

CALL seed_20261010_rent_manager_demo()$$
DROP PROCEDURE seed_20261010_rent_manager_demo$$

DELIMITER ;
