<template>
  <div id="app">
    <el-container class="app-container">
      <!-- 顶部导航栏 -->
      <el-header class="app-header" v-if="isLoggedIn">
        <div class="header-content">
          <div class="header-left">
            <el-button
              @click="toggleSidebar"
              :icon="Expand"
              circle
              size="small"
              class="sidebar-toggle"
            />
            <h1 class="app-title">🤖 AI 智能服务平台</h1>
          </div>
          <div class="user-info">
            <el-tag type="info" class="token-display">
              <el-icon><Coin /></el-icon>
              {{ userStore.user?.tokens || 0 }} Tokens
            </el-tag>
            <el-dropdown @command="handleCommand">
              <span class="user-name">
                {{ userStore.user?.username }}
                <el-icon><ArrowDown /></el-icon>
              </span>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="profile">个人资料</el-dropdown-item>
                  <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>
      </el-header>

      <el-container v-if="isLoggedIn">
        <!-- 侧边栏 -->
        <el-aside :width="sidebarWidth" class="app-sidebar">
          <el-menu
            :default-active="$route.path"
            router
            class="sidebar-menu"
            :collapse="isCollapse"
          >
            <!-- 仪表板 -->
            <el-menu-item index="/dashboard">
              <el-icon><Odometer /></el-icon>
              <span>仪表板</span>
            </el-menu-item>

            <!-- 对话服务 -->
            <el-sub-menu index="chat-group">
              <template #title>
                <el-icon><ChatDotRound /></el-icon>
                <span>对话服务</span>
              </template>
              <el-menu-item index="/chat">
                <el-icon><ChatLineRound /></el-icon>
                <span>智能对话</span>
              </el-menu-item>
              <el-menu-item index="/models">
                <el-icon><Setting /></el-icon>
                <span>模型管理</span>
              </el-menu-item>
            </el-sub-menu>

            <!-- 工具服务 -->
            <el-sub-menu index="tools-group">
              <template #title>
                <el-icon><Tools /></el-icon>
                <span>工具服务</span>
              </template>
              <el-menu-item index="/converter">
                <el-icon><DocumentCopy /></el-icon>
                <span>格式转换</span>
              </el-menu-item>
              <el-menu-item index="/homework">
                <el-icon><EditPen /></el-icon>
                <span>作业批改</span>
              </el-menu-item>
              <el-menu-item index="/subtitle">
                <el-icon><VideoPlay /></el-icon>
                <span>字幕处理</span>
              </el-menu-item>
            </el-sub-menu>

            <!-- 数据服务 -->
            <el-sub-menu index="data-group">
              <template #title>
                <el-icon><DataAnalysis /></el-icon>
                <span>数据服务</span>
              </template>
              <el-menu-item index="/history">
                <el-icon><Clock /></el-icon>
                <span>历史记录</span>
              </el-menu-item>
              <el-menu-item index="/stats">
                <el-icon><TrendCharts /></el-icon>
                <span>使用统计</span>
              </el-menu-item>
            </el-sub-menu>

            <!-- 个人中心 -->
            <el-menu-item index="/profile">
              <el-icon><User /></el-icon>
              <span>个人中心</span>
            </el-menu-item>
          </el-menu>
        </el-aside>

        <!-- 主要内容区域 -->
        <el-main class="app-main">
          <router-view v-slot="{ Component }">
            <component :is="Component" v-if="Component" />
          </router-view>
        </el-main>
      </el-container>

      <!-- 未登录时显示登录页面 -->
      <el-main v-else class="login-main">
        <router-view v-slot="{ Component }">
          <component :is="Component" v-if="Component" />
        </router-view>
      </el-main>
    </el-container>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from './stores/user'
import { ElMessage } from 'element-plus'
import {
  Expand,
  Odometer,
  ChatDotRound,
  ChatLineRound,
  Setting,
  Tools,
  DocumentCopy,
  EditPen,
  VideoPlay,
  DataAnalysis,
  Clock,
  TrendCharts,
  User,
  Coin,
  ArrowDown
} from '@element-plus/icons-vue'

const router = useRouter()
const userStore = useUserStore()

const isLoggedIn = computed(() => userStore.isLoggedIn)
const isCollapse = ref(false)
const sidebarWidth = computed(() => isCollapse.value ? '64px' : '200px')

const toggleSidebar = () => {
  isCollapse.value = !isCollapse.value
}

const handleCommand = (command) => {
  switch (command) {
    case 'profile':
      router.push('/profile')
      break
    case 'logout':
      userStore.logout()
      ElMessage.success('已退出登录')
      router.push('/login')
      break
  }
}
</script>

<style scoped>
.app-container {
  height: 100vh;
}

.app-header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 0;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  z-index: 1000;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 100%;
  padding: 0 20px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 15px;
}

.sidebar-toggle {
  background: rgba(255, 255, 255, 0.2);
  border: none;
  color: white;
}

.sidebar-toggle:hover {
  background: rgba(255, 255, 255, 0.3);
}

.app-title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 15px;
}

.token-display {
  background: rgba(255, 255, 255, 0.2);
  color: white;
  border: none;
  font-weight: 500;
}

.user-name {
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 5px;
  font-weight: 500;
  transition: opacity 0.3s;
}

.user-name:hover {
  opacity: 0.8;
}

.app-sidebar {
  background: white;
  border-right: 1px solid #e4e7ed;
  transition: width 0.3s;
}

.sidebar-menu {
  border-right: none;
  height: 100%;
}

.app-main {
  padding: 20px;
  background: #f5f7fa;
  overflow-y: auto;
}

.login-main {
  padding: 0;
  background: #f5f7fa;
}
</style>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: 'Helvetica Neue', Helvetica, 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', '微软雅黑', Arial, sans-serif;
}

#app {
  height: 100vh;
}
</style>