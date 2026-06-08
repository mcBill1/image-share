<template>
  <div class="users">
    <div class="users-header">
      <h2>用户管理</h2>
      <div class="header-actions">
        <el-button @click="refreshData" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button type="primary" @click="showCreate = true">
          <el-icon><Plus /></el-icon>
          创建用户
        </el-button>
      </div>
    </div>
    
    <el-card shadow="hover" class="table-card">
      <div class="table-wrapper" ref="tableWrapper">
        <el-table :data="displayUsers" border height="100%" v-loading="loading">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="username" label="用户名" />
          <el-table-column prop="role" label="角色" width="100">
            <template #default="scope">
              <el-tag :type="scope.row.role === 'admin' ? 'danger' : 'success'">
                {{ scope.row.role === 'admin' ? '管理员' : '用户' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="storage_limit_mb" label="空间限制" width="120">
            <template #default="scope">
              {{ scope.row.role === 'admin' ? '-' : scope.row.storage_limit_mb + ' MB' }}
            </template>
          </el-table-column>
          <el-table-column prop="image_limit" label="图片数量" width="100">
            <template #default="scope">
              {{ scope.row.role === 'admin' ? '-' : scope.row.image_limit }}
            </template>
          </el-table-column>
          <el-table-column prop="single_image_limit_mb" label="单图限制" width="120">
            <template #default="scope">
              {{ scope.row.role === 'admin' ? '-' : scope.row.single_image_limit_mb + ' MB' }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="250">
            <template #default="scope">
              <el-button type="text" @click="openResetPassword(scope.row)">重置密码</el-button>
              <template v-if="scope.row.role !== 'admin'">
                <el-button type="text" @click="editUser(scope.row)">编辑</el-button>
                <el-button type="text" @click="deleteUser(scope.row.id)">删除</el-button>
              </template>
            </template>
          </el-table-column>
        </el-table>
        <!-- 滚动加载提示 -->
        <div v-if="pageSize === 0 && hasMore" class="scroll-loading">
          <el-icon class="is-loading"><Loading /></el-icon>
          <span>加载中...</span>
        </div>
        <div v-if="pageSize === 0 && !hasMore && displayUsers.length > 0" class="scroll-loading">
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

    <el-dialog title="创建用户" v-model="showCreate">
      <el-form :model="form" label-width="120px">
        <el-form-item label="用户名">
          <el-input v-model="form.username" placeholder="2-10位，仅字母数字下划线" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" placeholder="至少6位，必须包含字母" />
        </el-form-item>
        <el-form-item label="空间限制(MB)">
          <el-input v-model.number="form.storage_limit_mb" />
        </el-form-item>
        <el-form-item label="图片数量限制">
          <el-input v-model.number="form.image_limit" />
        </el-form-item>
        <el-form-item label="单图限制(MB)">
          <el-input v-model.number="form.single_image_limit_mb" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleCreate">创建</el-button>
          <el-button @click="showCreate = false">取消</el-button>
        </el-form-item>
      </el-form>
    </el-dialog>

    <el-dialog title="编辑用户" v-model="showEdit">
      <el-form :model="editForm" label-width="120px">
        <el-form-item label="用户名">
          <el-input v-model="editForm.username" placeholder="2-10位，仅字母数字下划线" />
        </el-form-item>
        <el-form-item label="空间限制(MB)">
          <el-input v-model.number="editForm.storage_limit_mb" />
        </el-form-item>
        <el-form-item label="图片数量限制">
          <el-input v-model.number="editForm.image_limit" />
        </el-form-item>
        <el-form-item label="单图限制(MB)">
          <el-input v-model.number="editForm.single_image_limit_mb" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleEdit">保存</el-button>
          <el-button @click="showEdit = false">取消</el-button>
        </el-form-item>
      </el-form>
    </el-dialog>

    <el-dialog title="重置密码" v-model="showResetPassword" width="400px">
      <p>为用户 <strong>{{ resettingUsername }}</strong> 设置新密码</p>
      <el-input v-model="resetForm.newPassword" type="password" placeholder="至少6位，必须包含字母" style="margin-top: 15px;" />
      <template #footer>
        <el-button @click="showResetPassword = false">取消</el-button>
        <el-button type="primary" @click="handleResetPassword">确认重置</el-button>
      </template>
    </el-dialog>

    <el-dialog title="删除用户" v-model="showDeleteConfirm" width="400px">
      <p>确定删除这个用户吗？</p>
      <el-checkbox v-model="deleteWithFiles" style="margin-top: 15px;">同步删除已上传的文件</el-checkbox>
      <template #footer>
        <el-button @click="showDeleteConfirm = false">取消</el-button>
        <el-button type="danger" @click="handleDeleteUser">确认删除</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, onActivated, onDeactivated, nextTick } from 'vue'
import { Plus, Loading, Refresh } from '@element-plus/icons-vue'
import axios from '@/utils/axios'
import { ElMessage } from 'element-plus'
import { md5 } from 'js-md5'

const allUsers = ref<any[]>([])
const displayUsers = ref<any[]>([])
const showCreate = ref(false)
const showEdit = ref(false)
const showResetPassword = ref(false)
const editingUserId = ref(0)
const resettingUserId = ref(0)
const loading = ref(false)
const total = ref(0)

// 分页
const currentPage = ref(1)
const savedPageSize = localStorage.getItem('users_page_size')
const pageSize = ref(savedPageSize ? parseInt(savedPageSize) : 20)

// 滚动加载状态
const scrollOffset = ref(0)
const hasMore = ref(true)
const batchSize = 30
const tableWrapper = ref<HTMLElement>()
let scrollHandler: ((e: Event) => void) | null = null

const form = reactive({
  username: '',
  password: '',
  storage_limit_mb: 100,
  image_limit: 50,
  single_image_limit_mb: 10
})

const editForm = reactive({
  username: '',
  storage_limit_mb: 100,
  image_limit: 50,
  single_image_limit_mb: 10
})

const resetForm = reactive({
  newPassword: ''
})
const resettingUsername = ref('')

const showDeleteConfirm = ref(false)
const deletingUserId = ref(0)
const deleteWithFiles = ref(false)

function handlePageSizeChange(size: number) {
  localStorage.setItem('users_page_size', String(size))
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
      const res = await axios.get('/api/admin/users', {
        params: { page: currentPage.value, page_size: pageSize.value }
      })
      displayUsers.value = res.data.data || []
      total.value = res.data.total || 0
    } else {
      const res = await axios.get('/api/admin/users', {
        params: { offset: 0, limit: batchSize }
      })
      allUsers.value = res.data.data || []
      total.value = res.data.total || 0
      displayUsers.value = allUsers.value
      scrollOffset.value = allUsers.value.length
      hasMore.value = allUsers.value.length < total.value
      setupScrollListener()
    }
  } catch (error) {
    console.error('Failed to load users:', error)
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  if (loading.value || !hasMore.value || pageSize.value > 0) return
  loading.value = true
  try {
    const res = await axios.get('/api/admin/users', {
      params: { offset: scrollOffset.value, limit: batchSize }
    })
    const newUsers = res.data.data || []
    allUsers.value = [...allUsers.value, ...newUsers]
    displayUsers.value = allUsers.value
    scrollOffset.value = allUsers.value.length
    hasMore.value = allUsers.value.length < total.value
  } catch (error) {
    console.error('Failed to load more users:', error)
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

function editUser(user: any) {
  editingUserId.value = user.id
  editForm.username = user.username
  editForm.storage_limit_mb = user.storage_limit_mb
  editForm.image_limit = user.image_limit
  editForm.single_image_limit_mb = user.single_image_limit_mb
  showEdit.value = true
}

function openResetPassword(user: any) {
  resettingUserId.value = user.id
  resettingUsername.value = user.username
  resetForm.newPassword = ''
  showResetPassword.value = true
}

async function handleResetPassword() {
  if (!resetForm.newPassword) {
    ElMessage.error('请输入新密码')
    return
  }
  if (resetForm.newPassword.length < 6) {
    ElMessage.error('密码至少需要6位')
    return
  }
  if (!/[a-zA-Z]/.test(resetForm.newPassword)) {
    ElMessage.error('密码必须包含至少一个字母')
    return
  }
  try {
    await axios.put(`/api/admin/users/${resettingUserId.value}/password`, {
      new_password: md5(resetForm.newPassword)
    })
    ElMessage.success('密码重置成功')
    showResetPassword.value = false
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '重置失败')
  }
}

async function handleCreate() {
  if (!form.username || !form.password) {
    ElMessage.error('请填写用户名和密码')
    return
  }
  if (form.username.length < 2 || form.username.length > 10 || !/^[a-zA-Z0-9_]+$/.test(form.username)) {
    ElMessage.error('用户名必须为2-10位，仅允许字母、数字和下划线')
    return
  }
  if (form.password.length < 6) {
    ElMessage.error('密码至少需要6位')
    return
  }
  if (!/[a-zA-Z]/.test(form.password)) {
    ElMessage.error('密码必须包含至少一个字母')
    return
  }
  try {
    await axios.post('/api/admin/users', {
      ...form,
      password: md5(form.password)
    })
    ElMessage.success('创建成功')
    showCreate.value = false
    form.username = ''
    form.password = ''
    form.storage_limit_mb = 100
    form.image_limit = 50
    form.single_image_limit_mb = 10
    loadData()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '创建失败')
  }
}

async function handleEdit() {
  if (!editForm.username) {
    ElMessage.error('请填写用户名')
    return
  }
  if (editForm.username.length < 2 || editForm.username.length > 10 || !/^[a-zA-Z0-9_]+$/.test(editForm.username)) {
    ElMessage.error('用户名必须为2-10位，仅允许字母、数字和下划线')
    return
  }
  try {
    await axios.put(`/api/admin/users/${editingUserId.value}`, editForm)
    ElMessage.success('保存成功')
    showEdit.value = false
    loadData()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '保存失败')
  }
}

function deleteUser(id: number) {
  deletingUserId.value = id
  deleteWithFiles.value = false
  showDeleteConfirm.value = true
}

async function handleDeleteUser() {
  try {
    await axios.delete(`/api/admin/users/${deletingUserId.value}?delete_files=${deleteWithFiles.value}`)
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
.users {
  display: flex;
  flex-direction: column;
  min-height: 0;
  flex: 1;
}

.users-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-shrink: 0;
}

.users-header h2 {
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
  .users-header {
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
