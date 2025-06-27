<template>
  <div class="dashboard-container">
    <!-- 欢迎横幅 -->
    <el-card class="welcome-banner">
      <div class="banner-content">
        <div class="banner-left">
          <h1>🤖 欢迎使用 AI 智能服务平台</h1>
          <p>集成多种AI服务，为您提供智能化的工作和学习体验</p>
        </div>
        <div class="banner-stats">
          <el-statistic title="今日使用次数" :value="todayUsage" />
          <el-statistic title="可用Token" :value="userStore.user?.tokens || 0" />
        </div>
      </div>
    </el-card>

    <!-- 服务分类 -->
    <div class="services-section">
      <h2>🛠️ 服务中心</h2>
      
      <!-- 对话服务 -->
      <el-card class="service-category">
        <template #header>
          <div class="category-header">
            <el-icon class="category-icon"><ChatDotRound /></el-icon>
            <span>对话服务</span>
          </div>
        </template>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <div class="service-card" @click="navigateTo('/chat')">
              <div class="service-icon">
                <el-icon><ChatLineRound /></el-icon>
              </div>
              <div class="service-info">
                <h3>智能对话</h3>
                <p>与AI模型进行自然语言对话，获得智能回答</p>
                <el-tag type="success" size="small">{{ runningModels.length }} 个模型在线</el-tag>
              </div>
            </div>
          </el-col>
          <el-col :span="12">
            <div class="service-card" @click="navigateTo('/models')">
              <div class="service-icon">
                <el-icon><Setting /></el-icon>
              </div>
              <div class="service-info">
                <h3>模型管理</h3>
                <p>管理和配置AI模型，查看模型状态</p>
                <el-tag type="info" size="small">{{ totalModels }} 个模型可用</el-tag>
              </div>
            </div>
          </el-col>
        </el-row>
      </el-card>

      <!-- 工具服务 -->
      <el-card class="service-category">
        <template #header>
          <div class="category-header">
            <el-icon class="category-icon"><Tools /></el-icon>
            <span>工具服务</span>
          </div>
        </template>
        
        <el-row :gutter="20">
          <el-col :span="8">
            <div class="service-card" @click="navigateTo('/converter')">
              <div class="service-icon">
                <el-icon><DocumentCopy /></el-icon>
              </div>
              <div class="service-info">
                <h3>格式转换</h3>
                <p>智能转换各种文件格式</p>
                <el-tag type="primary" size="small">支持20+格式</el-tag>
              </div>
            </div>
          </el-col>
          <el-col :span="8">
            <div class="service-card" @click="navigateTo('/homework')">
              <div class="service-icon">
                <el-icon><EditPen /></el-icon>
              </div>
              <div class="service-info">
                <h3>作业批改</h3>
                <p>AI智能批改作业并提供反馈</p>
                <el-tag type="success" size="small">支持多学科</el-tag>
              </div>
            </div>
          </el-col>
          <el-col :span="8">
            <div class="service-card" @click="navigateTo('/subtitle')">
              <div class="service-icon">
                <el-icon><VideoPlay /></el-icon>
              </div>
              <div class="service-info">
                <h3>字幕处理</h3>
                <p>提取和翻译视频字幕</p>
                <el-tag type="warning" size="small">需要FFmpeg</el-tag>
              </div>
            </div>
          </el-col>
        </el-row>
      </el-card>

      <!-- 数据服务 -->
      <el-card class="service-category">
        <template #header>
          <div class="category-header">
            <el-icon class="category-icon"><DataAnalysis /></el-icon>
            <span>数据服务</span>
          </div>
        </template>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <div class="service-card" @click="navigateTo('/history')">
              <div class="service-icon">
                <el-icon><Clock /></el-icon>
              </div>
              <div class="service-info">
                <h3>历史记录</h3>
                <p>查看和管理使用历史记录</p>
                <el-tag type="info" size="small">{{ historyCount }} 条记录</el-tag>
              </div>
            </div>
          </el-col>
          <el-col :span="12">
            <div class="service-card" @click="navigateTo('/stats')">
              <div class="service-icon">
                <el-icon><TrendCharts /></el-icon>
              </div>
              <div class="service-info">
                <h3>使用统计</h3>
                <p>查看详细的使用统计和分析</p>
                <el-tag type="primary" size="small">实时更新</el-tag>
              </div>
            </div>
          </el-col>
        </el-row>
      </el-card>
    </div>

    <!-- 快速操作 -->
    <el-card class="quick-actions">
      <template #header>
        <span>⚡ 快速操作</span>
      </template>
      
      <div class="quick-buttons">
        <el-button type="primary" size="large" @click="navigateTo('/chat')">
          <el-icon><ChatLineRound /></el-icon>
          开始对话
        </el-button>
        <el-button type="success" size="large" @click="navigateTo('/converter')">
          <el-icon><DocumentCopy /></el-icon>
          格式转换
        </el-button>
        <el-button type="warning" size="large" @click="navigateTo('/homework')">
          <el-icon><EditPen /></el-icon>
          批改作业
        </el-button>
        <el-button type="info" size="large" @click="navigateTo('/models')">
          <el-icon><Setting /></el-icon>
          管理模型
        </el-button>
      </div>
    </el-card>

    <!-- 系统状态 -->
    <el-card class="system-status">
      <template #header>
        <span>📊 系统状态</span>
      </template>
      
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="status-item">
            <div class="status-value">{{ systemStatus.cpu }}%</div>
            <div class="status-label">CPU使用率</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="status-item">
            <div class="status-value">{{ systemStatus.memory }}%</div>
            <div class="status-label">内存使用率</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="status-item">
            <div class="status-value">{{ systemStatus.models }}</div>
            <div class="status-label">运行中模型</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="status-item">
            <div class="status-value">{{ systemStatus.uptime }}</div>
            <div class="status-label">运行时间</div>
          </div>
        </el-col>
      </el-row>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import {
  ChatDotRound,
  ChatLineRound,
  Setting,
  Tools,
  DocumentCopy,
  EditPen,
  VideoPlay,
  DataAnalysis,
  Clock,
  TrendCharts
} from '@element-plus/icons-vue'

