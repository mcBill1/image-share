<template>
  <div class="dashboard">
    <h2>仪表盘</h2>
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-icon images-icon">
            <el-icon><Picture /></el-icon>
          </div>
          <div class="stat-info">
            <p class="stat-value">{{ stats.image_count }}</p>
            <p class="stat-label">图片总数</p>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-icon users-icon">
            <el-icon><User /></el-icon>
          </div>
          <div class="stat-info">
            <p class="stat-value">{{ stats.user_count }}</p>
            <p class="stat-label">用户数量</p>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-icon tasks-icon">
            <el-icon><Link /></el-icon>
          </div>
          <div class="stat-info">
            <p class="stat-value">{{ stats.task_count }}</p>
            <p class="stat-label">游客链接</p>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-icon storage-icon">
            <el-icon><Coin /></el-icon>
          </div>
          <div class="stat-info">
            <p class="stat-value">{{ formatSize(stats.storage_used) }}</p>
            <p class="stat-label">图片占用</p>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 磁盘空间条状图 -->
    <el-card class="disk-card" shadow="hover">
      <template #header>
        <div class="disk-card-header">
          <span>磁盘空间</span>
          <span class="disk-summary">{{ formatSize(stats.disk_total - stats.disk_free) }} / {{ formatSize(stats.disk_total) }}</span>
        </div>
      </template>
      <div class="disk-bar-container">
        <div class="disk-bar">
          <div class="disk-bar-images" :style="{ width: imagePercent + '%' }"></div>
          <div class="disk-bar-other" :style="{ width: otherPercent + '%' }"></div>
        </div>
        <div class="disk-bar-labels">
          <div class="disk-label">
            <span class="disk-dot disk-dot-images"></span>
            <span>图片 {{ formatSize(stats.storage_used) }} ({{ imagePercent.toFixed(1) }}%)</span>
          </div>
          <div class="disk-label">
            <span class="disk-dot disk-dot-other"></span>
            <span>其他 {{ formatSize(otherUsed) }} ({{ otherPercent.toFixed(1) }}%)</span>
          </div>
          <div class="disk-label">
            <span class="disk-dot disk-dot-free"></span>
            <span>可用 {{ formatSize(stats.disk_free) }} ({{ freePercent.toFixed(1) }}%)</span>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Picture, User, Link, Coin } from '@element-plus/icons-vue'
import axios from '@/utils/axios'

const stats = ref({
  image_count: 0,
  user_count: 0,
  task_count: 0,
  storage_used: 0,
  disk_total: 0,
  disk_free: 0
})

const imagePercent = computed(() => {
  if (stats.value.disk_total === 0) return 0
  return (stats.value.storage_used / stats.value.disk_total) * 100
})

const otherUsed = computed(() => {
  const totalUsed = stats.value.disk_total - stats.value.disk_free
  const other = totalUsed - stats.value.storage_used
  return other > 0 ? other : 0
})

const otherPercent = computed(() => {
  if (stats.value.disk_total === 0) return 0
  return (otherUsed.value / stats.value.disk_total) * 100
})

const freePercent = computed(() => {
  if (stats.value.disk_total === 0) return 0
  return (stats.value.disk_free / stats.value.disk_total) * 100
})

function formatSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

async function loadStats() {
  try {
    const response = await axios.get('/api/admin/stats')
    stats.value = response.data
  } catch (error) {
    console.error('Failed to load stats:', error)
  }
}

onMounted(() => {
  loadStats()
})
</script>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  min-height: 0;
  flex: 1;
}

.dashboard h2 {
  margin: 0 0 20px 0;
  color: var(--el-text-color-primary);
  flex-shrink: 0;
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

.images-icon {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
}

.users-icon {
  background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
  color: #fff;
}

.tasks-icon {
  background: linear-gradient(135deg, #eb3349 0%, #f45c43 100%);
  color: #fff;
}

.storage-icon {
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

.disk-card {
  margin-top: 20px;
  flex-shrink: 0;
}

.disk-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
}

.disk-summary {
  color: var(--el-text-color-secondary);
  font-weight: 400;
  font-size: 14px;
}

.disk-bar-container {
  padding: 0;
}

.disk-bar {
  height: 28px;
  border-radius: 8px;
  background: #e8eaed;
  display: flex;
  overflow: hidden;
}

.disk-bar-images {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  height: 100%;
  min-width: 0;
  transition: width 0.5s ease;
}

.disk-bar-other {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  height: 100%;
  min-width: 0;
  transition: width 0.5s ease;
}

.disk-bar-labels {
  display: flex;
  gap: 24px;
  margin-top: 12px;
  flex-wrap: wrap;
}

.disk-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--el-text-color-regular);
}

.disk-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.disk-dot-images {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.disk-dot-other {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.disk-dot-free {
  background: #e8eaed;
}
</style>
