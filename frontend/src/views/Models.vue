<template>
  <div class="models-container">
    <div class="models-header">
      <h2>模型管理</h2>
      <p>管理和监控您的 AI 模型</p>
    </div>

    <!-- 模型统计卡片 -->
    <div class="stats-cards">
      <el-card class="stat-card">
        <div class="stat-content">
          <div class="stat-icon">
            <el-icon size="24" color="#67C23A"><CircleCheck /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-number">{{ runningModels.length }}</div>
            <div class="stat-label">运行中模型</div>
          </div>
        </div>
      </el-card>

      <el-card class="stat-card">
        <div class="stat-content">
          <div class="stat-icon">
            <el-icon size="24" color="#409EFF"><Collection /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-number">{{ availableModels.length }}</div>
            <div class="stat-label">可用模型</div>
          </div>
        </div>
      </el-card>

      <el-card class="stat-card">
        <div class="stat-content">
          <div class="stat-icon">
            <el-icon size="24" color="#E6A23C"><Monitor /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-number">{{ totalRequests }}</div>
            <div class="stat-label">总请求数</div>
          </div>
        </div>
      </el-card>
    </div>

    <!-- 可用模型列表 -->
    <el-card class="models-card">
      <template #header>
        <div class="card-header">
          <span>可用模型</span>
          <el-button type="primary" @click="refreshModels" :loading="loading">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>

      <el-table :data="availableModels" v-loading="loading">
        <el-table-column prop="modelName" label="模型名称" min-width="200">
          <template #default="{ row }">
            <div class="model-name">
              <strong>{{ row.modelName }}</strong>
              <div class="model-description">{{ row.description }}</div>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="modelFile" label="模型文件" min-width="250" />

        <el-table-column label="配置" min-width="200">
          <template #default="{ row }">
            <div class="model-config">
              <el-tag size="small">{{ row.contextLength }} 上下文</el-tag>
              <el-tag size="small" type="info">{{ row.threads }} 线程</el-tag>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <el-tag 
              :type="getModelStatus(row.modelName) === 'running' ? 'success' : 'info'"
              size="small"
            >
              {{ getModelStatus(row.modelName) === 'running' ? '运行中' : '未启动' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="200">
          <template #default="{ row }">
            <el-button-group>
              <el-button
                v-if="getModelStatus(row.modelName) !== 'running'"
                type="primary"
                size="small"
                @click="startModel(row.modelName)"
                :loading="modelLoading[row.modelName]"
              >
                启动
              </el-button>
              <el-button
                v-else
                type="danger"
                size="small"
                @click="stopModel(row.modelName)"
                :loading="modelLoading[row.modelName]"
              >
                停止
              </el-button>
              <el-button
                type="info"
                size="small"
                @click="testModel(row.modelName)"
                :disabled="getModelStatus(row.modelName) !== 'running'"
              >
                测试
              </el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 运行中模型详情 -->
    <el-card class="running-models-card" v-if="runningModels.length > 0">
      <template #header>
        <span>运行中模型详情</span>
      </template>

      <el-table :data="runningModels">
        <el-table-column prop="name" label="模型名称" />
        <el-table-column prop="port" label="端口" width="100" />
        <el-table-column label="运行时间" width="150">
          <template #default="{ row }">
            {{ formatUptime(row.start_time) }}
          </template>
        </el-table-column>
        <el-table-column prop="usage_count" label="使用次数" width="120" />
        <el-table-column label="最后使用" width="150">
          <template #default="{ row }">
            {{ formatTime(row.last_used) }}
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 测试对话框 -->
    <el-dialog v-model="testDialogVisible" title="模型测试" width="600px">
      <div class="test-dialog">
        <el-input
          v-model="testMessage"
          type="textarea"
          :rows="3"
          placeholder="输入测试消息..."
          maxlength="500"
          show-word-limit
        />
        <div class="test-result" v-if="testResult">
          <h4>响应结果：</h4>
          <div class="result-content">{{ testResult.response }}</div>
          <div class="result-meta">
            <span>使用 Token: {{ testResult.tokens_used }}</span>
            <span>模型: {{ testResult.model }}</span>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="testDialogVisible = false">取消</el-button>
        <el-button 
          type="primary" 
          @click="sendTestMessage"
          :loading="testLoading"
          :disabled="!testMessage.trim()"
        >
          发送测试
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../utils/api'

// 响应式数据
const loading = ref(false)
const availableModels = ref([])
const runningModels = ref([])
const modelLoading = reactive({})
const testDialogVisible = ref(false)
const testMessage = ref('')
const testResult = ref(null)
const testLoading = ref(false)
const currentTestModel = ref('')

// 计算属性
const totalRequests = computed(() => {
  return runningModels.value.reduce((total, model) => total + (model.usage_count || 0), 0)
})

// 获取可用模型
const fetchAvailableModels = async () => {
  try {
    const response = await api.get('/v1/models')
    availableModels.value = response.data.data || []
  } catch (error) {
    ElMessage.error('获取模型列表失败')
  }
}

// 获取运行中模型
const fetchRunningModels = async () => {
  try {
    const response = await api.get('/v1/models/running')
    runningModels.value = response.data.data || []
  } catch (error) {
    ElMessage.error('获取运行中模型失败')
  }
}

// 刷新模型数据
const refreshModels = async () => {
  loading.value = true
  try {
    await Promise.all([fetchAvailableModels(), fetchRunningModels()])
  } finally {
    loading.value = false
  }
}

// 获取模型状态
const getModelStatus = (modelName) => {
  return runningModels.value.some(model => model.name === modelName) ? 'running' : 'stopped'
}

// 启动模型
const startModel = async (modelName) => {
  modelLoading[modelName] = true
  try {
    await api.post(`/v1/models/${modelName}/start`)
    ElMessage.success(`模型 ${modelName} 启动成功`)
    await fetchRunningModels()
  } catch (error) {
    ElMessage.error(`启动模型失败: ${error.response?.data?.error || error.message}`)
  } finally {
    modelLoading[modelName] = false
  }
}

// 停止模型
const stopModel = async (modelName) => {
  modelLoading[modelName] = true
  try {
    await api.post(`/v1/models/${modelName}/stop`)
    ElMessage.success(`模型 ${modelName} 已停止`)
    await fetchRunningModels()
  } catch (error) {
    ElMessage.error(`停止模型失败: ${error.response?.data?.error || error.message}`)
  } finally {
    modelLoading[modelName] = false
  }
}

// 测试模型
const testModel = (modelName) => {
  currentTestModel.value = modelName
  testMessage.value = ''
  testResult.value = null
  testDialogVisible.value = true
}

// 发送测试消息
const sendTestMessage = async () => {
  if (!testMessage.value.trim()) return

  testLoading.value = true
  try {
    const response = await api.post(`/v1/models/${currentTestModel.value}/chat`, {
      message: testMessage.value,
      max_tokens: 200
    })
    testResult.value = response.data.data
    ElMessage.success('测试成功')
  } catch (error) {
    ElMessage.error(`测试失败: ${error.response?.data?.error || error.message}`)
  } finally {
    testLoading.value = false
  }
}

// 格式化时间
const formatTime = (timestamp) => {
  if (!timestamp) return '-'
  return new Date(timestamp).toLocaleString()
}

// 格式化运行时间
const formatUptime = (startTime) => {
  if (!startTime) return '-'
  const now = new Date()
  const start = new Date(startTime)
  const diff = now - start
  const hours = Math.floor(diff / (1000 * 60 * 60))
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))
  return `${hours}h ${minutes}m`
}

