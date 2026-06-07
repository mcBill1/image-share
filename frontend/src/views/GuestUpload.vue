<template>
  <div class="guest-upload">
    <el-card class="upload-card" shadow="hover">
      <div class="upload-header">
        <h2>游客上传</h2>
        <p v-if="task">已上传 / 最大上传：{{ task.uploaded_count }} / {{ task.max_count }}</p>
      </div>
      
      <div v-if="!canView" class="invalid-message">
        <el-alert type="error" title="链接已失效" show-icon />
      </div>
      
      <div v-else>
        <div v-if="canUpload">
          <el-upload
            class="upload-demo"
            :action="uploadUrl"
            :on-success="handleUploadSuccess"
            :on-error="handleUploadError"
            :file-list="uploadFiles"
            accept="image/*"
            multiple
          >
            <el-button type="primary" size="large">
              <el-icon><Upload /></el-icon>
              选择图片上传
            </el-button>
          </el-upload>
        </div>
        <div v-else>
          <el-alert type="warning" title="上传次数已满，仅可查看已上传的图片" show-icon style="margin-bottom: 20px;" />
        </div>
        
        <div v-if="uploadedImages.length > 0" class="uploaded-list">
          <h3>已上传图片</h3>
          <div class="image-grid">
            <div v-for="image in uploadedImages" :key="image.id" class="image-item">
              <img :src="image.public_url" class="preview-image" />
              <div class="image-actions">
                <el-button type="text" size="small" @click="copyLink(image)">复制链接</el-button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { Upload } from '@element-plus/icons-vue'
import axios from '@/utils/axios'
import { ElMessage } from 'element-plus'

const route = useRoute()
const code = computed(() => route.params.code as string)

const canView = ref(false)
const canUpload = ref(false)
const task = ref<any>(null)
const uploadFiles = ref<any[]>([])
const uploadedImages = ref<any[]>([])

const uploadUrl = computed(() => `/api/upload/${code.value}`)

async function loadTaskInfo() {
  try {
    const response = await axios.get(`/api/guest/${code.value}`)
    canView.value = true
    task.value = response.data.task
    uploadedImages.value = response.data.images || []
    // 检查是否可上传
    canUpload.value = task.value.uploaded_count < task.value.max_count
  } catch (error) {
    canView.value = false
    canUpload.value = false
  }
}

function handleUploadSuccess(response: any) {
  ElMessage.success('上传成功')
  uploadFiles.value = []
  uploadedImages.value.push(response)
  loadTaskInfo()
}

function handleUploadError(error: any) {
  ElMessage.error(error.response?.data?.message || error.response?.data?.error || '上传失败')
}

async function copyLink(image: any) {
  const fullUrl = window.location.origin + image.public_url
  try {
    await navigator.clipboard.writeText(fullUrl)
    ElMessage.success('链接已复制')
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

onMounted(() => {
  loadTaskInfo()
})
</script>

<style scoped>
.guest-upload {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.upload-card {
  width: 600px;
  padding: 40px;
}

.upload-header {
  text-align: center;
  margin-bottom: 30px;
}

.upload-header h2 {
  margin: 0 0 10px 0;
  color: var(--el-text-color-primary);
}

.upload-header p {
  margin: 0;
  color: var(--el-text-color-secondary);
}

.invalid-message {
  text-align: center;
}

.uploaded-list {
  margin-top: 30px;
}

.uploaded-list h3 {
  margin: 0 0 15px 0;
  color: var(--el-text-color-secondary);
}

.image-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 15px;
}

.image-item {
  position: relative;
}

.preview-image {
  width: 100%;
  height: 120px;
  object-fit: cover;
  border-radius: 4px;
}

.image-actions {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  background: rgba(0, 0, 0, 0.6);
  padding: 5px;
  text-align: center;
  border-radius: 0 0 4px 4px;
}

.image-actions button {
  color: #fff;
}
</style>
