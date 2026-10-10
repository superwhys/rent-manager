-- 给已有租金账单增加生成时的抄表明细快照。
-- 停止服务和其他写入并备份后，在目标数据库执行；新库由 AutoMigrate 建表。
-- MySQL DDL 隐式提交。此迁移只新增可空列，不修改或回填历史账单。
-- NULL 表示历史无快照，JSON [] 表示生成时没有抄表项。
-- 未收款的历史账单可通过重新生成保存当时的抄表快照；已收款或已取消账单保持不变。
-- 若新版服务已通过 AutoMigrate 增加该列，不要再次执行本迁移。

ALTER TABLE rent_bills ADD COLUMN meter_readings JSON NULL AFTER meter_fee;
