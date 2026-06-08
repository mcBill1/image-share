<template>
  <div class="change-password-container">
    <div class="change-password-card">
      <div class="change-password-header">
        <h2>修改密码</h2>
        <p>首次登录请修改默认密码</p>
      </div>
      <el-form :model="form" ref="formRef" label-width="100px">
        <el-form-item label="原密码" prop="oldPassword">
          <el-input v-model="form.oldPassword" type="password" placeholder="请输入原密码" />
        </el-form-item>
        <el-form-item label="新密码" prop="newPassword">
          <el-input v-model="form.newPassword" type="password" placeholder="请输入新密码" />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input v-model="form.confirmPassword" type="password" placeholder="请再次输入新密码" @keyup.enter="handleChangePassword" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleChangePassword" class="change-password-btn" :loading="loading">
            修改密码
          </el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'
import { md5 } from 'js-md5'

const router = useRouter()
const authStore = useAuthStore()
const formRef = ref()

const loading = ref(false)
const form = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

async function handleChangePassword() {
  if (!form.oldPassword || !form.newPassword || !form.confirmPassword) {
    ElMessage.error('请填写所有字段')
    return
  }

  if (form.newPassword.length < 6) {
    ElMessage.error('新密码至少需要6位')
    return
  }

  if (!/[a-zA-Z]/.test(form.newPassword)) {
    ElMessage.error('新密码必须包含至少一个字母')
    return
  }

  if (form.newPassword !== form.confirmPassword) {
    ElMessage.error('两次输入的密码不一致')
    return
  }

  if (form.oldPassword === form.newPassword) {
    ElMessage.error('新密码不能与旧密码相同')
    return
  }

  loading.value = true
  try {
    // 密码MD5加密后传输
    await authStore.changePassword(md5(form.oldPassword), md5(form.newPassword))
    ElMessage.success('密码修改成功')

    if (authStore.isAdmin) {
      router.push('/admin/dashboard')
    } else {
      router.push('/user/dashboard')
    }
  } catch (error: any) {
    const msg = error.response?.data?.error || '密码修改失败'
    ElMessage.error(msg)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.change-password-container {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100vw;
  height: 100vh;
  background: linear-gradient(135deg, #1e293b 0%, #334155 50%, #1e293b 100%);
  overflow: hidden;
}

html.dark .change-password-container {
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 50%, #0f172a 100%);
}

.change-password-card {
  width: 460px;
  padding: 48px 40px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  border-radius: 16px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
}

html.dark .change-password-card {
  background: rgba(30, 30, 30, 0.95);
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
}

.change-password-header {
  text-align: center;
  margin-bottom: 36px;
}

.change-password-header h2 {
  margin: 0 0 8px 0;
  color: #1e293b;
  font-size: 24px;
}

html.dark .change-password-header h2 {
  color: #e5eaf3;
}

.change-password-header p {
  margin: 0;
  color: #94a3b8;
  font-size: 14px;
}

.change-password-btn {
  width: 100%;
  height: 44px;
  border-radius: 8px;
  font-size: 15px;
}
</style>
