<template>
  <div class="user-dashboard">
    <h2>我的图库</h2>
    
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-icon used-icon">
            <el-icon><Picture /></el-icon>
          </div>
          <div class="stat-info">
            <p class="stat-value">{{ stats.used_count }}</p>
            <p class="stat-label">已上传图片</p>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-icon limit-icon">
            <el-icon><Files /></el-icon>
          </div>
          <div class="stat-info">
            <p class="stat-value">{{ stats.image_limit }}</p>
            <p class="stat-label">图片上限</p>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-icon storage-icon">
            <el-icon><Coin /></el-icon>
          </div>
          <div class="stat-info">
            <p class="stat-value">{{ formatSize(stats.used_size) }}</p>
            <p class="stat-label">已用空间</p>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-icon remaining-icon">
            <el-icon><TopRight /></el-icon>
          </div>
          <div class="stat-info">
            <p class="stat-value">{{ formatSize(stats.storage_limit) }}</p>
            <p class="stat-label">空间上限</p>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="hover" style="margin-top: 20px;">
      <h3>上传图片</h3>
      <el-upload
        class="upload-demo"
        action="/api/user/upload"
        :headers="{ Authorization: 'Bearer ' + authStore.token }"
        :on-success="handleUploadSuccess"
        :on-error="handleUploadError"
        :file-list="uploadFiles"
        accept="image/*"
        multiple
      >
        <el-button type="primary">
          <el-icon><Upload /></el-icon>
          选择图片
        </el-button>
      </el-upload>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Picture, Files, Coin, TopRight, Upload } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'
import { ElMessage } from 'element-plus'

const authStore = useAuthStore()

const stats = ref({
  used_count: 0,
  image_limit: 0,
  used_size: 0,
  storage_limit: 0
})

const uploadFiles = ref<any[]>([])

function formatSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

async function loadStats() {
  try {
    const response = await axios.get('/api/user/stats')
    stats.value.used_count = response.data.used_count
    stats.value.used_size = response.data.used_size
    stats.value.image_limit = response.data.image_limit
    stats.value.storage_limit = response.data.storage_limit
  } catch (error) {
    console.error('Failed to load stats:', error)
  }
}

function handleUploadSuccess() {
  ElMessage.success('上传成功')
  uploadFiles.value = []
  loadStats()
}

function handleUploadError(error: any) {
  ElMessage.error(error.response?.data?.error || '上传失败')
}

onMounted(() => {
  loadStats()
})
</script>

<style scoped>
.user-dashboard {
  display: flex;
  flex-direction: column;
  min-height: 0;
  flex: 1;
}

.user-dashboard h2 {
  margin: 0 0 20px 0;
  color: var(--el-text-color-primary);
  flex-shrink: 0;
}

.user-dashboard h3 {
  margin: 0 0 15px 0;
  color: var(--el-text-color-secondary);
  font-size: 16px;
}

.stat-card {
  display: flex;
  align-items: center;
  padding: 20px;
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  display: flex;
  justify-content: center;
  align-items: center;
  margin-right: 20px;
  font-size: 24px;
}

.used-icon {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
}

.limit-icon {
  background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
  color: #fff;
}

.storage-icon {
  background: linear-gradient(135deg, #eb3349 0%, #f45c43 100%);
  color: #fff;
}

.remaining-icon {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
  color: #fff;
}

.stat-info {
  flex: 1;
}

.stat-value {
  margin: 0;
  font-size: 24px;
  font-weight: bold;
  color: var(--el-text-color-primary);
}

.stat-label {
  margin: 5px 0 0 0;
  color: var(--el-text-color-secondary);
  font-size: 14px;
}
</style>
