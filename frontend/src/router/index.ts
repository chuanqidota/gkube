import { createRouter, createWebHistory } from 'vue-router'
import { getToken } from '@/utils/auth'
import { useAuthStore } from '@/stores/auth'
import i18n from '@/locales'

import { workloadRoutes } from './workload'
import { networkRoutes } from './network'
import { storageRoutes } from './storage'
import { configRoutes } from './config'
import { nodeRoutes } from './node'
import { clusterRoutes } from './cluster'
import { eventRoutes } from './event'
import { systemRoutes } from './system'

const { t } = i18n.global

const router = createRouter({
  history: createWebHistory(),
  routes: [
    // Standalone pages (no AppLayout)
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/login/LoginView.vue'),
      meta: { public: true },
    },
    // Fullscreen layout for terminal and logs (no sidebar/header, opens in new tab)
    {
      path: '/fullscreen',
      component: () => import('@/components/Layout/FullscreenLayout.vue'),
      children: [
        {
          path: 'terminal',
          name: 'Terminal',
          component: () => import('@/views/terminal/TerminalView.vue'),
          meta: { titleKey: 'terminal.title', icon: 'Promotion' },
        },
        {
          path: 'logs',
          name: 'Logs',
          component: () => import('@/views/logviewer/LogView.vue'),
          meta: { titleKey: 'log.title', icon: 'Document' },
        },
      ],
    },
    // AppLayout wrapper for all authenticated pages
    {
      path: '/',
      component: () => import('@/components/Layout/AppLayout.vue'),
      redirect: '/dashboard',
      children: [
        {
          path: 'dashboard',
          name: 'Dashboard',
          component: () => import('@/views/dashboard/DashboardView.vue'),
          meta: { titleKey: 'sidebar.dashboard', icon: 'Odometer' },
        },
        {
          path: 'system/overview',
          redirect: '/dashboard',
        },
        ...workloadRoutes,
        ...networkRoutes,
        ...storageRoutes,
        ...configRoutes,
        ...nodeRoutes,
        ...clusterRoutes,
        ...eventRoutes,
        ...systemRoutes,
        // 404 catch-all inside the app shell (keeps sidebar/header)
        {
          path: ':pathMatch(.*)*',
          name: 'NotFound',
          component: () => import('@/views/system/NotFound.vue'),
          meta: { titleKey: 'common.notFound' },
        },
      ],
    },
  ],
})

router.beforeEach(async (to, _from, next) => {
  const token = getToken()
  if (!to.meta.public && !token) {
    next({ path: '/login', query: to.fullPath !== '/' ? { redirect: to.fullPath } : {} })
    return
  }
  if (to.path === '/login' && token) {
    next('/dashboard')
    return
  }
  if (to.meta.requireAdmin) {
    const authStore = useAuthStore()
    if (!authStore.user) {
      next({ path: '/login', query: { redirect: to.fullPath } })
      return
    }
    if (!authStore.user.isSuperAdmin && !authStore.user.isAdmin) {
      next({ path: '/dashboard' })
      return
    }
  }
  if (token) {
    const authStore = useAuthStore()
    if (authStore.user && authStore.user.permissions === undefined) {
      try {
        await authStore.fetchPermissions()
        authStore.loadRoles()
      } catch {
        if (authStore.user) {
          authStore.user.permissions = null
        }
      }
    }
  }
  next()
})

router.afterEach((to) => {
  const title = to.meta.titleKey ? t(to.meta.titleKey as string) : (to.meta.title as string | undefined) || ''
  document.title = title ? `${title} - GKube` : `GKube - ${t('common.appDescription')}`
})

export default router
