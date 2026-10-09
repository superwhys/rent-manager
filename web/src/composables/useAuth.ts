import { ref, computed } from 'vue'
import type { User, LoginForm, AuthState, Permission, PermissionCheckOptions, PermissionModuleType, PermissionActionType } from '../types'

/**
 * 用户认证状态管理composable
 * 管理用户登录状态、用户信息、权限控制等
 */

// 系统权限列表（测试阶段写死，生产环境应从后端获取）
const systemPermissions: Permission[] = [
  // 仪表板权限
  { key: 'dashboard.view', name: '查看仪表板', description: '查看系统概览和统计数据', category: '仪表板' },
  
  // 物业管理权限
  { key: 'properties.view', name: '查看物业', description: '查看物业列表和详情', category: '物业管理' },
  { key: 'properties.create', name: '创建物业', description: '添加新的物业信息', category: '物业管理' },
  { key: 'properties.edit', name: '编辑物业', description: '修改物业信息', category: '物业管理' },
  { key: 'properties.delete', name: '删除物业', description: '删除物业记录', category: '物业管理' },
  
  // 房间管理权限
  { key: 'rooms.view', name: '查看房间', description: '查看房间列表和详情', category: '房间管理' },
  { key: 'rooms.create', name: '创建房间', description: '添加新的房间信息', category: '房间管理' },
  { key: 'rooms.edit', name: '编辑房间', description: '修改房间信息', category: '房间管理' },
  { key: 'rooms.delete', name: '删除房间', description: '删除房间记录', category: '房间管理' },
  
  // 租客管理权限
  { key: 'tenants.view', name: '查看租客', description: '查看租客列表和详情', category: '租客管理' },
  { key: 'tenants.create', name: '创建租客', description: '添加新的租客信息', category: '租客管理' },
  { key: 'tenants.edit', name: '编辑租客', description: '修改租客信息', category: '租客管理' },
  { key: 'tenants.delete', name: '删除租客', description: '删除租客记录', category: '租客管理' },
  
  // 账单管理权限
  { key: 'bills.view', name: '查看账单', description: '查看账单列表和详情', category: '账单管理' },
  { key: 'bills.create', name: '创建账单', description: '生成新的账单', category: '账单管理' },
  { key: 'bills.edit', name: '编辑账单', description: '修改账单信息', category: '账单管理' },
  { key: 'bills.delete', name: '删除账单', description: '删除账单记录', category: '账单管理' },
  { key: 'bills.export', name: '导出账单', description: '导出账单数据', category: '账单管理' },
  
  // 水电管理权限
  { key: 'utilities.view', name: '查看水电', description: '查看水电使用情况', category: '水电管理' },
  { key: 'utilities.create', name: '录入读数', description: '录入水电表读数', category: '水电管理' },
  { key: 'utilities.edit', name: '编辑读数', description: '修改水电表读数', category: '水电管理' },
  { key: 'utilities.delete', name: '删除读数', description: '删除水电表读数', category: '水电管理' },
  
  // 报表分析权限
  { key: 'reports.view', name: '查看报表', description: '查看各类统计报表', category: '报表分析' },
  { key: 'reports.export', name: '导出报表', description: '导出报表数据', category: '报表分析' },
  
  // 用户管理权限（仅主房东）
  { key: 'users.view', name: '查看用户', description: '查看用户列表和详情', category: '用户管理' },
  { key: 'users.create', name: '创建用户', description: '添加新的系统用户', category: '用户管理' },
  { key: 'users.edit', name: '编辑用户', description: '修改用户信息和权限', category: '用户管理' },
  { key: 'users.delete', name: '删除用户', description: '删除系统用户', category: '用户管理' },
  
  // 系统设置权限
  { key: 'settings.view', name: '查看设置', description: '查看系统设置', category: '系统设置' },
  { key: 'settings.edit', name: '修改设置', description: '修改系统配置', category: '系统设置' },
]

// 全局认证状态
const authState = ref<AuthState>({
  isAuthenticated: false,
  user: null,
  token: null
})

