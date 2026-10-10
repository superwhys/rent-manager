import { createRouter, createWebHistory } from 'vue-router'
import { useSession } from '@/stores/session'

const router = createRouter({
  history: createWebHistory(),
  scrollBehavior: () => ({ top: 0 }),
  routes: [
    { path: '/', redirect: '/dashboard' },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/auth/LoginView.vue'),
      meta: { public: true, title: '登录' },
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/views/auth/RegisterView.vue'),
      meta: { public: true, title: '注册' },
    },
    {
      path: '/dashboard',
      name: 'overview',
      component: () => import('@/views/overview/OverviewView.vue'),
      meta: { title: '经营概览' },
    },
    {
      path: '/properties',
      name: 'units',
      component: () => import('@/views/units/UnitListView.vue'),
      meta: { title: '房源管理' },
    },
    {
      path: '/rooms',
      name: 'rooms',
      component: () => import('@/views/rooms/RoomListView.vue'),
      meta: { title: '房间管理' },
    },
    {
      path: '/bills',
      name: 'bills',
      component: () => import('@/views/bills/BillListView.vue'),
      meta: { title: '账单管理' },
    },
    {
      path: '/utilities',
      name: 'readings',
      component: () => import('@/views/readings/MeterReadingView.vue'),
      meta: { title: '抄表管理' },
    },
    {
      path: '/users',
      name: 'access',
      component: () => import('@/views/access/AccessView.vue'),
      meta: { title: '角色与授权' },
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('@/views/settings/SettingsView.vue'),
      meta: { title: '设置' },
    },
    {
      path: '/tenants',
      name: 'tenants',
      component: () => import('@/views/tenants/TenantListView.vue'),
      meta: { title: '租客管理' },
    },
    {
      path: '/reports',
      name: 'reports',
      component: () => import('@/views/reports/ReportView.vue'),
      meta: { title: '经营报表' },
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('@/views/NotFoundView.vue'),
      meta: { public: true, title: '页面不存在' },
    },
  ],
})

router.beforeEach(async (to) => {
  document.title = `${String(to.meta.title ?? '工作台')} · 租序`
  const auth = useSession()
  if (to.meta.public) return
  try {
    await auth.initialize()
  } catch {
    return { path: '/login', query: { redirect: to.fullPath, reconnect: '1' } }
  }
  if (!auth.isAuthenticated.value)
    return {
      path: '/login',
      query: {
        redirect: to.fullPath,
        expired: auth.expired.value ? '1' : undefined,
      },
    }
})

export default router
