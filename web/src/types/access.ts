export interface Role {
  id: number
  account_id: string
  name: string
  is_default_owner: boolean
  is_active: boolean
  permission_group_ids: number[]
}

export interface RoleInput {
  name: string
  is_default_owner: boolean
  is_active?: boolean
}

export interface PermissionGroup {
  id: number
  code: string
  name: string
  description?: string | null
  sort_order: number
  is_active: boolean
  permission_codes: string[]
}

export interface Permission {
  code: string
  description: string
}

export interface PermissionGroupInput {
  code: string
  name: string
  description: string
  sort_order: number
}

export interface UnitGrant {
  id: number
  unit_id: number
  account_id: string
  role_id: number
  granted_by_account_id: string
  status: string
  role?: Role | null
}

export interface GrantsQuery {
  account_id?: string
  status?: string
  limit?: number
  offset?: number
}

export type BindingAction = 'add' | 'remove'
