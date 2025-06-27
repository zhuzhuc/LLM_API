<template>
  <div class="chat-container">
    <div class="chat-content">
      <!-- 消息列表 -->
      <div class="message-list" ref="messageListRef">
        <div v-if="messages.length === 0" class="empty-state">
          <el-icon size="64" color="#ccc"><ChatDotRound /></el-icon>
          <p>开始您的第一次对话吧！</p>
        </div>
        
        <div
          v-for="(message, index) in messages"
          :key="index"
          class="message-item"
          :class="{ 'user-message': message.role === 'user', 'assistant-message': message.role === 'assistant' }"
        >
          <div class="message-avatar">
            <el-avatar v-if="message.role === 'user'" :icon="UserFilled" />
            <el-avatar v-else background-color="#667eea">
              <el-icon><Robot /></el-icon>
            </el-avatar>
          </div>
          
          <div class="message-content">
            <div class="message-bubble">
              <div class="message-text">{{ message.content }}</div>
              <div class="message-meta">
                <span class="message-time">{{ formatTime(message.timestamp) }}</span>
                <span v-if="message.tokens" class="message-tokens">
                  <el-icon><Coin /></el-icon>
                  {{ message.tokens }}
                </span>
              </div>
            </div>
          </div>
        </div>
        
        <!-- 加载状态 -->
        <div v-if="loading" class="message-item assistant-message">
          <div class="message-avatar">
            <el-avatar background-color="#667eea">
              <el-icon><Robot /></el-icon>
            </el-avatar>
          </div>
          <div class="message-content">
            <div class="message-bubble loading">
              <div class="typing-indicator">
                <span></span>
                <span></span>
                <span></span>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <!-- 输入区域 -->
      <div class="input-area">
        <!-- 模型选择 -->
        <div class="model-selector">
          <el-select
            v-model="selectedModel"
            placeholder="选择模型"
            size="small"
            style="width: 200px"
            @change="onModelChange"
          >
            <el-option
              v-for="model in runningModels"
              :key="model.name"
              :label="model.name"
              :value="model.name"
            >
              <div class="model-option">
                <span>{{ model.name }}</span>
                <el-tag size="small" type="success">运行中</el-tag>
              </div>
            </el-option>
          </el-select>
          <el-button
            size="small"
            @click="refreshModels"
            :loading="modelsLoading"
          >
            <el-icon><Refresh /></el-icon>
          </el-button>
        </div>

        <div class="input-container">
          <el-input
            v-model="inputMessage"
            type="textarea"
            :rows="1"
            :autosize="{ minRows: 1, maxRows: 4 }"
            placeholder="输入您的问题..."
            class="message-input"
            @keydown.enter.exact.prevent="sendMessage"
            @keydown.enter.shift.exact="handleShiftEnter"
            :disabled="loading || !selectedModel"
          />
          <el-button
            type="primary"
            :icon="Position"
            class="send-button"
            :loading="loading"
            :disabled="!inputMessage.trim() || !selectedModel"
            @click="sendMessage"
          >
            发送
          </el-button>
        </div>
        <div class="input-tips">
          <span>按 Enter 发送，Shift + Enter 换行</span>
          <span class="token-info">
            剩余 {{ userStore.user?.tokens || 0 }} tokens
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, nextTick, onMounted } from 'vue'
import { useUserStore } from '../stores/user'
import { ElMessage } from 'element-plus'
import { UserFilled, Position, Robot, ChatDotRound, Coin, Refresh } from '@element-plus/icons-vue'
import api from '../utils/api'

const userStore = useUserStore()
const messageListRef = ref()
const inputMessage = ref('')
const loading = ref(false)
const messages = reactive([])
const selectedModel = ref('')
const runningModels = ref([])
const modelsLoading = ref(false)

// 格式化时间
const formatTime = (timestamp) => {
  const date = new Date(timestamp)
  return date.toLocaleTimeString('zh-CN', { 
    hour: '2-digit', 
    minute: '2-digit' 
  })
}

// 获取运行中的模型
const fetchRunningModels = async () => {
  modelsLoading.value = true
  try {
    const response = await api.get('/v1/models/running')
    runningModels.value = response.data.data || []

    // 如果没有选择模型且有可用模型，自动选择第一个
    if (!selectedModel.value && runningModels.value.length > 0) {
      selectedModel.value = runningModels.value[0].name
    }
  } catch (error) {
    ElMessage.error('获取运行中模型失败')
  } finally {
    modelsLoading.value = false
  }
}

// 刷新模型列表
const refreshModels = () => {
  fetchRunningModels()
}

// 模型选择变化
const onModelChange = (modelName) => {
  ElMessage.success(`已切换到模型: ${modelName}`)
}

