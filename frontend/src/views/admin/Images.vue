<template>
  <div class="images">
    <div class="images-header">
      <h2>图片管理</h2>
      <el-button type="primary" @click="showUpload = true">
        <el-icon><Upload /></el-icon>
        上传图片
      </el-button>
    </div>

    <el-card shadow="hover" class="table-card">
      <!-- 筛选栏 -->
      <div class="filter-bar">
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          value-format="YYYY-MM-DD"
          @change="applyFilters"
          style="width: 300px; margin-right: 12px"
        />
        <el-select v-model="ownerFilter" placeholder="按类型筛选" clearable @change="applyFilters" style="width: 140px; margin-right: 12px">
          <el-option label="管理员" value="admin" />
          <el-option label="用户" value="user" />
          <el-option label="游客" value="guest" />
        </el-select>
        <el-input v-model="ownerNameFilter" placeholder="用户名/游客识别码" clearable @clear="applyFilters" @keyup.enter="applyFilters" style="width: 180px; margin-right: 12px">
          <template #append>
            <el-button @click="applyFilters">筛选</el-button>
          </template>
        </el-input>
        <el-button @click="resetFilters">重置</el-button>
      </div>

      <div class="table-wrapper" ref="tableWrapper">
        <el-table :data="displayImages" border height="100%" v-loading="loading">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="original_name" label="原始名称" show-overflow-tooltip />
          <el-table-column prop="file_size" label="大小" width="100">
            <template #default="scope">
              {{ formatSize(scope.row.file_size) }}
            </template>
          </el-table-column>
          <el-table-column prop="width" label="分辨率" width="120">
            <template #default="scope">
              {{ scope.row.width && scope.row.height ? scope.row.width + ' × ' + scope.row.height : '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="owner_type" label="类型" width="150">
            <template #default="scope">
              <el-tag :type="getOwnerTypeTag(scope.row.owner_type)">
                {{ getOwnerTypeText(scope.row) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="upload_time" label="上传时间" width="180">
            <template #default="scope">
              {{ formatTime(scope.row.upload_time) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200">
            <template #default="scope">
              <el-button type="text" @click="previewImage(scope.row)">预览</el-button>
              <el-button type="text" @click="copyLink(scope.row)">复制链接</el-button>
              <el-button type="text" @click="deleteImage(scope.row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
        <!-- 滚动加载提示 -->
        <div v-if="pageSize === 0 && hasMore" class="scroll-loading">
          <el-icon class="is-loading"><Loading /></el-icon>
          <span>加载中...</span>
        </div>
        <div v-if="pageSize === 0 && !hasMore && displayImages.length > 0" class="scroll-loading">
          <span>已加载全部 {{ total }} 条</span>
        </div>
      </div>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <span class="pagination-total">共 {{ total }} 条</span>
        <el-select v-model="pageSize" style="width: 120px" @change="handlePageSizeChange">
          <el-option :value="10" label="10 条/页" />
          <el-option :value="20" label="20 条/页" />
          <el-option :value="30" label="30 条/页" />
          <el-option :value="50" label="50 条/页" />
          <el-option :value="0" label="不分页" />
        </el-select>
        <el-pagination
          v-if="pageSize > 0"
          v-model:current-page="currentPage"
          :page-size="pageSize"
          :total="total"
          layout="prev, pager, next"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>

    <el-dialog title="上传图片" v-model="showUpload">
      <el-upload
        class="upload-demo"
        action="/api/admin/upload"
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
    </el-dialog>

    <el-dialog title="图片预览" v-model="showPreview" width="600px">
      <img :src="previewImageUrl" style="width: 100%" />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, onActivated, onDeactivated, nextTick } from 'vue'
import { Upload, Loading } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'
import { ElMessage } from 'element-plus'
import { copyToClipboard } from '@/utils/clipboard'

const authStore = useAuthStore()

const allImages = ref<any[]>([])
const displayImages = ref<any[]>([])
const showUpload = ref(false)
const showPreview = ref(false)
const previewImageUrl = ref('')
const uploadFiles = ref<any[]>([])
const loading = ref(false)
const total = ref(0)

// 筛选
const dateRange = ref<string[]>([])
const ownerFilter = ref('')
const ownerNameFilter = ref('')
const useFilter = ref(false)

// 分页：0=不分页
const currentPage = ref(1)
const savedPageSize = localStorage.getItem('admin_page_size')
const pageSize = ref(savedPageSize ? parseInt(savedPageSize) : 20)

// 滚动加载状态
const scrollOffset = ref(0)
const hasMore = ref(true)
const batchSize = 30
const tableWrapper = ref<HTMLElement>()
let scrollHandler: ((e: Event) => void) | null = null

// 轮询
let timer: ReturnType<typeof setInterval> | null = null
let isActive = false

function formatSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

function formatTime(timeStr: string): string {
  if (!timeStr) return ''
  const d = new Date(timeStr)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function getOwnerTypeTag(type: string): string {
  switch (type) {
    case 'admin': return 'primary'
    case 'user': return 'success'
    case 'guest': return 'warning'
    default: return 'info'
  }
}

function getOwnerTypeText(row: any): string {
  switch (row.owner_type) {
    case 'admin': return '管理员'
    case 'user': return '用户：' + (row.owner_name || row.owner_id)
    case 'guest': return '游客：' + (row.task_code || row.owner_id)
    default: return row.owner_type
  }
}

// 筛选逻辑 - 前端筛选（因为后端暂不支持筛选分页）
const filteredImages = computed(() => {
  let result = [...allImages.value]

  if (dateRange.value && dateRange.value.length === 2) {
    const start = new Date(dateRange.value[0])
    const end = new Date(dateRange.value[1])
    end.setHours(23, 59, 59, 999)
    result = result.filter((img: any) => {
      const t = new Date(img.upload_time)
      return t >= start && t <= end
    })
  }

  if (ownerFilter.value) {
    result = result.filter((img: any) => img.owner_type === ownerFilter.value)
  }

  if (ownerNameFilter.value) {
    const filterVal = ownerNameFilter.value.trim().toLowerCase()
    result = result.filter((img: any) => {
      if (img.owner_type === 'user') {
        return (img.owner_name && img.owner_name.toLowerCase().includes(filterVal))
      }
      if (img.owner_type === 'guest') {
        return (img.task_code && img.task_code.toLowerCase().includes(filterVal))
      }
      return false
    })
  }

  return result
})

function applyFilters() {
  useFilter.value = !!(dateRange.value?.length || ownerFilter.value || ownerNameFilter.value)
  currentPage.value = 1
  if (useFilter.value) {
    // 有筛选条件时使用前端筛选
    displayImages.value = filteredImages.value
    total.value = filteredImages.value.length
  } else {
    // 无筛选条件时使用后端分页
    loadData()
  }
}

function resetFilters() {
  dateRange.value = []
  ownerFilter.value = ''
  ownerNameFilter.value = ''
  useFilter.value = false
  currentPage.value = 1
  loadData()
}

function handlePageSizeChange(size: number) {
  localStorage.setItem('admin_page_size', String(size))
  currentPage.value = 1
  scrollOffset.value = 0
  hasMore.value = true
  loadData()
}

function handlePageChange() {
  loadData()
}

async function loadData() {
  if (useFilter.value) {
    // 筛选模式下使用前端数据
    if (pageSize.value > 0) {
      const start = (currentPage.value - 1) * pageSize.value
      displayImages.value = filteredImages.value.slice(start, start + pageSize.value)
    } else {
      displayImages.value = filteredImages.value
    }
    total.value = filteredImages.value.length
    return
  }

  loading.value = true
  try {
    if (pageSize.value > 0) {
      // 分页模式：只加载当前页
      const res = await axios.get('/api/admin/images', {
        params: { page: currentPage.value, page_size: pageSize.value }
      })
      displayImages.value = res.data.data || []
      total.value = res.data.total || 0
    } else {
      // 不分页模式：滚动加载
      const res = await axios.get('/api/admin/images', {
        params: { offset: 0, limit: batchSize }
      })
      allImages.value = res.data.data || []
      total.value = res.data.total || 0
      displayImages.value = allImages.value
      scrollOffset.value = allImages.value.length
      hasMore.value = allImages.value.length < total.value
      // 数据加载完后绑定滚动监听
      setupScrollListener()
    }
  } catch (error) {
    console.error('Failed to load images:', error)
  } finally {
    loading.value = false
  }
}

// 滚动加载更多
async function loadMore() {
  if (loading.value || !hasMore.value || pageSize.value > 0 || useFilter.value) return
  loading.value = true
  try {
    const res = await axios.get('/api/admin/images', {
      params: { offset: scrollOffset.value, limit: batchSize }
    })
    const newImages = res.data.data || []
    allImages.value = [...allImages.value, ...newImages]
    displayImages.value = allImages.value
    scrollOffset.value = allImages.value.length
    hasMore.value = allImages.value.length < total.value
  } catch (error) {
    console.error('Failed to load more images:', error)
  } finally {
    loading.value = false
  }
}

function setupScrollListener() {
  if (scrollHandler) return
  // 延迟确保 el-table 内部 DOM 已渲染
  setTimeout(() => {
    const wrapper = tableWrapper.value
    if (!wrapper) return
    const tableBody = wrapper.querySelector('.el-table__body-wrapper')
    if (!tableBody) return
    scrollHandler = () => {
      const el = tableBody as HTMLElement
      // 距离底部小于200px时预加载
      if (el.scrollHeight - el.scrollTop - el.clientHeight < 200 && hasMore.value && !loading.value) {
        loadMore()
      }
    }
    tableBody.addEventListener('scroll', scrollHandler)
  }, 300)
}

function removeScrollListener() {
  if (scrollHandler) {
    const wrapper = tableWrapper.value
    if (wrapper) {
      const tableBody = wrapper.querySelector('.el-table__body-wrapper')
      if (tableBody) {
        tableBody.removeEventListener('scroll', scrollHandler!)
      }
    }
    scrollHandler = null
  }
}

function previewImage(image: any) {
  previewImageUrl.value = image.public_url
  showPreview.value = true
}

async function copyLink(image: any) {
  const fullUrl = window.location.origin + image.public_url
  const ok = await copyToClipboard(fullUrl)
  if (ok) {
    ElMessage.success('链接已复制')
  } else {
    ElMessage.error('复制失败')
  }
}

async function deleteImage(id: number) {
  if (!confirm('确定删除这张图片吗？')) return
  try {
    await axios.delete(`/api/admin/images/${id}`)
    ElMessage.success('删除成功')
    loadData()
  } catch (error) {
    ElMessage.error('删除失败')
  }
}

function handleUploadSuccess() {
  ElMessage.success('上传成功')
  showUpload.value = false
  uploadFiles.value = []
  loadData()
}

function handleUploadError() {
  ElMessage.error('上传失败')
}

function startPolling() {
  if (timer) clearInterval(timer)
  timer = setInterval(() => {
    if (isActive) loadData()
  }, 5000)
}

function stopPolling() {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
}

onMounted(() => {
  loadData()
  startPolling()
})

onActivated(() => {
  isActive = true
  loadData()
  if (pageSize.value === 0) {
    nextTick(() => setupScrollListener())
  }
})

onDeactivated(() => {
  isActive = false
  removeScrollListener()
})

onUnmounted(() => {
  stopPolling()
  removeScrollListener()
})
</script>

<style scoped>
.images {
  display: flex;
  flex-direction: column;
  min-height: 0;
  flex: 1;
}

.images-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-shrink: 0;
}

.images-header h2 {
  margin: 0;
  color: var(--el-text-color-primary);
}

.table-card {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.table-card :deep(.el-card__body) {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.filter-bar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  margin-bottom: 16px;
  flex-shrink: 0;
}

.table-wrapper {
  flex: 1;
  min-height: 0;
  position: relative;
}

.scroll-loading {
  text-align: center;
  padding: 8px;
  color: var(--el-text-color-secondary);
  font-size: 13px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  margin-top: 16px;
  flex-shrink: 0;
}

.pagination-total {
  color: var(--el-text-color-secondary);
  font-size: 13px;
  margin-right: 12px;
}
</style>
