<template>
  <div class="tasks">
    <div class="tasks-header">
      <h2>游客链接管理</h2>
      <div class="header-actions">
        <el-button @click="refreshData" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button type="primary" @click="showCreate = true">
          <el-icon><Link /></el-icon>
          创建游客链接
        </el-button>
      </div>
    </div>
    
    <el-card shadow="hover" class="table-card">
      <div class="table-wrapper" ref="tableWrapper">
        <el-table :data="displayTasks" border height="100%" v-loading="loading">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="code" label="识别码" width="120" />
          <el-table-column prop="max_count" label="上传上限" width="100" />
          <el-table-column prop="uploaded_count" label="已上传" width="100" />
          <el-table-column prop="expire_time" label="到期时间" width="180">
            <template #default="scope">
              {{ formatTime(scope.row.expire_time) }}
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100">
            <template #default="scope">
              <el-tag :type="getStatusType(scope.row)">
                {{ getStatusText(scope.row) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" width="180">
            <template #default="scope">
              {{ formatTime(scope.row.created_at) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="220">
            <template #default="scope">
              <el-button type="text" @click="copyLink(scope.row)">复制链接</el-button>
              <el-button type="text" @click="editTask(scope.row)">编辑</el-button>
              <el-button type="text" @click="confirmDeleteTask(scope.row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
        <!-- 滚动加载提示 -->
        <div v-if="pageSize === 0 && hasMore" class="scroll-loading">
          <el-icon class="is-loading"><Loading /></el-icon>
          <span>加载中...</span>
        </div>
        <div v-if="pageSize === 0 && !hasMore && displayTasks.length > 0" class="scroll-loading">
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

    <el-dialog title="创建游客链接" v-model="showCreate">
      <el-form :model="form" label-width="120px">
        <el-form-item label="上传上限">
          <el-input v-model.number="form.max_count" />
        </el-form-item>
        <el-form-item label="有效期(天)">
          <el-input v-model.number="form.expire_days" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleCreate">创建</el-button>
          <el-button @click="showCreate = false">取消</el-button>
        </el-form-item>
      </el-form>
    </el-dialog>

    <el-dialog title="编辑游客链接" v-model="showEdit">
      <el-form :model="editForm" label-width="120px">
        <el-form-item label="上传上限">
          <el-input v-model.number="editForm.max_count" />
        </el-form-item>
        <el-form-item label="有效期(天)">
          <el-input v-model.number="editForm.expire_days" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleEdit">保存</el-button>
          <el-button @click="showEdit = false">取消</el-button>
        </el-form-item>
      </el-form>
    </el-dialog>

    <el-dialog title="删除游客链接" v-model="showDeleteConfirm" width="400px">
      <p>确定删除这个游客链接吗？</p>
      <el-checkbox v-model="deleteWithFiles" style="margin-top: 15px;">同步删除已上传的文件</el-checkbox>
      <template #footer>
        <el-button @click="showDeleteConfirm = false">取消</el-button>
        <el-button type="danger" @click="handleDelete">确认删除</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, onActivated, onDeactivated, nextTick } from 'vue'
import { Link, Loading, Refresh } from '@element-plus/icons-vue'
import axios from '@/utils/axios'
import { ElMessage } from 'element-plus'
import { copyToClipboard } from '@/utils/clipboard'

const allTasks = ref<any[]>([])
const displayTasks = ref<any[]>([])
const showCreate = ref(false)
const showEdit = ref(false)
const showDeleteConfirm = ref(false)
const editingTaskId = ref(0)
const deletingTaskId = ref(0)
const deleteWithFiles = ref(false)
const loading = ref(false)
const total = ref(0)

// 分页
const currentPage = ref(1)
const savedPageSize = localStorage.getItem('tasks_page_size')
const pageSize = ref(savedPageSize ? parseInt(savedPageSize) : 20)

// 滚动加载状态
const scrollOffset = ref(0)
const hasMore = ref(true)
const batchSize = 30
const tableWrapper = ref<HTMLElement>()
let scrollHandler: ((e: Event) => void) | null = null

const form = reactive({
  max_count: 5,
  expire_days: 7
})

const editForm = reactive({
  max_count: 5,
  expire_days: 7
})

function formatTime(timeStr: string): string {
  if (!timeStr) return ''
  const d = new Date(timeStr)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function getStatusType(task: any): string {
  if (task.status !== 1) return 'danger'
  if (new Date(task.expire_time) < new Date()) return 'warning'
  if (task.uploaded_count >= task.max_count) return 'info'
  return 'success'
}

function getStatusText(task: any): string {
  if (task.status !== 1) return '已禁用'
  if (new Date(task.expire_time) < new Date()) return '已过期'
  if (task.uploaded_count >= task.max_count) return '已满'
  return '有效'
}

function handlePageSizeChange(size: number) {
  localStorage.setItem('tasks_page_size', String(size))
  currentPage.value = 1
  scrollOffset.value = 0
  hasMore.value = true
  loadData()
}

function handlePageChange() {
  loadData()
}

async function loadData() {
  loading.value = true
  try {
    if (pageSize.value > 0) {
      const res = await axios.get('/api/admin/tasks', {
        params: { page: currentPage.value, page_size: pageSize.value }
      })
      displayTasks.value = res.data.data || []
      total.value = res.data.total || 0
    } else {
      const res = await axios.get('/api/admin/tasks', {
        params: { offset: 0, limit: batchSize }
      })
      allTasks.value = res.data.data || []
      total.value = res.data.total || 0
      displayTasks.value = allTasks.value
      scrollOffset.value = allTasks.value.length
      hasMore.value = allTasks.value.length < total.value
      setupScrollListener()
    }
  } catch (error) {
    console.error('Failed to load tasks:', error)
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  if (loading.value || !hasMore.value || pageSize.value > 0) return
  loading.value = true
  try {
    const res = await axios.get('/api/admin/tasks', {
      params: { offset: scrollOffset.value, limit: batchSize }
    })
    const newTasks = res.data.data || []
    allTasks.value = [...allTasks.value, ...newTasks]
    displayTasks.value = allTasks.value
    scrollOffset.value = allTasks.value.length
    hasMore.value = allTasks.value.length < total.value
  } catch (error) {
    console.error('Failed to load more tasks:', error)
  } finally {
    loading.value = false
  }
}

function setupScrollListener() {
  if (scrollHandler) return
  setTimeout(() => {
    const wrapper = tableWrapper.value
    if (!wrapper) return
    const tableBody = wrapper.querySelector('.el-table__body-wrapper')
    if (!tableBody) return
    scrollHandler = () => {
      const el = tableBody as HTMLElement
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

function editTask(task: any) {
  editingTaskId.value = task.id
  editForm.max_count = task.max_count
  const remainingDays = Math.ceil((new Date(task.expire_time).getTime() - Date.now()) / (1000 * 60 * 60 * 24))
  editForm.expire_days = Math.max(1, remainingDays)
  showEdit.value = true
}

function confirmDeleteTask(task: any) {
  deletingTaskId.value = task.id
  deleteWithFiles.value = false
  showDeleteConfirm.value = true
}

async function copyLink(task: any) {
  const fullUrl = window.location.origin + '/upload/' + task.code
  const ok = await copyToClipboard(fullUrl)
  if (ok) {
    ElMessage.success('链接已复制')
  } else {
    ElMessage.error('复制失败')
  }
}

async function handleCreate() {
  try {
    await axios.post('/api/admin/tasks', form)
    ElMessage.success('创建成功')
    showCreate.value = false
    form.max_count = 5
    form.expire_days = 7
    loadData()
  } catch (error) {
    ElMessage.error('创建失败')
  }
}

async function handleEdit() {
  try {
    await axios.put(`/api/admin/tasks/${editingTaskId.value}`, editForm)
    ElMessage.success('保存成功')
    showEdit.value = false
    loadData()
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

async function handleDelete() {
  try {
    await axios.delete(`/api/admin/tasks/${deletingTaskId.value}?delete_files=${deleteWithFiles.value}`)
    ElMessage.success('删除成功')
    showDeleteConfirm.value = false
    loadData()
  } catch (error) {
    ElMessage.error('删除失败')
  }
}

function refreshData() {
  currentPage.value = 1
  scrollOffset.value = 0
  hasMore.value = true
  loadData()
}

onMounted(() => {
  loadData()
})

onActivated(() => {
  loadData()
  if (pageSize.value === 0) {
    nextTick(() => setupScrollListener())
  }
})

onDeactivated(() => {
  removeScrollListener()
})

onUnmounted(() => {
  removeScrollListener()
})
</script>

<style scoped>
.tasks {
  display: flex;
  flex-direction: column;
  min-height: 0;
  flex: 1;
}

.tasks-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-shrink: 0;
}

.tasks-header h2 {
  margin: 0;
  color: var(--el-text-color-primary);
}

.header-actions {
  display: flex;
  gap: 8px;
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

/* 手机端适配 */
@media (max-width: 768px) {
  .tasks-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .header-actions {
    width: 100%;
  }

  .pagination-wrapper {
    flex-wrap: wrap;
    justify-content: center;
    gap: 8px;
  }

  .table-wrapper :deep(.el-table) {
    font-size: 12px;
  }

  .table-wrapper :deep(.el-table .el-button--text) {
    padding: 4px 6px;
    font-size: 12px;
  }
}
</style>