// 组件挂载时获取数据
onMounted(() => {
  refreshModels()
})
</script>

<style scoped>
.models-container {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

.models-header {
  margin-bottom: 24px;
}

.models-header h2 {
  margin: 0 0 8px 0;
  color: #333;
  font-size: 24px;
}

.models-header p {
  margin: 0;
  color: #666;
  font-size: 14px;
}

.stats-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  margin-bottom: 24px;
}

.stat-card {
  border: none;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.stat-icon {
  flex-shrink: 0;
}

.stat-info {
  flex: 1;
}

.stat-number {
  font-size: 24px;
  font-weight: 600;
  color: #333;
  line-height: 1;
}

.stat-label {
  font-size: 12px;
  color: #666;
  margin-top: 4px;
}

.models-card, .running-models-card {
  margin-bottom: 24px;
  border: none;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.model-name strong {
  display: block;
  color: #333;
  margin-bottom: 4px;
}

.model-description {
  font-size: 12px;
  color: #666;
}

.model-config {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.test-dialog {
  padding: 16px 0;
}

.test-result {
  margin-top: 16px;
  padding: 16px;
  background: #f5f7fa;
  border-radius: 8px;
}

.test-result h4 {
  margin: 0 0 12px 0;
  color: #333;
  font-size: 14px;
}

.result-content {
  background: white;
  padding: 12px;
  border-radius: 6px;
  margin-bottom: 12px;
  line-height: 1.6;
  color: #333;
}

.result-meta {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: #666;
}
</style>
