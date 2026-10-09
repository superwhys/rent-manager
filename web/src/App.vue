<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { PropertyOption } from '@/types'
import { useAuth } from '@/composables/useAuth'
import ConfirmModal from '@/components/ConfirmModal.vue'


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

const selectedProperty = ref<number>(1)
const sidebarVisible = ref<boolean>(false)
const showLogoutConfirm = ref<boolean>(false)

const { isAuthenticated, currentUser, isLandlord, logout, initAuth } = useAuth()
const router = useRouter()
const toggleSidebar = () => {
  sidebarVisible.value = !sidebarVisible.value
}

const closeSidebar = () => {
  sidebarVisible.value = false
}

/**
 * 用户登出
 */
const handleLogout = () => {
  showLogoutConfirm.value = true
}

/**
 * 确认退出登录
 */
const confirmLogout = () => {
  showLogoutConfirm.value = false
  logout()
  router.push('/login')
  ElMessage.success('已退出登录')
}

/**
 * 取消退出登录
 */
const cancelLogout = () => {
  showLogoutConfirm.value = false
}

// 组件挂载时初始化认证状态
onMounted(() => {
  initAuth()
})
</script>

<template>
  <div v-if="$route.path === '/login'" class="login-page">
    <router-view />
  </div>

  <el-container v-else class="main-container">
    <el-header class="header">
      <div class="header-content">
        <div class="header-left">
          <button class="menu-toggle" @click="toggleSidebar">
            <span class="hamburger"></span>
            <span class="hamburger"></span>
            <span class="hamburger"></span>
          </button>
          <div class="logo">🏠 智能房租管理系统</div>
        </div>
        <div class="header-controls">
          <div class="property-selector" v-if="isAuthenticated">
            <el-select v-model="selectedProperty" class="property-select" placeholder="选择物业" teleported
              popper-class="high-z-index-popper">
              <el-option v-for="property in propertyOptions" :key="property.value" :label="property.label"
                :value="property.value" />
            </el-select>
          </div>
          <!-- 用户信息区域 -->
          <div class="user-info" v-if="isAuthenticated && currentUser">
            <el-dropdown trigger="click">
              <div class="user-avatar">
                <span class="avatar-text">{{ currentUser.realName.charAt(0) }}</span>
                <span class="user-name">{{ currentUser.realName }}</span>
                <span class="user-role">{{ isLandlord ? '主房东' : '员工' }}</span>
              </div>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item>个人信息</el-dropdown-item>
                  <el-dropdown-item>修改密码</el-dropdown-item>
                  <el-dropdown-item divided @click="handleLogout">退出登录</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>
      </div>
    </el-header>

    <el-container>
      <div v-if="sidebarVisible" class="sidebar-overlay" @click="closeSidebar"></div>

      <el-aside v-if="isAuthenticated" class="sidebar" :class="{ 'sidebar-open': sidebarVisible }">
        <nav style="padding: 20px;">
          <ul class="nav-menu">
            <PermissionGuard permission="dashboard.view">
              <NavItem to="/dashboard" icon="📊" text="仪表板" @click="closeSidebar" />
            </PermissionGuard>
            <PermissionGuard permission="rooms.view">
              <NavItem to="/rooms" icon="🏠" text="房间管理" @click="closeSidebar" />
            </PermissionGuard>
            <PermissionGuard permission="tenants.view">
              <NavItem to="/tenants" icon="👥" text="租客管理" @click="closeSidebar" />
            </PermissionGuard>
            <PermissionGuard permission="bills.view">
              <NavItem to="/bills" icon="💰" text="账单管理" @click="closeSidebar" />
            </PermissionGuard>
            <PermissionGuard permission="utilities.view">
              <NavItem to="/utilities" icon="⚡" text="水电管理" @click="closeSidebar" />
            </PermissionGuard>
            <!-- <PermissionGuard permission="reports.view">
              <NavItem to="/reports" icon="📈" text="报表分析" @click="closeSidebar" />
            </PermissionGuard> -->
            <PermissionGuard permission="users.view">
              <NavItem v-if="isLandlord" to="/users" icon="👤" text="用户管理" @click="closeSidebar" />
            </PermissionGuard>
            <PermissionGuard permission="properties.view">
              <NavItem to="/properties" icon="🏢" text="物业管理" @click="closeSidebar" />
            </PermissionGuard>
            <PermissionGuard permission="settings.view">
              <NavItem to="/settings" icon="⚙️" text="系统设置" @click="closeSidebar" />
            </PermissionGuard>
          </ul>
        </nav>
      </el-aside>

      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>

  <ConfirmModal :visible="showLogoutConfirm" title="确认退出" message="确定要退出登录吗？退出后需要重新登录才能访问系统。" confirm-text="确定退出"
    cancel-text="取消" type="warning" @confirm="confirmLogout" @cancel="cancelLogout" @close="cancelLogout" />
