import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      redirect: '/login'
    },
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/Login.vue')
    },
    {
      path: '/profile/change-password',
      name: 'ChangePassword',
      component: () => import('@/views/ChangePassword.vue')
    },
    {
      path: '/admin',
      name: 'Admin',
      component: () => import('@/views/admin/AdminLayout.vue'),
      children: [
        {
          path: 'dashboard',
          name: 'AdminDashboard',
          component: () => import('@/views/admin/Dashboard.vue')
        },
        {
          path: 'images',
          name: 'AdminImages',
          component: () => import('@/views/admin/Images.vue')
        },
        {
          path: 'users',
          name: 'AdminUsers',
          component: () => import('@/views/admin/Users.vue')
        },
        {
          path: 'tasks',
          name: 'AdminTasks',
          component: () => import('@/views/admin/Tasks.vue')
        },
        {
          path: 'logs',
          name: 'AdminLogs',
          component: () => import('@/views/admin/Logs.vue')
        }
      ]
    },
    {
      path: '/user',
      name: 'User',
      component: () => import('@/views/user/UserLayout.vue'),
      children: [
        {
          path: 'dashboard',
          name: 'UserDashboard',
          component: () => import('@/views/user/Dashboard.vue')
        },
        {
          path: 'images',
          name: 'UserImages',
          component: () => import('@/views/user/Images.vue')
        }
      ]
    },
    {
      path: '/upload/:code',
      name: 'GuestUpload',
      component: () => import('@/views/GuestUpload.vue')
    }
  ]
})

let verified = false

router.beforeEach(async (to, _from, next) => {
  const authStore = useAuthStore()
  
  if (to.path.startsWith('/upload/')) {
    next()
    return
  }

  if (to.path === '/login') {
    if (authStore.isAuthenticated) {
      const valid = await authStore.verify()
      if (valid) {
        verified = true
        if (authStore.forceChangePassword) {
          // 需要改密码时，让用户留在登录页重新输入密码
          next()
          return
        } else if (authStore.isAdmin) {
          next('/admin/dashboard')
          return
        } else {
          next('/user/dashboard')
          return
        }
      }
    }
    next()
    return
  }

  if (to.path === '/profile/change-password') {
    // 改密码页需要先登录
    if (!authStore.token) {
      next('/login')
      return
    }
    if (!verified) {
      const valid = await authStore.verify()
      verified = true
      if (!valid) {
        next('/login')
        return
      }
    }
    next()
    return
  }

  if (!authStore.token) {
    next('/login')
    return
  }

  // 首次导航时验证token有效性
  if (!verified) {
    const valid = await authStore.verify()
    verified = true
    if (!valid) {
      next('/login')
      return
    }
  }

  if (authStore.forceChangePassword) {
    // 需要改密码时，跳转到登录页（用户需重新输入密码后再跳转改密码页）
    next('/login')
    return
  }

  if (to.path.startsWith('/admin') && authStore.role !== 'admin') {
    next('/user/dashboard')
    return
  }

  if (to.path.startsWith('/user') && authStore.role === 'admin') {
    next('/admin/dashboard')
    return
  }

  next()
})

export default router