// 模拟用户数据（实际项目中应该从后端获取）
const mockUsers: User[] = [
  {
    id: '1',
    username: 'landlord001',
    email: 'landlord@example.com',
    phone: '13800138000',
    role: 'landlord',
    status: 'active',
    realName: '张主房东',
    createDate: '2024-01-01',
    lastLoginDate: null,
    managedProperties: [1, 2, 3], // 主房东可以管理所有物业
    permissions: ['all'] // 主房东拥有所有权限
  },
  {
    id: '2',
    username: 'employee001',
    email: 'employee@example.com',
    phone: '13800138001',
    role: 'employee',
    status: 'active',
    realName: '李管理员',
    createDate: '2024-01-15',
    lastLoginDate: null,
    managedProperties: [1, 2], // 管理员可以管理多个物业
    permissions: [
      // 仪表板权限
      'dashboard.view',
      
      // 物业管理权限（部分）
      'properties.view', 'properties.edit',
      
      // 房间管理权限（全部）
      'rooms.view', 'rooms.create', 'rooms.edit', 'rooms.delete',
      
      // 租客管理权限（全部）
      'tenants.view', 'tenants.create', 'tenants.edit', 'tenants.delete',
      
      // 账单管理权限（全部）
      'bills.view', 'bills.create', 'bills.edit', 'bills.delete', 'bills.export',
      
      // 水电管理权限（全部）
      'utilities.view', 'utilities.create', 'utilities.edit', 'utilities.delete',
      
      // 报表权限（查看和导出）
      'reports.view', 'reports.export',
      
      // 设置权限（仅查看）
      'settings.view'
    ]
  },
  {
    id: '3',
    username: 'employee002',
    email: 'employee2@example.com',
    phone: '13800138002',
    role: 'employee',
    status: 'active',
    realName: '王只读员工',
    createDate: '2024-02-01',
    lastLoginDate: null,
    managedProperties: [2], // 只读员工只能查看指定物业
    permissions: [
      // 仪表板权限
      'dashboard.view',
      
      // 物业管理权限（仅查看）
      'properties.view',
      
      // 房间管理权限（仅查看）
      'rooms.view',
      
      // 租客管理权限（仅查看）
      'tenants.view', 
      
      // 账单管理权限（仅查看）
      'bills.view',
      
      // 水电管理权限（仅查看）
      'utilities.view',
      
      // 设置权限（仅查看）
      'settings.view'
    ]
  },
  {
    id: '4',
    username: 'employee003',
    email: 'employee3@example.com',
    phone: '13800138003',
    role: 'employee',
    status: 'active',
    realName: '赵财务员工',
    createDate: '2024-02-15',
    lastLoginDate: null,
    managedProperties: [1, 2, 3], // 财务可以查看所有物业
    permissions: [
      // 仪表板权限
      'dashboard.view',
      
      // 房间管理权限（仅查看）
      'rooms.view',
      
      // 租客管理权限（仅查看）
      'tenants.view',
      
      // 账单管理权限（全部）
      'bills.view', 'bills.create', 'bills.edit', 'bills.export',
      
      // 水电管理权限（录入和查看）
      'utilities.view', 'utilities.create', 'utilities.edit',
      
      // 报表权限（全部）
      'reports.view', 'reports.export'
    ]
  }
]

