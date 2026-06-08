<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <h2>ImageShare</h2>
        <p>图床管理与分享系统</p>
      </div>
      <el-form :model="form" ref="formRef" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="form.password" type="password" placeholder="请输入密码" @keyup.enter="handleLogin" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleLogin" class="login-btn" :loading="loading">
            登录
          </el-button>
        </el-form-item>
      </el-form>
      <div class="login-footer">
        <p>默认管理员: admin / image123456</p>
      </div>
      <div class="theme-toggle">
        <el-dropdown trigger="click" @command="handleThemeChange">
          <el-button text size="small">
            <el-icon><Sunny v-if="themeStore.mode === 'light'" /><Moon v-else-if="themeStore.mode === 'dark'" /><Monitor v-else /></el-icon>
            <span class="theme-label">{{ themeLabel }}</span>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="system"><el-icon><Monitor /></el-icon>跟随系统</el-dropdown-item>
              <el-dropdown-item command="light"><el-icon><Sunny /></el-icon>亮色模式</el-dropdown-item>
              <el-dropdown-item command="dark"><el-icon><Moon /></el-icon>暗色模式</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useThemeStore } from '@/stores/theme'
import { Sunny, Moon, Monitor } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import type { ThemeMode } from '@/stores/theme'
import { md5 } from 'js-md5'

const router = useRouter()
const authStore = useAuthStore()
const themeStore = useThemeStore()
const formRef = ref()

const loading = ref(false)
const form = reactive({
  username: '',
  password: ''
})

const themeLabel = computed(() => {
  switch (themeStore.mode) {
    case 'light': return '亮色'
    case 'dark': return '暗色'
    default: return '跟随系统'
  }
})

function handleThemeChange(command: string) {
  themeStore.setTheme(command as ThemeMode)
}

// 已登录用户自动跳转（需要改密码时留在登录页，让用户重新输入密码）
onMounted(async () => {
  if (authStore.isAuthenticated) {
    const valid = await authStore.verify()
    if (valid && !authStore.forceChangePassword) {
      if (authStore.isAdmin) {
        router.push('/admin/dashboard')
      } else {
        router.push('/user/dashboard')
      }
    }
  }
})

async function handleLogin() {
  if (!form.username || !form.password) {
    ElMessage.error('请填写用户名和密码')
    return
  }

  loading.value = true
  try {
    // 密码MD5加密后传输
    const hashedPassword = md5(form.password)
    await authStore.login(form.username, hashedPassword)
    ElMessage.success('登录成功')

    if (authStore.forceChangePassword) {
      router.push('/profile/change-password')
    } else if (authStore.isAdmin) {
      router.push('/admin/dashboard')
    } else {
      router.push('/user/dashboard')
    }
  } catch (error) {
    ElMessage.error('登录失败：用户名或密码错误')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100vw;
  height: 100vh;
  background: linear-gradient(135deg, #1e293b 0%, #334155 50%, #1e293b 100%);
  overflow: hidden;
}

html.dark .login-container {
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 50%, #0f172a 100%);
}

.login-card {
  width: 420px;
  max-width: 90%;
  padding: 48px 40px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  border-radius: 16px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
}

html.dark .login-card {
  background: rgba(30, 30, 30, 0.95);
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
}

.login-header {
  text-align: center;
  margin-bottom: 36px;
}

.login-header h2 {
  margin: 0 0 8px 0;
  color: #1e293b;
  font-size: 28px;
  letter-spacing: 2px;
}

html.dark .login-header h2 {
  color: #e5eaf3;
}

.login-header p {
  margin: 0;
  color: #94a3b8;
  font-size: 14px;
}

.login-btn {
  width: 100%;
  height: 44px;
  border-radius: 8px;
  font-size: 15px;
}

.login-footer {
  margin-top: 24px;
  text-align: center;
}

.login-footer p {
  margin: 0;
  color: #94a3b8;
  font-size: 12px;
}

.theme-toggle {
  margin-top: 16px;
  text-align: center;
}

.theme-label {
  margin-left: 4px;
  font-size: 12px;
}
</style>
