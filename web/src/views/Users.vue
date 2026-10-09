<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { User, CreateUserForm, UserStats, PropertyOption } from '../types'
import { useAuth } from '../composables/useAuth'
import UserForm from '../components/UserForm.vue'
import StatCard from '../components/StatCard.vue'
import StatusBadge from '../components/StatusBadge.vue'
import DataTable from '../components/DataTable.vue'

/**
 * 用户管理页面
 * 主房东可以管理员工账户，查看用户统计信息
 */

const { isLandlord, hasPermission } = useAuth()

// 页面状态
const userFormVisible = ref(false)
const currentEditUser = ref<User | null>(null)
const isEditMode = ref(false)

// 搜索筛选
const searchForm = reactive({
  keyword: '',
  role: '',
  status: ''
})

// 模拟用户数据（实际项目中从后端获取）
const allUsers = ref<User[]>([
  {
    id: '1',
    username: 'landlord001',
    email: 'landlord@example.com',
    phone: '13800138000',
    role: 'landlord',
    status: 'active',
    realName: '张主房东',
    createDate: '2024-01-01',
    lastLoginDate: '2024-12-20',
    managedProperties: [1, 2, 3],
    permissions: ['all']
  },
  {
    id: '2',
    username: 'employee001',
    email: 'employee001@example.com',
    phone: '13800138001',
    role: 'employee',
    status: 'active',
    realName: '李管理员',
    createDate: '2024-01-15',
    lastLoginDate: '2024-12-19',
    managedProperties: [1],
    permissions: ['rooms', 'tenants', 'bills']
  },
  {
    id: '3',
    username: 'employee002',
    email: 'employee002@example.com',
    phone: '13800138002',
    role: 'employee',
    status: 'inactive',
    realName: '王助理',
    createDate: '2024-02-01',
    lastLoginDate: null,
    managedProperties: [2, 3],
    permissions: ['rooms', 'tenants']
  }
])

// 物业选项（从App.vue获取，实际项目中应该从API获取）
const propertyOptions = ref<PropertyOption[]>([
  {
    value: 1,
    address: '广东省深圳市光明区马田街道福华东八巷三号',
    label: '深圳光明区 - 福华东八巷三号'
  },
  {
    value: 2,
    address: '广东省深圳市南山区科技园南区深南大道',
    label: '深圳南山区 - 科技园南区'
  },
  {
    value: 3,
    address: '广东省深圳市福田区华强北路赛格广场',
    label: '深圳福田区 - 华强北赛格广场'
  }
])

// 筛选后的用户列表
const filteredUsers = computed(() => {
  return allUsers.value.filter(user => {
    const keywordMatch = !searchForm.keyword ||
      user.realName.includes(searchForm.keyword) ||
      user.username.includes(searchForm.keyword) ||
      user.phone.includes(searchForm.keyword)

    const roleMatch = !searchForm.role || user.role === searchForm.role
    const statusMatch = !searchForm.status || user.status === searchForm.status

    return keywordMatch && roleMatch && statusMatch
  })
})

// 用户统计数据
const userStats = computed((): UserStats => {
  const total = allUsers.value.length
  const active = allUsers.value.filter(u => u.status === 'active').length
  const landlords = allUsers.value.filter(u => u.role === 'landlord').length
  const employees = allUsers.value.filter(u => u.role === 'employee').length

  return {
    totalUsers: { name: '总用户数', value: total },
    activeUsers: { name: '活跃用户', value: active },
    landlords: { name: '主房东', value: landlords },
    employees: { name: '员工账户', value: employees }
  }
})



// 表格列配置
const columns = [
  { key: 'realName', label: '姓名', width: '100px' },
  { key: 'username', label: '用户名', ellipsis: true },
  { key: 'phone', label: '手机号', ellipsis: true },
  { key: 'role', label: '角色' },
  { key: 'status', label: '状态' },
  { key: 'managedProperties', label: '管理物业', width: '250px' },
  { key: 'lastLoginDate', label: '最后登录', ellipsis: true },
  { key: 'createDate', label: '创建时间', ellipsis: true }
]

// 状态配置
const statusConfig = {
  active: { text: '正常', class: 'active' },
  inactive: { text: '禁用', class: 'inactive' }
}

const roleConfig = {
  landlord: { text: '主房东', class: 'landlord' },
  employee: { text: '员工', class: 'employee' }
}

/**
 * 获取物业名称
 */
const getPropertyNames = (propertyIds: number[]) => {
  return propertyIds
    .map(id => propertyOptions.value.find(p => p.value === id)?.label)
    .filter(Boolean)
    .join(', ')
}

/**
 * 格式化日期
 */
