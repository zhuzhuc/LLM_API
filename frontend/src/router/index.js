import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '../stores/user'

const routes = [
  {
    path: '/',
    redirect: '/dashboard'
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('../views/Dashboard.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { requiresGuest: true }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('../views/Register.vue'),
    meta: { requiresGuest: true }
  },
  {
    path: '/chat',
    name: 'Chat',
    component: () => import('../views/Chat.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/profile',
    name: 'Profile',
    component: () => import('../views/Profile.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/history',
    name: 'History',
    component: () => import('../views/History.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/stats',
    name: 'Stats',
    component: () => import('../views/Stats.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/models',
    name: 'Models',
    component: () => import('../views/Models.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/converter',
    name: 'Converter',
    component: () => import('../views/Converter.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/homework',
    name: 'Homework',
    component: () => import('../views/Homework.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/subtitle',
    name: 'Subtitle',
    component: () => import('../views/Subtitle.vue'),
    meta: { requiresAuth: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore()

  console.log('路由守卫 - 目标路径:', to.path)
  console.log('路由守卫 - token:', userStore.token)
  console.log('路由守卫 - user:', userStore.user)
  console.log('路由守卫 - isLoggedIn:', userStore.isLoggedIn)

  // 如果有token但没有用户信息，尝试获取用户信息
  if (userStore.token && !userStore.user) {
    console.log('尝试获取用户信息...')
    try {
      await userStore.fetchProfile()
    } catch (error) {
      console.log('获取用户信息失败，清除token')
      // 如果获取失败，清除token
      userStore.logout()
    }
  }

  // 避免无限循环重定向
  if (to.path === '/login' && userStore.isLoggedIn) {
    console.log('用户已登录，重定向到仪表板')
    next('/dashboard')
    return
  }

  if (to.meta.requiresAuth && !userStore.isLoggedIn) {
    console.log('需要认证但未登录，重定向到登录页面')
    next('/login')
  } else if (to.meta.requiresGuest && userStore.isLoggedIn) {
    console.log('访客页面但已登录，重定向到仪表板')
    next('/dashboard')
  } else {
    console.log('正常导航')
    next()
  }
})

export default router