const router = useRouter()
const userStore = useUserStore()

const todayUsage = ref(0)
const runningModels = ref([])
const totalModels = ref(6)
const historyCount = ref(0)

const systemStatus = ref({
  cpu: 45,
  memory: 62,
  models: 4,
  uptime: '2天3小时'
})

const navigateTo = (path) => {
  router.push(path)
}

const loadDashboardData = () => {
  // 加载今日使用次数
  const today = new Date().toDateString()
  const chatHistory = JSON.parse(localStorage.getItem('chat_history') || '[]')
  todayUsage.value = chatHistory.filter(item => 
    new Date(item.timestamp).toDateString() === today
  ).length

  // 加载历史记录数量
  const homeworkHistory = JSON.parse(localStorage.getItem('homework_history') || '[]')
  const subtitleHistory = JSON.parse(localStorage.getItem('subtitle_history') || '[]')
  historyCount.value = chatHistory.length + homeworkHistory.length + subtitleHistory.length

  // 模拟运行中的模型
  runningModels.value = [
    'deepseek-coder-1.3b',
    'qwen2-7b-instruct',
    'mistral-7b-instruct'
  ]
}

onMounted(() => {
  loadDashboardData()
})
</script>

<style scoped>
.dashboard-container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 20px;
}

.welcome-banner {
  margin-bottom: 30px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.banner-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.banner-left h1 {
  margin: 0 0 10px 0;
  font-size: 28px;
}

.banner-left p {
  margin: 0;
  font-size: 16px;
  opacity: 0.9;
}

.banner-stats {
  display: flex;
  gap: 40px;
}

.services-section h2 {
  margin: 0 0 20px 0;
  color: #303133;
}

.service-category {
  margin-bottom: 20px;
}

.category-header {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 16px;
  font-weight: 600;
}

.category-icon {
  font-size: 20px;
  color: #409eff;
}

.service-card {
  display: flex;
  align-items: center;
  gap: 15px;
  padding: 20px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s;
  height: 100px;
}

.service-card:hover {
  border-color: #409eff;
  box-shadow: 0 2px 12px rgba(64, 158, 255, 0.1);
  transform: translateY(-2px);
}

.service-card .service-icon {
  font-size: 32px;
  color: #409eff;
  min-width: 40px;
}

.service-info h3 {
  margin: 0 0 5px 0;
  color: #303133;
  font-size: 16px;
}

.service-info p {
  margin: 0 0 8px 0;
  color: #606266;
  font-size: 14px;
}

.quick-actions {
  margin: 20px 0;
}

.quick-buttons {
  display: flex;
  gap: 15px;
  justify-content: center;
  flex-wrap: wrap;
}

.system-status {
  margin-top: 20px;
}

.status-item {
  text-align: center;
  padding: 20px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
}

.status-value {
  font-size: 24px;
  font-weight: 600;
  color: #409eff;
  margin-bottom: 5px;
}

.status-label {
  font-size: 14px;
  color: #909399;
}
</style>
