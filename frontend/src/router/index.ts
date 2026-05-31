import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { getToken } from '@/utils/auth'

export const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { title: '登录', hidden: true },
  },
  {
    path: '/',
    component: () => import('@/components/Layout/AppLayout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: 'Dashboard', icon: 'Odometer' },
      },
      {
        path: 'clusters',
        name: 'Clusters',
        component: () => import('@/views/Clusters.vue'),
        meta: { title: '集群管理', icon: 'Connection' },
      },
      {
        path: 'system/users',
        name: 'Users',
        component: () => import('@/views/system/Users.vue'),
        meta: { title: '用户管理', icon: 'User', parent: '系统管理' },
      },
      {
        path: 'system/roles',
        name: 'Roles',
        component: () => import('@/views/system/Roles.vue'),
        meta: { title: '角色管理', icon: 'UserFilled', parent: '系统管理' },
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to, _from, next) => {
  const token = getToken()
  if (to.path !== '/login' && !token) {
    next('/login')
  } else {
    next()
  }
})

export default router
