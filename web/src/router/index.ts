import { createRouter, createWebHistory } from 'vue-router'
import { useAuth } from '@/composables/useAuth'
import Login from '@/views/Login.vue'
import Dashboard from '@/views/Dashboard.vue'
import Rooms from '@/views/Rooms.vue'
import Tenants from '@/views/Tenants.vue'
import Bills from '@/views/Bills.vue'
import Utilities from '@/views/Utilities.vue'
import Reports from '@/views/Reports.vue'
import Settings from '@/views/Settings.vue'
import Users from '@/views/Users.vue'
import Properties from '@/views/Properties.vue'

// 路由配置
const routes = [
    {
        path: '/login',
        name: 'Login',
        component: Login,
        meta: {
            requiresAuth: false,
            title: '用户登录'
        }
    },
    {
        path: '/',
        redirect: '/dashboard'
    },
    {
        path: '/dashboard',
        name: 'Dashboard',
        component: Dashboard,
        meta: {
            requiresAuth: true,
            title: '仪表板',
            permission: 'dashboard.view'
        }
    },
    {
        path: '/rooms',
        name: 'Rooms',
        component: Rooms,
        meta: {
            requiresAuth: true,
            title: '房间管理',
            permission: 'rooms.view'
        }
    },
    {
        path: '/tenants',
        name: 'Tenants',
        component: Tenants,
        meta: {
            requiresAuth: true,
            title: '租客管理',
            permission: 'tenants.view'
        }
    },
    {
        path: '/bills',
        name: 'Bills',
        component: Bills,
        meta: {
            requiresAuth: true,
            title: '账单管理',
            permission: 'bills.view'
        }
    },
    {
        path: '/utilities',
        name: 'Utilities',
        component: Utilities,
        meta: {
            requiresAuth: true,
            title: '水电管理',
            permission: 'utilities.view'
        }
    },
    {
        path: '/reports',
        name: 'Reports',
        component: Reports,
        meta: {
            requiresAuth: true,
            title: '报表分析',
            permission: 'reports.view',
            requiresLandlord: true
        }
    },
    {
        path: '/settings',
        name: 'Settings',
        component: Settings,
        meta: {
            requiresAuth: true,
            title: '系统设置',
            permission: 'settings.view'
        }
    },
    {
        path: '/users',
        name: 'Users',
        component: Users,
        meta: {
            requiresAuth: true,
            title: '用户管理',
            permission: 'users.view',
            requiresLandlord: true // 只有主房东可以访问
        }
    },
    {
        path: '/properties',
        name: 'Properties',
        component: Properties,
        meta: {
            requiresAuth: true,
            title: '物业管理',
            permission: 'properties.view'
        }
    }
]

const router = createRouter({
    history: createWebHistory(),
    routes
})

// 全局路由守卫
router.beforeEach((to, _from, next) => {
    const { isAuthenticated, hasPermission, isLandlord } = useAuth()

    // 设置页面标题
    if (to.meta?.title) {
        document.title = `${to.meta.title} - 智能房租管理系统`
    }

    // 如果是登录页面
    if (to.path === '/login') {
        // 如果已经登录，重定向到首页
        if (isAuthenticated.value) {
            next('/dashboard')
        } else {
            next()
        }
        return
    }

    // 检查是否需要认证
    if (to.meta?.requiresAuth !== false) {
        if (!isAuthenticated.value) {
            // 未登录，重定向到登录页
            next('/login')
            return
        }

        // 检查权限
        const requiredPermission = to.meta?.permission as string
        if (requiredPermission && !hasPermission(requiredPermission)) {
            // 没有权限，显示错误信息并返回首页
            console.warn(`用户没有访问 ${to.path} 的权限`)
            next('/dashboard')
            return
        }

        // 检查是否需要主房东权限
        if (to.meta?.requiresLandlord && !isLandlord.value) {
            console.warn(`该页面只有主房东可以访问: ${to.path}`)
            next('/dashboard')
            return
        }
    }

    next()
})

export default router 