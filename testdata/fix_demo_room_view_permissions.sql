-- 为已导入的演示抄表、财务权限组补齐房间页面的目录查看权限。
-- 先完成 migrations/20261010_builtin_permissions.sql（旧库），再在演示测试库执行。
-- 可重复执行，仅添加固定权限编码的缺失绑定；停用、删除或缺失权限组不修改。

INSERT INTO permission_group_permissions (permission_group_id, permission_code)
SELECT g.id, p.code
FROM permission_groups g CROSS JOIN (
    SELECT 'room_type.view' AS code UNION ALL SELECT 'meter_item.view'
) p
WHERE g.code IN ('rent_demo.readings', 'rent_demo.finance')
  AND g.is_active = 1 AND g.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM permission_group_permissions existing
      WHERE existing.permission_group_id = g.id AND existing.permission_code = p.code
  );
