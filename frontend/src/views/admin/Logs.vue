<template>
  <div class="logs-container">
    <div class="logs-header">
      <div class="logs-title">
        <h2>系统日志</h2>
        <span class="log-file-name" v-if="fileName">当前文件: {{ fileName }}</span>
      </div>
    </div>
    <el-alert
      type="info"
      :closable="false"
      show-icon
      style="margin-bottom: 12px"
    >
      更多日志，请访问 /logs 目录查看
    </el-alert>
    <div class="logs-content" ref="logContent" v-loading="loading">
      <pre v-if="logContent_text">{{ logContent_text }}</pre>
      <el-empty v-else description="暂无日志" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { onActivated, onDeactivated } from 'vue'
import axios from '@/utils/axios'

const logContent_text = ref('')
const fileName = ref('')
const loading = ref(false)
const logContent = ref<HTMLElement>()
let timer: ReturnType<typeof setInterval> | null = null
let isActive = false

async function fetchLogs() {
  loading.value = true
  try {
    const res = await axios.get('/api/admin/logs', { params: { lines: 300 } })
    logContent_text.value = res.data.content || ''
    fileName.value = res.data.file_name || ''
  } catch {
    logContent_text.value = ''
  } finally {
    loading.value = false
  }
}

function scrollToBottom() {
  if (logContent.value) {
    logContent.value.scrollTop = logContent.value.scrollHeight
  }
}

function startPolling() {
  if (timer) clearInterval(timer)
  timer = setInterval(() => {
    if (isActive) fetchLogs()
  }, 5000)
}

function stopPolling() {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
}

watch(logContent_text, () => {
  setTimeout(scrollToBottom, 50)
})

onMounted(() => {
  fetchLogs()
  startPolling()
})

onActivated(() => {
  isActive = true
  fetchLogs()
})

onDeactivated(() => {
  isActive = false
})

onUnmounted(() => {
  stopPolling()
})
</script>

<style scoped>
.logs-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
}

.logs-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  flex-shrink: 0;
}

.logs-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logs-title h2 {
  margin: 0;
  font-size: 18px;
}

.log-file-name {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.logs-content {
  flex: 1;
  overflow: auto;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  min-height: 0;
}

.logs-content pre {
  margin: 0;
  padding: 16px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
  color: var(--el-text-color-primary);
}
</style>
