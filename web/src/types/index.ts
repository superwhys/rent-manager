// 月份选项类型
export interface MonthOption {
  value: number
  label: string
}

// 年份选项类型
export interface YearOption {
  value: number
  label: string
}

// 物业地址选项类型
export interface PropertyOption {
  value: number
  address: string
  label: string
}

export interface DashboardCard {
  name: string
  value: number
}

export interface DashboardData {
  // 房间总数
  totalRooms: DashboardCard
  // 租客总数
  totalTenants: DashboardCard
  // 账单总数
  totalBills: DashboardCard
  // 水电费总数
  totalUtilities: DashboardCard
  // 报表总数
  totalReports: DashboardCard
  // 系统设置总数
  totalSettings: DashboardCard
}

// 房间状态类型
export type RoomStatus = 'occupied' | 'vacant' | 'pending'

// 房间信息接口
export interface RoomInfo {
  id: string
  status: RoomStatus
  rent: number
  tenant: string
  // 费用配置
  feeConfig?: {
    baseRent: number
    networkFee: number
    elevatorFee: number
    otherFee: number
    waterPrice: number
    waterMinimum: number
    electricPrice: number
    electricMinimum: number
    gasPrice: number
    gasMinimum: number
  }
  // 抄表记录
  utilityReading?: {
    date: string
    lastWaterReading: number
    lastElectricReading: number
    lastGasReading: number
    waterReading: number
    electricReading: number
    gasReading: number
    waterUsage: number
    electricUsage: number
    gasUsage: number
  }
  // 账单记录
  bill?: {
    id: string
    rent: number
    waterFee: number
    electricFee: number
    gasFee: number
    otherFee: number
    amount: number
    status: BillStatus
    createDate: string
  }
}

// 扩展房间管理数据类型
export interface Room {
  number: string
  floor: number
  area: number
  rent: number
  tenant: string | null
  status: RoomStatus
  moveInDate: string | null
}

// 租客信息类型
export interface Tenant {
  name: string
  phone: string
  idCard: string
  room: string
  deposit: number
  moveInDate: string
  contractStatus: 'active' | 'pending' | 'expired'
}

// 账单状态类型
export type BillStatus = 'paid' | 'unpaid'

// 账单信息类型
export interface Bill {
  id: string
  room: string
  tenant: string
  rent: number
  waterFee: number
  electricFee: number
  gasFee: number
  otherFee: number
  amount: number
  status: BillStatus
  createDate: string
}

// 账单统计数据类型
export interface BillStats {
  totalAmount: DashboardCard
  paidAmount: DashboardCard
  unpaidAmount: DashboardCard
  paymentRate: DashboardCard
}

// 水电表读数类型
export interface UtilityReading {
  room: string
  date: string
  waterReading: number
  electricReading: number
  gasReading: number
  waterUsage: number
  electricUsage: number
  gasUsage: number
}

// 水电费用设置类型
export interface UtilitySettings {
  waterPrice: number
  waterMinimum: number
  electricPrice: number
  electricMinimum: number
  gasPrice: number
  gasMinimum: number
}

// 系统设置类型
export interface SystemSettings {
  billGenerateDate: number
  overdueReminderDays: number
  autoBackup: 'daily' | 'weekly' | 'monthly'
}

// 图表数据类型
export interface ChartData {
  label: string
  value: number
}

// 报表数据类型
export interface ReportData {
  incomeChart: ChartData[]
  occupancyChart: ChartData[]
  utilityChart: ChartData[]
  yearComparison: {
    currentYear: number
    lastYear: number
    growthRate: string
  }
}

// 表格操作按钮类型
export interface TableAction {
  label: string
  type: 'primary' | 'success' | 'warning' | 'danger'
  handler: () => void
}

// 状态标签配置类型
export interface StatusConfig {
  [key: string]: {
    text: string
    class: string
  }
}

// 房间费用详情类型
export interface RoomFeeDetails {
  roomId: string
  baseRent: number
  networkFee: number
  elevatorFee: number
  otherFee: number
  lastMonthWater: number
  currentMonthWater: number
  waterPrice: number
  waterMinimum: number
  lastMonthElectric: number
  currentMonthElectric: number
  electricPrice: number
  electricMinimum: number
  lastMonthGas: number
  currentMonthGas: number
  gasPrice: number
  gasMinimum: number
}

// 费用计算结果类型
export interface FeeCalculationResult {
  waterTotal: number
  electricTotal: number
  gasTotal: number
  grandTotal: number
}

// 用户角色类型
export type UserRole = 'landlord' | 'employee'

// 用户状态类型
export type UserStatus = 'active' | 'inactive'

// 用户信息接口
export interface User {
  id: string
  username: string
  email: string
  phone: string
  role: UserRole
  status: UserStatus
  realName: string
  createDate: string
  lastLoginDate: string | null
  managedProperties: number[] // 允许管理的物业ID列表（主房东可管理所有）
  permissions: string[] // 权限列表
}

// 登录表单接口
export interface LoginForm {
  username: string
  password: string
}

// 创建用户表单接口
export interface CreateUserForm {
  username: string
  email: string
  phone: string
  realName: string
  password: string
  confirmPassword: string
  role: UserRole
  managedProperties: number[] // 分配给员工的物业权限
}

// 用户管理统计数据类型
export interface UserStats {
  totalUsers: DashboardCard
  activeUsers: DashboardCard
  landlords: DashboardCard
  employees: DashboardCard
}

// 认证状态接口
export interface AuthState {
  isAuthenticated: boolean
  user: User | null
  token: string | null
}

// 权限配置接口
export interface Permission {
  key: string
  name: string
  description: string
  category: string
}

// 权限检查选项接口
export interface PermissionCheckOptions {
  requireAll?: boolean // 是否需要全部权限，默认false（只需要其中一个）
  fallback?: any // 无权限时的备用内容
}

// 权限模块常量
export const PermissionModule = {
  DASHBOARD: 'dashboard',
  PROPERTIES: 'properties',
  ROOMS: 'rooms',
  TENANTS: 'tenants',
  BILLS: 'bills',
  UTILITIES: 'utilities',
  REPORTS: 'reports',
  USERS: 'users',
  SETTINGS: 'settings'
} as const

// 权限操作常量
export const PermissionAction = {
  VIEW: 'view',
  CREATE: 'create',
  EDIT: 'edit',
  DELETE: 'delete',
  EXPORT: 'export',
  IMPORT: 'import'
} as const

// 类型定义
export type PermissionModuleType = typeof PermissionModule[keyof typeof PermissionModule]
export type PermissionActionType = typeof PermissionAction[keyof typeof PermissionAction]

// 物业状态类型
export type PropertyStatus = 'active' | 'inactive' | 'maintenance'

// 物业详细信息接口
export interface Property {
  id: number
  name: string
  address: string
  totalRooms: number
  occupiedRooms: number
  monthlyIncome: number
  status: PropertyStatus
  createDate: string
  defaultRent: number
  depositMultiple: number
  networkFee: number
  elevatorFee: number
}

// 创建物业表单接口
export interface CreatePropertyForm {
  name: string
  address: string
  defaultRent: number
  depositMultiple: number
  networkFee: number
  elevatorFee: number
}

// 物业统计数据类型
export interface PropertyStats {
  totalProperties: DashboardCard
  activeProperties: DashboardCard
  totalIncome: DashboardCard
  occupancyRate: DashboardCard
} 