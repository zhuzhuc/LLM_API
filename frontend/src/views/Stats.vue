<template>
  <div class="stats-container">
    <div class="stats-content">
      <el-card class="stats-card">
        <template #header>
          <div class="card-header">
            <h2>使用统计</h2>
            <el-button @click="refreshStats" :loading="loading">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </div>
        </template>
        
        <div class="stats-overview">
          <div v-if="loading" class="loading-state">
            <el-skeleton :rows="3" animated />
          </div>
          
          <div v-else class="stats-grid">
            <!-- 统计卡片 -->
            <div class="stat-card">
              <div class="stat-icon total-calls">
                <el-icon size="24"><ChatDotRound /></el-icon>
              </div>
              <div class="stat-content">
                <div class="stat-value">{{ stats.total_calls || 0 }}</div>
                <div class="stat-label">总对话次数</div>
              </div>
            </div>
            
            <div class="stat-card">
              <div class="stat-icon total-tokens">
                <el-icon size="24"><Coin /></el-icon>
              </div>
              <div class="stat-content">
                <div class="stat-value">{{ stats.total_tokens || 0 }}</div>
                <div class="stat-label">总消耗 Token</div>
              </div>
            </div>
            
            <div class="stat-card">
              <div class="stat-icon avg-tokens">
                <el-icon size="24"><TrendCharts /></el-icon>
              </div>
              <div class="stat-content">
                <div class="stat-value">{{ formatAverage(stats.average_tokens) }}</div>
                <div class="stat-label">平均每次消耗</div>
              </div>
            </div>
            
            <div class="stat-card">
              <div class="stat-icon current-tokens">
                <el-icon size="24"><Wallet /></el-icon>
              </div>
              <div class="stat-content">
                <div class="stat-value">{{ userStore.user?.tokens || 0 }}</div>
                <div class="stat-label">剩余 Token</div>
              </div>
            </div>
          </div>
        </div>
        
        <!-- 详细信息 -->
        <el-divider />
        
        <div class="stats-details">
          <h3>使用详情</h3>
          
          <el-row :gutter="20">
            <el-col :span="12">
              <el-card class="detail-card">
                <template #header>
                  <h4>Token 使用情况</h4>
                </template>
                
                <div class="token-progress">
                  <div class="progress-info">
                    <span>已使用</span>
                    <span>{{ stats.total_tokens || 0 }} / {{ (stats.total_tokens || 0) + (userStore.user?.tokens || 0) }}</span>
                  </div>
                  <el-progress
                    :percentage="getTokenUsagePercentage()"
                    :color="getProgressColor()"
                    :stroke-width="12"
                  />
                </div>
                
                <div class="token-breakdown">
                  <div class="breakdown-item">
                    <span class="label">已消耗：</span>
                    <span class="value consumed">{{ stats.total_tokens || 0 }}</span>
                  </div>
                  <div class="breakdown-item">
                    <span class="label">剩余：</span>
                    <span class="value remaining">{{ userStore.user?.tokens || 0 }}</span>
                  </div>
                  <div class="breakdown-item">
                    <span class="label">总计：</span>
                    <span class="value total">{{ (stats.total_tokens || 0) + (userStore.user?.tokens || 0) }}</span>
                  </div>
                </div>
              </el-card>
            </el-col>
            
            <el-col :span="12">
              <el-card class="detail-card">
                <template #header>
                  <h4>使用效率</h4>
                </template>
                
                <div class="efficiency-metrics">
                  <div class="metric-item">
                    <div class="metric-label">平均每次对话</div>
                    <div class="metric-value">{{ formatAverage(stats.average_tokens) }} Token</div>
                  </div>
                  
                  <div class="metric-item">
                    <div class="metric-label">使用频率</div>
                    <div class="metric-value">{{ getUsageFrequency() }}</div>
                  </div>
                  
                  <div class="metric-item">
                    <div class="metric-label">效率评级</div>
                    <div class="metric-value">
                      <el-tag :type="getEfficiencyType()">{{ getEfficiencyLabel() }}</el-tag>
                    </div>
                  </div>
                </div>
              </el-card>
            </el-col>
          </el-row>
        </div>
        
        <!-- 建议 -->
        <el-divider />
        
        <div class="stats-suggestions">
          <h3>使用建议</h3>
          <div class="suggestions-list">
            <el-alert
              v-for="suggestion in getSuggestions()"
              :key="suggestion.type"
              :title="suggestion.title"
              :description="suggestion.description"
              :type="suggestion.type"
              show-icon
              :closable="false"
              class="suggestion-item"
            />
          </div>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useUserStore } from '../stores/user'
import { ElMessage } from 'element-plus'
import { ChatDotRound, Coin, TrendCharts, Wallet, Refresh } from '@element-plus/icons-vue'
import api from '../utils/api'

const userStore = useUserStore()
const loading = ref(false)
const stats = reactive({
  total_calls: 0,
  total_tokens: 0,
  average_tokens: 0
})

