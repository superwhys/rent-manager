<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import AppIcon from '@/components/common/AppIcon.vue'
import { useSession } from '@/stores/session'
import { resetUnits, useUnits } from '@/stores/units'
import { errorMessage } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const auth = useSession()
const { units, selectedUnitId, loadUnits, loading, error } = useUnits()
const drawer = ref(false)
const signingOut = ref(false)
const navItems = [
  { to: '/dashboard', icon: 'grid', label: '经营概览' },
  { to: '/properties', icon: 'building', label: '房源管理' },
  { to: '/rooms', icon: 'home', label: '房间管理' },
  { to: '/bills', icon: 'wallet', label: '账单管理' },
  { to: '/utilities', icon: 'droplet', label: '抄表管理' },
]
const otherItems = [
  { to: '/tenants', icon: 'users', label: '租客管理', pending: true },
  { to: '/reports', icon: 'chart', label: '经营报表', pending: true },
  { to: '/users', icon: 'shield', label: '角色与授权', pending: false },
  { to: '/settings', icon: 'settings', label: '设置', pending: false },
]
const scopedPage = computed(() =>
  ['/rooms', '/bills', '/utilities', '/users'].includes(route.path),
)
watch(
  () => auth.session.value?.account.ID,
  (accountId) => {
    resetUnits()
    if (accountId) void loadUnits()
  },
  { immediate: true, flush: 'sync' },
)

async function logout() {
  if (signingOut.value) return
  signingOut.value = true
  try {
    await auth.logout()
    await router.replace('/login')
  } catch (cause) {
    ElMessage.error(errorMessage(cause))
  } finally {
    signingOut.value = false
  }
}
</script>

<template>
  <div class="workspace">
    <a class="skip-link" href="#main-content">跳到主要内容</a>
    <aside class="sidebar" aria-label="主导航">
      <RouterLink to="/dashboard" class="brand">
        <span class="brand-mark"><AppIcon name="building" :size="25" /></span>
        <span><strong>租序</strong><small>租赁管理工作台</small></span>
      </RouterLink>
      <div class="workspace-label">我的工作空间</div>
      <nav>
        <RouterLink
          v-for="item in navItems"
          :key="item.to"
          :to="item.to"
          class="nav-link"
        >
          <AppIcon :name="item.icon" /><span>{{ item.label }}</span>
          <span v-if="route.path === item.to" class="active-dot" />
        </RouterLink>
        <div class="nav-divider" />
        <RouterLink
          v-for="item in otherItems"
          :key="item.to"
          :to="item.to"
          class="nav-link"
        >
          <AppIcon :name="item.icon" /><span>{{ item.label }}</span>
          <small v-if="item.pending" class="pending-label">待开放</small>
        </RouterLink>
      </nav>
      <div class="sidebar-bottom">
        <div class="calm-note">
          <span class="status-dot" /> 让每一处房源，井井有条
        </div>
        <RouterLink to="/settings" class="account-link">
          <span class="avatar">{{
            auth.accountName.value.slice(0, 1).toUpperCase()
          }}</span>
          <span class="account-copy">
            <strong>{{ auth.accountName.value }}</strong>
            <span v-if="auth.isSuperAdmin.value" class="badge account-badge">
              超级管理员
            </span>
            <small v-else>账号与偏好</small>
          </span>
          <AppIcon name="chevron-right" :size="15" />
        </RouterLink>
      </div>
    </aside>

    <div class="workspace-main">
      <header class="topbar">
        <div class="breadcrumb">
          <button
            class="icon-button mobile-menu"
            aria-label="打开导航"
            @click="drawer = true"
          >
            <AppIcon name="menu" />
          </button>
          <span class="breadcrumb-root">工作台</span
          ><span class="breadcrumb-separator">/</span>
          <span>{{ route.meta.title }}</span>
        </div>
        <div class="topbar-actions">
          <div v-if="scopedPage" class="unit-picker">
            <AppIcon name="building" :size="17" />
            <el-select
              v-model="selectedUnitId"
              :loading="loading"
              aria-label="当前房源"
              placeholder="选择房源"
              :disabled="loading"
            >
              <el-option
                v-for="unit in units"
                :key="unit.id"
                :value="unit.id"
                :label="unit.name"
              />
            </el-select>
          </div>
          <span v-else class="workspace-status"
            ><span class="status-dot" /> 我的租赁工作台</span
          >
          <button
            class="icon-button logout-button"
            :disabled="signingOut"
            aria-label="退出登录"
            title="退出登录"
            @click="logout"
          >
            <AppIcon name="logout" :size="18" />
          </button>
        </div>
      </header>
      <main id="main-content" class="page-content">
        <el-alert
          v-if="error"
          type="error"
          :closable="false"
          class="units-error"
        >
          房源列表加载失败：{{ error }}
          <el-button link type="primary" @click="loadUnits">重试</el-button>
        </el-alert>
        <slot />
        <footer class="page-footer">
          <span>租序</span><span>租赁管理工作台</span>
        </footer>
      </main>
    </div>
    <el-drawer
      v-model="drawer"
      title="我的工作空间"
      direction="ltr"
      size="280px"
      class="navigation-drawer"
    >
      <nav class="drawer-nav">
        <RouterLink
          v-for="item in [...navItems, ...otherItems]"
          :key="item.to"
          :to="item.to"
          class="nav-link"
          @click="drawer = false"
        >
          <AppIcon :name="item.icon" /><span>{{ item.label }}</span>
        </RouterLink>
      </nav>
    </el-drawer>
  </div>