const formatDate = (dateString: string | null) => {
  if (!dateString) return '从未登录'
  return new Date(dateString).toLocaleDateString('zh-CN')
}

/**
 * 新建用户
 */
const handleCreateUser = () => {
  currentEditUser.value = null
  isEditMode.value = false
  userFormVisible.value = true
}

/**
 * 编辑用户
 */
const handleEdit = (user: User) => {
  currentEditUser.value = user
  isEditMode.value = true
  userFormVisible.value = true
}

/**
 * 删除用户
 */
const handleDelete = async (user: User) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除用户 "${user.realName}" 吗？此操作不可恢复。`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    // 模拟删除操作（实际项目中调用后端API）
    const index = allUsers.value.findIndex(u => u.id === user.id)
    if (index > -1) {
      allUsers.value.splice(index, 1)
      ElMessage.success('用户删除成功')
    }
  } catch {
    // 用户取消删除
  }
}

/**
 * 切换用户状态
 */
const handleToggleUserStatus = async (user: User) => {
  const newStatus = user.status === 'active' ? 'inactive' : 'active'
  const action = newStatus === 'active' ? '启用' : '禁用'

  try {
    await ElMessageBox.confirm(
      `确定要${action}用户 "${user.realName}" 吗？`,
      `确认${action}`,
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    // 模拟状态切换（实际项目中调用后端API）
    user.status = newStatus
    ElMessage.success(`用户${action}成功`)
  } catch {
    // 用户取消操作
  }
}

/**
 * 重置密码
 */
const handleResetPassword = async (user: User) => {
  try {
    await ElMessageBox.confirm(
      `确定要重置用户 "${user.realName}" 的密码吗？新密码将为：123456`,
      '确认重置密码',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    // 模拟重置密码（实际项目中调用后端API）
    ElMessage.success('密码重置成功，新密码：123456')
  } catch {
    // 用户取消操作
  }
}

/**
 * 用户表单确认
 */
const handleUserFormConfirm = (userData: CreateUserForm, isEdit: boolean, editUserId?: string) => {
  if (isEdit && editUserId) {
    // 编辑模式
    const user = allUsers.value.find(u => u.id === editUserId)
    if (user) {
      Object.assign(user, {
        email: userData.email,
        phone: userData.phone,
        realName: userData.realName,
        role: userData.role,
        managedProperties: userData.managedProperties
      })
    }
  } else {
    // 创建模式
    const newUser: User = {
      id: Date.now().toString(),
      username: userData.username,
      email: userData.email,
      phone: userData.phone,
      role: userData.role,
      status: 'active',
      realName: userData.realName,
      createDate: new Date().toISOString().split('T')[0],
      lastLoginDate: null,
      managedProperties: userData.managedProperties,
      permissions: userData.role === 'landlord' ? ['all'] : ['rooms', 'tenants', 'bills']
    }
    allUsers.value.push(newUser)
  }

  userFormVisible.value = false
}

/**
 * 重置搜索
 */
const handleResetSearch = () => {
  searchForm.keyword = ''
  searchForm.role = ''
  searchForm.status = ''
}

// 页面加载时的初始化
onMounted(() => {
  // 检查权限
  if (!isLandlord.value && !hasPermission('users')) {
    ElMessage.error('您没有权限访问此页面')
    return
  }
})
</script>

<template>
  <div class="users-container">
    <!-- 统计卡片 -->
    <div class="stats-section">
      <StatCard :name="userStats.totalUsers.name" :value="userStats.totalUsers.value"
        background-color="linear-gradient(135deg, #11998e 0%, #38ef7d 100%)" />
      <StatCard :name="userStats.activeUsers.name" :value="userStats.activeUsers.value"
        background-color="linear-gradient(135deg, #667eea 0%, #764ba2 100%)" />
      <StatCard :name="userStats.landlords.name" :value="userStats.landlords.value"
        background-color="linear-gradient(135deg, #fc4a1a 0%, #f7b733 100%)" />
      <StatCard :name="userStats.employees.name" :value="userStats.employees.value"
        background-color="linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)" />
    </div>

    <!-- 页面头部和操作区域 -->
    <div class="section-header">
      <h2 class="section-title">👤 用户管理</h2>
      <button v-if="isLandlord" class="btn btn-primary" @click="handleCreateUser">
        ➕ 新建用户
      </button>
    </div>

    <!-- 搜索筛选区域 -->
    <div class="search-section">
      <div class="search-form">
        <el-input v-model="searchForm.keyword" placeholder="搜索用户名、姓名或手机号" clearable style="width: 300px" />
        <el-select v-model="searchForm.role" placeholder="角色" clearable style="width: 120px">
          <el-option label="主房东" value="landlord" />
          <el-option label="员工" value="employee" />
        </el-select>
        <el-select v-model="searchForm.status" placeholder="状态" clearable style="width: 120px">
          <el-option label="正常" value="active" />
          <el-option label="禁用" value="inactive" />
        </el-select>
        <button class="btn btn-secondary" @click="handleResetSearch">重置</button>
      </div>
    </div>

    <!-- 用户表格 -->
    <DataTable :columns="columns" :data="filteredUsers" :show-actions="true" @edit="handleEdit" @delete="handleDelete">
      <template #role="{ value }">
        <StatusBadge :status="value" :statusConfig="roleConfig" />
      </template>

      <template #status="{ value }">
        <StatusBadge :status="value" :statusConfig="statusConfig" />
      </template>

      <template #managedProperties="{ value }">
        <el-tooltip :content="getPropertyNames(value)" placement="top">
          <span class="property-text">
            {{ getPropertyNames(value) }}
          </span>
        </el-tooltip>
      </template>

      <template #lastLoginDate="{ value }">
        {{ formatDate(value) }}
      </template>

      <template #createDate="{ value }">
        {{ formatDate(value) }}
      </template>

      <!-- 自定义操作按钮 -->
      <template #actions="{ item }">
        <div class="action-buttons">
          <button v-if="isLandlord" class="btn btn-sm btn-primary" @click="handleEdit(item)">
            编辑
          </button>
          <button v-if="isLandlord && item.role !== 'landlord'" class="btn btn-sm"
            :class="item.status === 'active' ? 'btn-warning' : 'btn-success'" @click="handleToggleUserStatus(item)">
            {{ item.status === 'active' ? '禁用' : '启用' }}
          </button>
          <el-dropdown v-if="isLandlord" trigger="click">
            <button class="btn btn-sm btn-secondary">
              更多 ▼
            </button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="handleResetPassword(item)">
                  重置密码
                </el-dropdown-item>
                <el-dropdown-item v-if="item.role !== 'landlord'" @click="handleDelete(item)">
                  删除用户
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </template>
    </DataTable>

    <!-- 用户表单弹窗 -->
    <UserForm v-model:visible="userFormVisible" :user="currentEditUser" :property-options="propertyOptions"
      :is-edit="isEditMode" @confirm="handleUserFormConfirm" />
  </div>
</template>

<style scoped>
.users-container {
  padding: 20px;
}

.section-title {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 20px;
  color: #333;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
  flex-wrap: wrap;
  gap: 15px;
}

.stats-section {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
  margin-bottom: 30px;
}

.search-section {
  margin-bottom: 30px;
}

.search-form {
  display: flex;
  gap: 15px;
  align-items: center;
  flex-wrap: wrap;
  padding: 20px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 12px;
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.1);
}

.btn {
  padding: 12px 24px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.btn-sm {
  padding: 6px 12px;
  font-size: 12px;
}

.btn-primary {
  background: #4f46e5;
  color: white;
}

.btn-primary:hover {
  background: #4338ca;
  transform: translateY(-2px);
}

.btn-secondary {
  background: #6b7280;
  color: white;
}

.btn-secondary:hover {
  background: #4b5563;
  transform: translateY(-2px);
}

.btn-warning {
  background: #f59e0b;
  color: white;
}

.btn-warning:hover {
  background: #d97706;
  transform: translateY(-2px);
}

.btn-success {
  background: #10b981;
  color: white;
}

.btn-success:hover {
  background: #059669;
  transform: translateY(-2px);
}

.action-buttons {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}

.property-text {
  display: inline-block;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .search-form {
    flex-direction: column;
    align-items: stretch;
  }

  .search-form>* {
    width: 100% !important;
  }

  .stats-section {
    grid-template-columns: repeat(2, 1fr);
  }

  .section-header {
    flex-direction: column;
    align-items: stretch;
  }
}

@media (max-width: 480px) {
  .stats-section {
    grid-template-columns: 1fr;
  }

  .users-container {
    padding: 10px;
  }

  .action-buttons {
    justify-content: center;
  }
}

/* StatusBadge 样式补充 */
:deep(.status-badge.landlord) {
  background: #dbeafe;
  color: #1e40af;
}

:deep(.status-badge.employee) {
  background: #fef3c7;
  color: #92400e;
}

:deep(.status-badge.success),
:deep(.status-badge.active) {
  background: #dcfce7;
  color: #166534;
}

:deep(.status-badge.danger),
:deep(.status-badge.inactive) {
  background: #fee2e2;
  color: #991b1b;
}
</style>