// 滚动到底部
const scrollToBottom = () => {
  nextTick(() => {
    if (messageListRef.value) {
      messageListRef.value.scrollTop = messageListRef.value.scrollHeight
    }
  })
}

// 处理 Shift + Enter
const handleShiftEnter = (event) => {
  // 允许默认行为（换行）
}

// 发送消息
const sendMessage = async () => {
  if (!inputMessage.value.trim() || loading.value || !selectedModel.value) return

  const userMessage = {
    role: 'user',
    content: inputMessage.value.trim(),
    timestamp: Date.now()
  }

  messages.push(userMessage)
  const currentMessage = inputMessage.value.trim()
  inputMessage.value = ''
  scrollToBottom()

  loading.value = true

  try {
    const response = await api.post(`/v1/models/${selectedModel.value}/chat`, {
      message: currentMessage,
      max_tokens: 200
    })

    const { response: aiResponse, tokens_used, model } = response.data.data

    // 添加AI回复
    messages.push({
      role: 'assistant',
      content: aiResponse,
      timestamp: Date.now(),
      tokens: tokens_used,
      model: model
    })

    // 更新用户token数量
    userStore.updateTokens(userStore.user.tokens - tokens_used)

    scrollToBottom()
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '发送失败，请重试')

    // 如果是token不足的错误，从消息列表中移除用户消息
    if (error.response?.status === 400 && error.response?.data?.error?.includes('token')) {
      messages.pop()
    }
  } finally {
    loading.value = false
  }
}

// 初始化
onMounted(() => {
  fetchRunningModels()
  // 可以在这里加载历史消息
})
</script>

<style scoped>
.chat-container {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f5f7fa;
}

.chat-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  max-width: 800px;
  margin: 0 auto;
  width: 100%;
  padding: 20px;
}

.message-list {
  flex: 1;
  overflow-y: auto;
  padding: 20px 0;
  scroll-behavior: smooth;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #999;
}

.empty-state p {
  margin-top: 16px;
  font-size: 16px;
}

.message-item {
  display: flex;
  margin-bottom: 20px;
  animation: fadeIn 0.3s ease-in;
}

.user-message {
  flex-direction: row-reverse;
}

.message-avatar {
  margin: 0 12px;
  flex-shrink: 0;
}

.message-content {
  flex: 1;
  max-width: 70%;
}

.user-message .message-content {
  display: flex;
  justify-content: flex-end;
}

.message-bubble {
  padding: 12px 16px;
  border-radius: 12px;
  word-wrap: break-word;
  line-height: 1.5;
}

.user-message .message-bubble {
  background: #667eea;
  color: white;
  border-bottom-right-radius: 4px;
}

.assistant-message .message-bubble {
  background: white;
  border: 1px solid #e4e7ed;
  border-bottom-left-radius: 4px;
}

.message-bubble.loading {
  padding: 16px;
}

.message-text {
  margin-bottom: 8px;
  white-space: pre-wrap;
}

.message-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
  opacity: 0.7;
}

.user-message .message-meta {
  color: rgba(255, 255, 255, 0.8);
}

.assistant-message .message-meta {
  color: #999;
}

.message-tokens {
  display: flex;
  align-items: center;
  gap: 2px;
}

.typing-indicator {
  display: flex;
  gap: 4px;
}

.typing-indicator span {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #667eea;
  animation: typing 1.4s infinite ease-in-out;
}

.typing-indicator span:nth-child(1) {
  animation-delay: -0.32s;
}

.typing-indicator span:nth-child(2) {
  animation-delay: -0.16s;
}

.input-area {
  background: white;
  border-radius: 12px;
  padding: 16px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.model-selector {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f0f0f0;
}

.model-option {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.input-container {
  display: flex;
  gap: 12px;
  align-items: flex-end;
}

.message-input {
  flex: 1;
}

.send-button {
  height: 40px;
  padding: 0 20px;
}

.input-tips {
  display: flex;
  justify-content: space-between;
  margin-top: 8px;
  font-size: 12px;
  color: #999;
}

.token-info {
  color: #667eea;
  font-weight: 500;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes typing {
  0%, 80%, 100% {
    transform: scale(0);
  }
  40% {
    transform: scale(1);
  }
}

/* 滚动条样式 */
.message-list::-webkit-scrollbar {
  width: 6px;
}

.message-list::-webkit-scrollbar-track {
  background: transparent;
}

.message-list::-webkit-scrollbar-thumb {
  background: #ddd;
  border-radius: 3px;
}

.message-list::-webkit-scrollbar-thumb:hover {
  background: #ccc;
}
</style>