export function useAuth() {
  /**
   * 用户登录
   */
  const login = async (loginForm: LoginForm): Promise<{ success: boolean; message: string }> => {
    try {
      // 模拟登录验证（实际项目中应该调用后端API）
      const user = mockUsers.find(u => u.username === loginForm.username)

      if (!user) {
        return { success: false, message: '用户名不存在' }
      }

      // 简单的密码验证（实际项目中应该使用加密验证）
      if (loginForm.password !== '123456') {
        return { success: false, message: '密码错误' }
      }

      if (user.status !== 'active') {
        return { success: false, message: '用户已被禁用' }
      }

      // 更新认证状态
      const token = `token_${user.id}_${Date.now()}`
      authState.value = {
        isAuthenticated: true,
        user: { ...user, lastLoginDate: new Date().toISOString() },
        token
      }

      // 保存到localStorage
      localStorage.setItem('auth_token', token)
      localStorage.setItem('user_info', JSON.stringify(user))

      return { success: true, message: '登录成功' }
    } catch (error) {
      return { success: false, message: '登录失败，请重试' }
    }
  }

  /**
   * 用户登出
   */
  const logout = () => {
    authState.value = {
      isAuthenticated: false,
      user: null,
      token: null
    }

    // 清除localStorage
    localStorage.removeItem('auth_token')
    localStorage.removeItem('user_info')
  }

  /**
   * 检查用户权限
   */
  const hasPermission = (permission: string): boolean => {
    if (!authState.value.user) return false

    // 主房东拥有所有权限
    if (authState.value.user.role === 'landlord') return true

    // 检查用户权限列表
    return authState.value.user.permissions.includes(permission) ||
      authState.value.user.permissions.includes('all')
  }

  /**
   * 检查多个权限
   */
  const hasPermissions = (permissions: string[], options: PermissionCheckOptions = {}): boolean => {
    if (!authState.value.user) return false
    
    // 主房东拥有所有权限
    if (authState.value.user.role === 'landlord') return true

    const { requireAll = false } = options

    if (requireAll) {
      // 需要拥有所有权限
      return permissions.every(permission => hasPermission(permission))
    } else {
      // 只需要拥有其中一个权限
      return permissions.some(permission => hasPermission(permission))
    }
  }

  /**
   * 检查模块权限
   */
  const hasModulePermission = (module: PermissionModuleType, action: PermissionActionType): boolean => {
    const permissionKey = `${module}.${action}`
    return hasPermission(permissionKey)
  }

  /**
   * 获取用户拥有的权限列表
   */
  const getUserPermissions = (): string[] => {
    if (!authState.value.user) return []
    
    // 主房东拥有所有权限
    if (authState.value.user.role === 'landlord') {
      return systemPermissions.map(p => p.key)
    }

    return authState.value.user.permissions
  }

  /**
   * 获取系统所有权限
   */
  const getAllPermissions = (): Permission[] => {
    return systemPermissions
  }

  /**
   * 根据分类获取权限
   */
  const getPermissionsByCategory = (category: string): Permission[] => {
    return systemPermissions.filter(p => p.category === category)
  }

  /**
   * 检查用户是否可以管理特定物业
   */
  const canManageProperty = (propertyId: number): boolean => {
    if (!authState.value.user) return false

    // 主房东可以管理所有物业
    if (authState.value.user.role === 'landlord') return true

    // 检查员工的物业权限
    return authState.value.user.managedProperties.includes(propertyId)
  }

  /**
   * 初始化认证状态
   */
  const initAuth = () => {
    const token = localStorage.getItem('auth_token')
    const userInfo = localStorage.getItem('user_info')

    if (token && userInfo) {
      try {
        const user = JSON.parse(userInfo)
        authState.value = {
          isAuthenticated: true,
          user,
          token
        }
      } catch (error) {
        console.error('Failed to parse user info from localStorage:', error)
        logout()
      }
    }
  }

  // 计算属性
  const isAuthenticated = computed(() => authState.value.isAuthenticated)
  const currentUser = computed(() => authState.value.user)
  const isLandlord = computed(() => authState.value.user?.role === 'landlord')
  const isEmployee = computed(() => authState.value.user?.role === 'employee')

  return {
    // 状态
    authState: computed(() => authState.value),
    isAuthenticated,
    currentUser,
    isLandlord,
    isEmployee,

    // 权限方法
    hasPermission,
    hasPermissions,
    hasModulePermission,
    getUserPermissions,
    getAllPermissions,
    getPermissionsByCategory,
    canManageProperty,

    // 认证方法
    login,
    logout,
    initAuth
  }
}