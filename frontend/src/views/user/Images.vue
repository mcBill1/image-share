<template>
  <div class="user-images">
    <div class="images-header">
      <h2>图片管理</h2>
    </div>

    <el-card shadow="hover" class="table-card">
      <!-- 日期筛选 -->
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
        <el-button @click="resetFilters">重置</el-button>
      </div>

      <div class="table-wrapper">
        <el-table :data="pagedImages" border height="100%">
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
          <el-table-column prop="upload_time" label="上传时间" width="180">
            <template #default="scope">
              {{ formatTime(scope.row.upload_time) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="180">
            <template #default="scope">
              <el-button type="text" @click="previewImage(scope.row)">预览</el-button>
              <el-button type="text" @click="copyLink(scope.row)">复制链接</el-button>
              <el-button type="text" @click="deleteImage(scope.row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <span class="pagination-total">共 {{ filteredImages.length }} 条</span>
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
          :total="filteredImages.length"
          layout="prev, pager, next"
        />
      </div>
    </el-card>

    <el-dialog title="图片预览" v-model="showPreview" width="600px">
      <img :src="previewImageUrl" style="width: 100%" />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import axios from '@/utils/axios'
import { ElMessage } from 'element-plus'
import { copyToClipboard } from '@/utils/clipboard'

const images = ref<any[]>([])
const showPreview = ref(false)
const previewImageUrl = ref('')

// 筛选
const dateRange = ref<string[]>([])

// 分页：0=不分页
const currentPage = ref(1)
const savedPageSize = localStorage.getItem('user_page_size')
const pageSize = ref(savedPageSize ? parseInt(savedPageSize) : 20)

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

// 筛选逻辑
const filteredImages = computed(() => {
  let result = [...images.value]

  if (dateRange.value && dateRange.value.length === 2) {
    const start = new Date(dateRange.value[0])
    const end = new Date(dateRange.value[1])
    end.setHours(23, 59, 59, 999)
    result = result.filter((img: any) => {
      const t = new Date(img.upload_time)
      return t >= start && t <= end
    })
  }

  return result
})

const pagedImages = computed(() => {
  if (pageSize.value === 0) return filteredImages.value
  const start = (currentPage.value - 1) * pageSize.value
  return filteredImages.value.slice(start, start + pageSize.value)
})

function applyFilters() {
  currentPage.value = 1
}

function resetFilters() {
  dateRange.value = []
  currentPage.value = 1
}

function handlePageSizeChange(size: number) {
  localStorage.setItem('user_page_size', String(size))
  currentPage.value = 1
}

async function loadImages() {
  try {
    const response = await axios.get('/api/user/images')
    images.value = response.data
  } catch (error) {
    console.error('Failed to load images:', error)
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
    await axios.delete(`/api/user/images/${id}`)
    images.value = images.value.filter(img => img.id !== id)
    ElMessage.success('删除成功')
  } catch (error) {
    ElMessage.error('删除失败')
  }
}

onMounted(() => {
  loadImages()
})
</script>

<style scoped>
.user-images {
  display: flex;
  flex-direction: column;
  min-height: 0;
  flex: 1;
}

.images-header {
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