// 格式化平均值
const formatAverage = (value) => {
  if (!value) return '0'
  return Math.round(value * 100) / 100
}

// 获取Token使用百分比
const getTokenUsagePercentage = () => {
  const total = (stats.total_tokens || 0) + (userStore.user?.tokens || 0)
  if (total === 0) return 0
  return Math.round(((stats.total_tokens || 0) / total) * 100)
}

// 获取进度条颜色
const getProgressColor = () => {
  const percentage = getTokenUsagePercentage()
  if (percentage < 50) return '#67c23a'
  if (percentage < 80) return '#e6a23c'
  return '#f56c6c'
}

// 获取使用频率
const getUsageFrequency = () => {
  const calls = stats.total_calls || 0
  if (calls === 0) return '暂无数据'
  if (calls < 10) return '轻度使用'
  if (calls < 50) return '中度使用'
  return '重度使用'
}

// 获取效率类型
const getEfficiencyType = () => {
  const avg = stats.average_tokens || 0
  if (avg < 50) return 'success'
  if (avg < 100) return 'warning'
  return 'danger'
}

// 获取效率标签
const getEfficiencyLabel = () => {
  const avg = stats.average_tokens || 0
  if (avg < 50) return '高效'
  if (avg < 100) return '一般'
  return '待优化'
}

// 获取建议
const getSuggestions = () => {
  const suggestions = []
  const remaining = userStore.user?.tokens || 0
  const avg = stats.average_tokens || 0
  const total = stats.total_calls || 0
  
  if (remaining < 100) {
    suggestions.push({
      type: 'warning',
      title: 'Token 余额不足',
      description: '您的 Token 余额较低，建议及时充值以免影响使用。'
    })
  }
  
  if (avg > 100) {
    suggestions.push({
      type: 'info',
      title: '优化对话效率',
      description: '您的平均 Token 消耗较高，建议精简问题描述以提高使用效率。'
    })
  }
  
  if (total > 0 && avg < 30) {
    suggestions.push({
      type: 'success',
      title: '使用效率良好',
      description: '您的 Token 使用效率很高，继续保持这种良好的使用习惯。'
    })
  }
  
  if (total === 0) {
    suggestions.push({
      type: 'info',
      title: '开始使用',
      description: '欢迎使用 LLM Chat！您可以开始与 AI 对话，体验智能问答服务。'
    })
  }
  
  return suggestions
}

// 获取统计数据
const fetchStats = async () => {
  loading.value = true
  try {
    const response = await api.get('/llm/stats')
    Object.assign(stats, response.data)
  } catch (error) {
    ElMessage.error('获取统计数据失败')
  } finally {
    loading.value = false
  }
}

// 刷新统计
const refreshStats = async () => {
  await Promise.all([
    fetchStats(),
    userStore.fetchProfile()
  ])
}

// 初始化
onMounted(() => {
  refreshStats()
})
</script>

<style scoped>
.stats-container {
  min-height: 100vh;
  background: #f5f7fa;
  padding: 20px;
}

.stats-content {
  max-width: 1000px;
  margin: 0 auto;
}

.stats-card {
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h2 {
  margin: 0;
  color: #333;
}

.loading-state {
  padding: 20px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
  margin-bottom: 20px;
}

.stat-card {
  display: flex;
  align-items: center;
  padding: 20px;
  background: white;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  transition: transform 0.3s, box-shadow 0.3s;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16px;
  color: white;
}

.stat-icon.total-calls {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.stat-icon.total-tokens {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.stat-icon.avg-tokens {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
}

.stat-icon.current-tokens {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: #333;
  line-height: 1;
}

.stat-label {
  font-size: 14px;
  color: #666;
  margin-top: 4px;
}

.stats-details h3,
.stats-suggestions h3 {
  margin: 0 0 20px 0;
  color: #333;
  font-size: 18px;
}

.detail-card {
  height: 100%;
}

.detail-card h4 {
  margin: 0;
  color: #333;
  font-size: 16px;
}

.token-progress {
  margin-bottom: 20px;
}

.progress-info {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  font-size: 14px;
  color: #666;
}

.token-breakdown {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.breakdown-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 14px;
}

.breakdown-item .label {
  color: #666;
}

.breakdown-item .value {
  font-weight: 500;
}

.breakdown-item .value.consumed {
  color: #f56c6c;
}

.breakdown-item .value.remaining {
  color: #67c23a;
}

.breakdown-item .value.total {
  color: #333;
}

.efficiency-metrics {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.metric-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
}

.metric-item:last-child {
  border-bottom: none;
}

.metric-label {
  font-size: 14px;
  color: #666;
}

.metric-value {
  font-weight: 500;
  color: #333;
}

.suggestions-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.suggestion-item {
  border-radius: 8px;
}

:deep(.el-alert__description) {
  margin-top: 4px;
  line-height: 1.5;
}
</style>