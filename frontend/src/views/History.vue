<template>
  <div class="history-container">
    <div class="history-content">
      <el-card class="history-card">
        <template #header>
          <div class="card-header">
            <h2>聊天记录</h2>
            <el-button @click="refreshHistory" :loading="loading">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </div>
        </template>
        
        <div class="history-list">
          <div v-if="loading && history.length === 0" class="loading-state">
            <el-skeleton :rows="5" animated />
          </div>
          
          <div v-else-if="history.length === 0" class="empty-state">
            <el-icon size="64" color="#ccc"><ChatDotRound /></el-icon>
            <p>暂无聊天记录</p>
            <el-button type="primary" @click="$router.push('/chat')">
              开始聊天
            </el-button>
          </div>
          
          <div v-else>
            <div
              v-for="(item, index) in history"
              :key="item.id"
              class="history-item"
            >
              <div class="item-header">
                <div class="item-info">
                  <span class="item-endpoint">{{ item.endpoint }}</span>
                  <el-tag size="small" type="info">
                    <el-icon><Coin /></el-icon>
                    {{ item.tokens_consumed }}
                  </el-tag>
                </div>
                <span class="item-time">{{ formatDateTime(item.created_at) }}</span>
              </div>
              
              <div class="item-content">
                <div class="request-section">
                  <h4>请求内容</h4>
                  <div class="content-text">{{ getRequestMessage(item.request_data) }}</div>
                </div>
                
                <div class="response-section">
                  <h4>回复内容</h4>
                  <div class="content-text">{{ getResponseMessage(item.response_data) }}</div>
                </div>
              </div>
            </div>
            
            <!-- 分页 -->
            <div class="pagination-container" v-if="total > pageSize">
              <el-pagination
                v-model:current-page="currentPage"
                :page-size="pageSize"
                :total="total"
                layout="prev, pager, next, total"
                @current-change="handlePageChange"
              />
            </div>
          </div>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { ChatDotRound, Refresh, Coin } from '@element-plus/icons-vue'
import api from '../utils/api'

const loading = ref(false)
const history = reactive([])
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

// 格式化日期时间
const formatDateTime = (dateString) => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

// 获取请求消息
const getRequestMessage = (requestData) => {
  try {
    if (typeof requestData === 'string') {
      const data = JSON.parse(requestData)
      return data.message || requestData
    }
    return requestData?.message || JSON.stringify(requestData)
  } catch {
    return requestData || '-'
  }
}

// 获取回复消息
const getResponseMessage = (responseData) => {
  try {
    if (typeof responseData === 'string') {
      const data = JSON.parse(responseData)
      return data.response || responseData
    }
    return responseData?.response || JSON.stringify(responseData)
  } catch {
    return responseData || '-'
  }
}

// 获取聊天记录
const fetchHistory = async () => {
  loading.value = true
  try {
    const response = await api.get('/llm/history', {
      params: {
        page: currentPage.value,
        limit: pageSize.value
      }
    })
    
    const { calls, total: totalCount } = response.data
    history.splice(0, history.length, ...calls)
    total.value = totalCount
  } catch (error) {
    ElMessage.error('获取聊天记录失败')
  } finally {
    loading.value = false
  }
}

// 刷新记录
const refreshHistory = () => {
  currentPage.value = 1
  fetchHistory()
}

// 页面变化
const handlePageChange = (page) => {
  currentPage.value = page
  fetchHistory()
}

// 初始化
onMounted(() => {
  fetchHistory()
})
</script>

<style scoped>
.history-container {
  min-height: 100vh;
  background: #f5f7fa;
  padding: 20px;
}

.history-content {
  max-width: 1000px;
  margin: 0 auto;
}

.history-card {
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

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: #999;
}

.empty-state p {
  margin: 16px 0 24px 0;
  font-size: 16px;
}

.history-item {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  margin-bottom: 16px;
  overflow: hidden;
  transition: box-shadow 0.3s;
}

.history-item:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: #f8f9fa;
  border-bottom: 1px solid #e4e7ed;
}

.item-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.item-endpoint {
  font-weight: 500;
  color: #333;
}

.item-time {
  font-size: 12px;
  color: #999;
}

.item-content {
  padding: 16px;
}

.request-section,
.response-section {
  margin-bottom: 16px;
}

.request-section:last-child,
.response-section:last-child {
  margin-bottom: 0;
}

.request-section h4,
.response-section h4 {
  margin: 0 0 8px 0;
  font-size: 14px;
  color: #666;
  font-weight: 500;
}

.request-section h4 {
  color: #667eea;
}

.response-section h4 {
  color: #67c23a;
}

.content-text {
  background: #f8f9fa;
  padding: 12px;
  border-radius: 6px;
  font-size: 14px;
  line-height: 1.5;
  color: #333;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 200px;
  overflow-y: auto;
}

.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid #e4e7ed;
}

/* 滚动条样式 */
.content-text::-webkit-scrollbar {
  width: 6px;
}

.content-text::-webkit-scrollbar-track {
  background: transparent;
}

.content-text::-webkit-scrollbar-thumb {
  background: #ddd;
  border-radius: 3px;
}

.content-text::-webkit-scrollbar-thumb:hover {
  background: #ccc;
}
</style>