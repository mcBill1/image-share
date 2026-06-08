<template>
  <div class="user-dashboard">
    <h2>我的图库</h2>
    
    <el-row :gutter="20">
      <el-col :xs="12" :sm="6">
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
      <el-col :xs="12" :sm="6">
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
      <el-col :xs="12" :sm="6">
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
      <el-col :xs="12" :sm="6">
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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Picture, Files, Coin, TopRight } from '@element-plus/icons-vue'
import axios from '@/utils/axios'

const stats = ref({
  used_count: 0,
  image_limit: 0,
  used_size: 0,
  storage_limit: 0
})

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

.stat-card {
  display: flex;
  align-items: center;
  padding: 20px;
  margin-bottom: 12px;
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
  flex-shrink: 0;
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
  min-width: 0;
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

/* 手机端适配 */
@media (max-width: 768px) {
  .stat-card {
    padding: 12px;
  }

  .stat-icon {
    width: 44px;
    height: 44px;
    font-size: 18px;
    margin-right: 12px;
  }

  .stat-value {
    font-size: 18px;
  }

  .stat-label {
    font-size: 12px;
  }
}
</style>
