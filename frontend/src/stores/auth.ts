import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from '@/utils/axios'

// 解析JWT token获取payload
function parseJWT(token: string): { user_id: number; role: string } | null {
  try {
    const base64Url = token.split('.')[1]
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map(c => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    )
    return JSON.parse(jsonPayload)
  } catch {
    return null
  }
}

export const useAuthStore = defineStore('auth', () => {
  // 从localStorage恢复登录状态
  const token = ref<string>(localStorage.getItem('auth_token') || '')
  const decoded = token.value ? parseJWT(token.value) : null
  const role = ref<string>(decoded?.role || '')
  const userId = ref<number>(decoded?.user_id || 0)
  const forceChangePassword = ref<boolean>(false)

  const isAuthenticated = computed(() => !!token.value)
  const isAdmin = computed(() => role.value === 'admin')
  const isUser = computed(() => role.value === 'user')

  async function login(username: string, password: string) {
    const response = await axios.post('/api/auth/login', { username, password })
    token.value = response.data.token
    userId.value = response.data.user_id
    role.value = response.data.role
    forceChangePassword.value = response.data.force_change_password === 1
    // 存储到localStorage，刷新后可恢复
    localStorage.setItem('auth_token', response.data.token)
    return response.data
  }

  async function changePassword(oldPassword: string, newPassword: string) {
    await axios.put('/api/profile/password', { oldPassword, newPassword })
    forceChangePassword.value = false
  }

  async function logout() {
    try {
      await axios.post('/api/auth/logout')
    } catch {
      // 即使请求失败也继续清理
    }
    token.value = ''
    userId.value = 0
    role.value = ''
    forceChangePassword.value = false
    localStorage.removeItem('auth_token')
  }

  // 通过后端验证token有效性并刷新用户信息
  async function verify() {
    try {
      const response = await axios.get('/api/auth/verify')
      userId.value = response.data.user_id
      role.value = response.data.role
      forceChangePassword.value = response.data.force_change_password === 1
      return true
    } catch {
      token.value = ''
      userId.value = 0
      role.value = ''
      localStorage.removeItem('auth_token')
      return false
    }
  }

  return {
    token,
    userId,
    role,
    forceChangePassword,
    isAuthenticated,
    isAdmin,
    isUser,
    login,
    changePassword,
    logout,
    verify
  }
})