</template>

<style scoped>
.workspace {
  min-height: 100dvh;
}
.sidebar {
  position: fixed;
  inset: 0 auto 0 0;
  width: 234px;
  padding: 2rem 1.05rem 1.2rem;
  background: rgba(238, 243, 241, 0.88);
  backdrop-filter: blur(24px);
  border-right: 1px solid rgba(213, 225, 220, 0.55);
  display: flex;
  flex-direction: column;
  z-index: 30;
}
.brand {
  display: flex;
  align-items: center;
  gap: 0.8rem;
  padding: 0 0.85rem;
  margin-bottom: 3.1rem;
}
.brand-mark {
  height: 44px;
  width: 44px;
  display: grid;
  place-items: center;
  border-radius: 13px;
  color: white;
  background: #267860;
  box-shadow: 0 4px 10px #1c604918;
}
.brand strong {
  display: block;
  font-size: 1.55rem;
  font-weight: 650;
  letter-spacing: 0.1em;
}
.brand small {
  display: block;
  font-size: 0.59rem;
  color: #788b80;
  letter-spacing: 0.13em;
  margin-top: 0.1rem;
}
.workspace-label {
  font-size: 0.72rem;
  color: #8a9891;
  padding: 0 1rem;
  margin-bottom: 0.8rem;
}
.nav-link {
  min-height: 44px;
  display: flex;
  align-items: center;
  gap: 0.9rem;
  color: #64746c;
  padding: 0.8rem 1rem;
  margin-bottom: 0.4rem;
  border-radius: 9px;
  font-size: 0.91rem;
  transition:
    background 0.15s,
    color 0.15s;
}
.nav-link:hover {
  background: #e3ece6;
}
.nav-link.router-link-active {
  background: #dcebe2;
  color: #276b50;
  font-weight: 600;
}
.active-dot {
  margin-left: auto;
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: #448668;
}
.nav-divider {
  height: 1px;
  background: #dce5df;
  margin: 1.35rem 0.85rem;
}
.pending-label {
  font-size: 0.61rem;
  margin-left: auto;
  color: #97a29c;
  white-space: nowrap;
  font-weight: 400;
}
.sidebar-bottom {
  margin-top: auto;
  padding-top: 3rem;
}
.calm-note {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  font-size: 0.69rem;
  color: #90a095;
  padding: 0 0.85rem 1.3rem;
}
.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #6ca18b;
  flex-shrink: 0;
}
.account-link {
  border-top: 1px solid #dce5df;
  padding: 1.2rem 0.6rem 0;
  display: flex;
  align-items: center;
  gap: 0.75rem;
}
.avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  color: #648373;
  background: #dce7df;
}
.account-copy {
  flex: 1;
  min-width: 0;
}
.account-copy strong {
  font-weight: 550;
  display: block;
  font-size: 0.8rem;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
.account-copy small {
  color: #8b9890;
  font-size: 0.68rem;
}
.account-badge {
  margin-top: 0.25rem;
  padding: 0.15rem 0.4rem;
  font-size: 0.65rem;
}
.workspace-main {
  margin-left: 234px;
}
.topbar {
  height: 74px;
  position: sticky;
  top: 0;
  z-index: 20;
  padding: 0 2.65rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  background: rgba(246, 248, 249, 0.85);
  backdrop-filter: blur(20px);
}
.topbar::after {
  content: '';
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  height: 12px;
  pointer-events: none;
  background: linear-gradient(#f6f8f9aa, transparent);
}
.breadcrumb {
  display: flex;
  align-items: center;
  gap: 0.9rem;
  font-size: 0.8rem;
  white-space: nowrap;
  color: #627169;
}
.breadcrumb-root {
  color: #939c97;
}
.breadcrumb-separator {
  color: #bec6c1;
}
.topbar-actions {
  display: flex;
  align-items: center;
  gap: 1.4rem;
}
.workspace-status {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #809086;
  font-size: 0.75rem;
}
.icon-button {
  border: 0;
  background: transparent;
  border-radius: 8px;
  width: 38px;
  height: 38px;
  display: grid;
  place-items: center;
  color: #77887d;
}
.icon-button:hover {
  background: #e9eeeb;
}
.unit-picker {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #819188;
}
.unit-picker .el-select {
  width: 190px;
}
.page-content {
  max-width: 1500px;
  margin: 0 auto;
  padding: 1.85rem 2.65rem 0;
}
.page-footer {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  padding: 2.5rem 0 1.5rem;
  color: #a4aeaa;
  font-size: 0.68rem;
}
.page-footer span:last-child {
  font-size: 0.6rem;
  letter-spacing: 0.15em;
}
.units-error {
  margin-bottom: 1.5rem;
}
.mobile-menu {
  display: none;
}
.skip-link {
  position: fixed;
  top: -100px;
  left: 1rem;
  z-index: 9999;
  padding: 1rem;
  background: white;
}
.skip-link:focus {
  top: 1rem;
}
@media (min-width: 1600px) {
  .page-content {
    padding-top: 2.5rem;
  }
}
@media (max-width: 1100px) {
  .sidebar {
    width: 210px;
    padding-left: 0.7rem;
    padding-right: 0.7rem;
  }
  .workspace-main {
    margin-left: 210px;
  }
  .topbar {
    padding: 0 1.5rem;
  }
  .page-content {
    padding: 1.5rem 1.5rem 0;
  }
}
@media (max-width: 800px) {
  .sidebar {
    display: none;
  }
  .workspace-main {
    margin-left: 0;
  }
  .mobile-menu {
    display: grid;
  }
  .topbar {
    padding: 0 1rem;
    height: 64px;
  }
  .page-content {
    padding: 1.25rem 1rem 0;
  }
  .breadcrumb-root,
  .breadcrumb-separator,
  .workspace-status {
    display: none;
  }
  .breadcrumb {
    gap: 0.5rem;
  }
  .topbar-actions {
    gap: 0.25rem;
  }
  .unit-picker {
    gap: 0.35rem;
  }
  .unit-picker .el-select {
    width: 145px;
  }
  .unit-picker > svg {
    display: none;
  }
  .page-footer span:last-child {
    display: none;
  }
}
@media (prefers-reduced-transparency: reduce), (prefers-contrast: more) {
  .sidebar {
    background: #eef3f1;
    backdrop-filter: none;
  }
  .topbar {
    background: #f6f8f9;
    backdrop-filter: none;
  }
}
</style>
