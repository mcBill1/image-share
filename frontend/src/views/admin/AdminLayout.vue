<template>
  <el-container class="admin-layout">
    <!-- 手机端遮罩 -->
    <div class="mobile-overlay" :class="{ active: !isCollapsed }" @click="isCollapsed = true"></div>
    <el-aside :width="isCollapsed ? '64px' : '220px'" class="admin-sidebar" :class="{ 'mobile-sidebar': isMobile, 'mobile-show': isMobile && !isCollapsed }">
      <div class="logo" :class="{ collapsed: isCollapsed && !isMobile }">
        <h2 v-show="!isCollapsed || isMobile">ImageShare</h2>
        <span v-show="isCollapsed && !isMobile" class="logo-icon">IS</span>
      </div>
      <div class="sidebar-menu">
        <div
          v-for="item in menuItems"
          :key="item.path"
          class="menu-item"
          :class="{ active: activeMenu === item.path }"
          @click="router.push(item.path); if (isMobile) isCollapsed = true"
        >
          <el-icon :size="20"><component :is="item.icon" /></el-icon>
          <span v-show="!isCollapsed || isMobile" class="menu-text">{{ item.label }}</span>
        </div>
      </div>
      <div class="collapse-btn" @click="isCollapsed = !isCollapsed">
        <el-icon :size="18">
          <Fold v-if="!isCollapsed" />
          <Expand v-else />
        </el-icon>
      </div>
    </el-aside>
    <el-container class="main-container">
      <el-header class="admin-header">
        <div class="header-left" v-if="isMobile">
          <el-button text @click="isCollapsed = false">
            <el-icon :size="20"><Expand /></el-icon>
          </el-button>
        </div>
        <div class="header-right">
          <el-dropdown trigger="click" @command="handleThemeChange">
            <el-button text class="theme-btn">
              <el-icon :size="18">
                <Sunny v-if="themeStore.mode === 'light'" />
                <Moon v-else-if="themeStore.mode === 'dark'" />
                <Monitor v-else />
              </el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="system" :class="{ 'is-active': themeStore.mode === 'system' }">
                  <el-icon><Monitor /></el-icon>跟随系统
                </el-dropdown-item>
                <el-dropdown-item command="light" :class="{ 'is-active': themeStore.mode === 'light' }">
                  <el-icon><Sunny /></el-icon>亮色模式
                </el-dropdown-item>
                <el-dropdown-item command="dark" :class="{ 'is-active': themeStore.mode === 'dark' }">
                  <el-icon><Moon /></el-icon>暗色模式
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <span class="username">{{ authStore.role === 'admin' ? '管理员' : '用户' }}</span>
          <el-button text @click="handleLogout">退出登录</el-button>
        </div>
      </el-header>
      <el-main class="admin-main">
        <router-view v-slot="{ Component }">
          <keep-alive>
            <component :is="Component" />
          </keep-alive>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useThemeStore } from '@/stores/theme'
import { Odometer, Picture, User, Link, Lock, Fold, Expand, Sunny, Moon, Monitor, Document } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import type { ThemeMode } from '@/stores/theme'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const themeStore = useThemeStore()
const isCollapsed = ref(false)
const isMobile = ref(false)

function checkMobile() {
  isMobile.value = window.innerWidth <= 768
  if (isMobile.value) {
    isCollapsed.value = true
  }
}

const activeMenu = computed(() => route.path)

const menuItems = [
  { path: '/admin/dashboard', icon: Odometer, label: '仪表盘' },
  { path: '/admin/images', icon: Picture, label: '图片管理' },
  { path: '/admin/users', icon: User, label: '用户管理' },
  { path: '/admin/tasks', icon: Link, label: '游客链接' },
  { path: '/admin/logs', icon: Document, label: '系统日志' },
  { path: '/profile/change-password', icon: Lock, label: '修改密码' },
]

function handleThemeChange(command: string) {
  themeStore.setTheme(command as ThemeMode)
}

async function handleLogout() {
  await authStore.logout()
  ElMessage.success('已退出登录')
  router.push('/login')
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})
</script>

<style scoped>
.admin-layout {
  height: 100vh;
  overflow: hidden;
}

.admin-sidebar {
  background: rgba(30, 41, 59, 0.95);
  backdrop-filter: blur(10px);
  color: #fff;
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease;
  overflow: hidden;
}

.logo {
  padding: 20px;
  text-align: center;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  min-height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s ease;
  overflow: hidden;
  white-space: nowrap;
}

.logo h2 {
  margin: 0;
  font-size: 18px;
  color: #fff;
  letter-spacing: 1px;
}

.logo.collapsed {
  padding: 20px 0;
}

.logo-icon {
  font-size: 18px;
  font-weight: bold;
  color: #60a5fa;
}

.sidebar-menu {
  flex: 1;
  padding: 8px 0;
  overflow-y: auto;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 4px 8px;
  padding: 0 20px;
  height: 44px;
  border-radius: 8px;
  color: rgba(255, 255, 255, 0.65);
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
  overflow: hidden;
}

.menu-item:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #fff;
}

.menu-item.active {
  background: rgba(96, 165, 250, 0.15);
  color: #60a5fa;
}

.menu-item .el-icon {
  flex-shrink: 0;
}

.menu-text {
  font-size: 14px;
  transition: opacity 0.2s ease;
}

.collapse-btn {
  padding: 16px;
  text-align: center;
  cursor: pointer;
  color: rgba(255, 255, 255, 0.5);
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  transition: all 0.2s ease;
}

.collapse-btn:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.06);
}

.main-container {
  background: #f0f2f5;
}

html.dark .main-container {
  background: #141414;
}

.admin-header {
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(10px);
  padding: 0 24px;
  display: flex;
  justify-content: flex-end;
  align-items: center;
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
  height: 56px;
}

html.dark .admin-header {
  background: rgba(30, 30, 30, 0.8);
  border-bottom-color: rgba(255, 255, 255, 0.06);
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.theme-btn {
  padding: 4px 8px;
}

.username {
  color: #666;
  font-size: 14px;
}

html.dark .username {
  color: #aaa;
}

.admin-main {
  padding: 20px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.admin-main > div {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.header-left {
  margin-right: auto;
}

/* 手机端适配 */
.mobile-overlay {
  display: none;
}

@media (max-width: 768px) {
  .mobile-overlay {
    display: block;
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    z-index: 99;
    opacity: 0;
    pointer-events: none;
    transition: opacity 0.3s ease;
  }

  .mobile-overlay.active {
    opacity: 1;
    pointer-events: auto;
  }

  .admin-sidebar.mobile-sidebar {
    position: fixed;
    top: 0;
    left: 0;
    bottom: 0;
    z-index: 100;
    transform: translateX(-100%);
    transition: transform 0.3s ease;
    width: 220px !important;
  }

  .admin-sidebar.mobile-show {
    transform: translateX(0);
  }

  .admin-header {
    padding: 0 12px;
  }

  .admin-main {
    padding: 12px;
  }

  .username {
    display: none;
  }
}
</style>
