import { request } from '@/api/request'
import type {
  BindingAction,
  GrantsQuery,
  Permission,
  PermissionGroup,
  PermissionGroupInput,
  Role,
  RoleInput,
  UnitGrant,
} from '@/types/access'

export const listRoles = () => request<Role[]>('/roles')
export const createRole = (body: RoleInput) =>
  request<Role[]>('/roles/batch', { method: 'POST', body: { roles: [body] } })
export const updateRole = (id: number, body: RoleInput) =>
  request<Role>(`/roles/${id}`, { method: 'PUT', body })
export const deleteRole = (id: number) =>
  request<void>(`/roles/${id}`, { method: 'DELETE' })

export const listPermissionGroups = () =>
  request<PermissionGroup[]>('/admin/permission-groups')
export const listPermissions = () => request<Permission[]>('/admin/permissions')
export const createPermissionGroup = (body: PermissionGroupInput) =>
  request<PermissionGroup[]>('/admin/permission-groups/batch', {
    method: 'POST',
    body: { permission_groups: [body] },
  })
export const changeRoleGroups = (
  id: number,
  ids: number[],
  action: BindingAction,
) =>
  request<void>(`/roles/${id}/permission-groups`, {
    method: action === 'add' ? 'POST' : 'DELETE',
    body: { permission_group_ids: ids },
  })
export const changeGroupPermissions = (
  id: number,
  codes: string[],
  action: BindingAction,
) =>
  request<void>(`/admin/permission-groups/${id}/permissions`, {
    method: action === 'add' ? 'POST' : 'DELETE',
    body: { permission_codes: codes },
  })

export const listUnitGrants = (id: number, query: GrantsQuery = {}) =>
  request<UnitGrant[]>(`/units/${id}/grants`, { query: { ...query } })
export const grantUnitRoles = (
  unitId: number,
  accountId: string,
  roleIds: number[],
) =>
  request<UnitGrant[]>(`/units/${unitId}/accounts/${accountId}/roles`, {
    method: 'POST',
    body: { role_ids: roleIds },
  })
export const revokeUnitRoles = (
  unitId: number,
  accountId: string,
  roleIds: number[],
) =>
  request<void>(`/units/${unitId}/accounts/${accountId}/roles`, {
    method: 'DELETE',
    body: { role_ids: roleIds },
  })