</template>

<style scoped>
.login-page {
  width: 100vw;
  height: 100vh;
  margin: 0;
  padding: 0;
  overflow: hidden;
}

.main-container {
  min-height: 100vh;
}

.header {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 20px;
  padding: 20px 30px;
  margin-bottom: 30px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  height: auto;
  min-height: 70px;
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  position: relative;
  z-index: 100;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 20px;
  width: 100%;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 15px;
}

.menu-toggle {
  display: none;
  flex-direction: column;
  justify-content: space-around;
  width: 30px;
  height: 30px;
  background: transparent;
  border: none;
  cursor: pointer;
  padding: 0;
  z-index: 1001;
}

.hamburger {
  width: 100%;
  height: 3px;
  background-color: #4f46e5;
  border-radius: 2px;
  transition: all 0.3s ease;
}

.logo {
  font-size: 28px;
  font-weight: bold;
  color: #4f46e5;
}

.header-controls {
  display: flex;
  gap: 15px;
  align-items: center;
  flex-wrap: wrap;
}



.property-selector {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  flex-wrap: wrap;
}

.property-select {
  min-width: 200px;
}

.user-info {
  margin-left: 20px;
}

.user-avatar {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 8px 12px;
  border-radius: 12px;
  background: rgba(79, 70, 229, 0.1);
  transition: all 0.3s ease;
}

.user-avatar:hover {
  background: rgba(79, 70, 229, 0.2);
  transform: translateY(-1px);
}

.avatar-text {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: #4f46e5;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 14px;
}

.user-name {
  font-weight: 600;
  color: #374151;
  font-size: 14px;
}

.user-role {
  font-size: 12px;
  color: #6b7280;
  background: rgba(156, 163, 175, 0.2);
  padding: 2px 6px;
  border-radius: 6px;
}

.sidebar {
  width: 280px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 20px;
  margin: 0 20px 20px 0;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  transition: transform 0.3s ease;
  z-index: 1000;
}

.nav-menu {
  list-style: none;
  padding: 0;
  margin: 0;
}

.main-content {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 20px;
  margin: 0 0 20px 0;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  transition: margin-left 0.3s ease;
}

.sidebar-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.5);
  z-index: 999;
  display: none;
}

/* 响应式设计 */
@media (max-width: 1100px) {
  .menu-toggle {
    display: flex;
  }

  .sidebar {
    position: fixed;
    top: 0;
    left: 0;
    height: 100vh;
    margin: 0;
    border-radius: 0 20px 20px 0;
    transform: translateX(-100%);
    z-index: 1000;
  }

  .sidebar-open {
    transform: translateX(0);
  }

  .sidebar-overlay {
    display: block;
  }

  .main-content {
    margin: 0 0 20px 0;
  }

  .header {
    padding: 15px 20px;
    margin: 10px 0 20px 0;
  }

  .header-content {
    flex-direction: column;
    align-items: flex-start;
    gap: 15px;
  }

  .header-left {
    width: 100%;
    justify-content: space-between;
  }

  .logo {
    font-size: 24px;
  }

  .header-controls {
    width: 100%;
    flex-direction: column;
    align-items: stretch;
    gap: 15px;
  }



  .property-selector {
    width: 100%;
    justify-content: flex-start;
  }

  .property-select {
    flex: 1;
    min-width: 0;
  }
}

@media (max-width: 480px) {
  .header {
    padding: 12px 15px;
    border-radius: 15px;
    margin: 10px 0;
  }

  .logo {
    font-size: 20px;
  }



  .property-select {
    min-width: auto;
    width: 100%;
  }

  .sidebar {
    width: 260px;
  }

  .main-content {
    margin: 0 0 10px 0;
  }
}

/* :deep(.el-dropdown-menu) {
  z-index: 3000 !important;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
  border: 1px solid rgba(0, 0, 0, 0.05);
}

:deep(.el-dropdown-menu__item) {
  border-radius: 8px;
  margin: 4px 8px;
  transition: all 0.2s ease;
}

:deep(.el-dropdown-menu__item:hover) {
  background: rgba(79, 70, 229, 0.1);
  color: #4f46e5;
}

:deep(.el-dropdown-menu__item.is-divided) {
  border-top: 1px solid rgba(0, 0, 0, 0.05);
  margin-top: 8px;
}

:deep(.el-message) {
  z-index: 5000 !important;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
  border: none;
}

:deep(.el-message.el-message--success) {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
}

:deep(.el-message.el-message--warning) {
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  color: white;
}

:deep(.el-message.el-message--error) {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
} */
</style>
