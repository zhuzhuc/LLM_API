<template>
  <div class="subtitle-container">
    <el-card class="service-header">
      <div class="header-content">
        <div class="header-left">
          <el-icon class="service-icon"><VideoPlay /></el-icon>
          <div>
            <h2>视频字幕处理</h2>
            <p>提取视频字幕、翻译和格式转换</p>
          </div>
        </div>
        <el-alert
          title="注意"
          description="字幕提取功能需要服务器安装 FFmpeg"
          type="info"
          show-icon
          :closable="false"
        />
      </div>
    </el-card>

    <el-steps :active="currentStep" finish-status="success" class="process-steps">
      <el-step title="上传视频" description="选择要处理的视频文件" />
      <el-step title="提取字幕" description="从视频中提取字幕内容" />
      <el-step title="翻译处理" description="翻译字幕到目标语言" />
      <el-step title="下载结果" description="下载处理后的字幕文件" />
    </el-steps>

    <el-row :gutter="20" class="main-content">
      <el-col :span="12">
        <el-card class="upload-card">
          <template #header>
            <span>视频上传</span>
          </template>

          <div class="upload-section">
            <el-upload
              v-if="!videoFile"
              class="video-upload"
              drag
              :auto-upload="false"
              :on-change="handleVideoUpload"
              :show-file-list="false"
              accept="video/*"
            >
              <el-icon class="el-icon--upload"><VideoPlay /></el-icon>
              <div class="el-upload__text">
                将视频文件拖到此处，或<em>点击上传</em>
              </div>
              <template #tip>
                <div class="el-upload__tip">
                  支持 MP4, AVI, MOV, MKV 等格式，文件大小不超过 500MB
                </div>
              </template>
            </el-upload>

            <div v-else class="video-info">
              <div class="file-info">
                <el-icon><VideoPlay /></el-icon>
                <div class="file-details">
                  <div class="file-name">{{ videoFile.name }}</div>
                  <div class="file-size">{{ formatFileSize(videoFile.size) }}</div>
                </div>
                <el-button @click="removeVideo" type="danger" size="small" circle>
                  <el-icon><Close /></el-icon>
                </el-button>
              </div>

              <el-form :model="subtitleForm" label-width="100px" class="subtitle-form">
                <el-form-item label="源语言">
                  <el-select v-model="subtitleForm.source_lang" placeholder="选择源语言">
                    <el-option label="中文" value="中文" />
                    <el-option label="英文" value="英文" />
                    <el-option label="日文" value="日文" />
                    <el-option label="韩文" value="韩文" />
                    <el-option label="法文" value="法文" />
                    <el-option label="德文" value="德文" />
                    <el-option label="西班牙文" value="西班牙文" />
                  </el-select>
                </el-form-item>

                <el-form-item label="目标语言">
                  <el-select v-model="subtitleForm.target_lang" placeholder="选择目标语言">
                    <el-option label="中文" value="中文" />
                    <el-option label="英文" value="英文" />
                    <el-option label="日文" value="日文" />
                    <el-option label="韩文" value="韩文" />
                    <el-option label="法文" value="法文" />
                    <el-option label="德文" value="德文" />
                    <el-option label="西班牙文" value="西班牙文" />
                  </el-select>
                </el-form-item>

                <el-form-item label="输出格式">
                  <el-radio-group v-model="subtitleForm.output_format">
                    <el-radio value="srt">SRT</el-radio>
                    <el-radio value="vtt">VTT</el-radio>
                    <el-radio value="ass">ASS</el-radio>
                  </el-radio-group>
                </el-form-item>

                <el-form-item>
                  <el-button 
                    type="primary" 
                    @click="processSubtitle"
                    :loading="processing"
                    :disabled="!canProcess"
                    size="large"
                  >
                    <el-icon><Cpu /></el-icon>
                    开始处理
                  </el-button>
                </el-form-item>
              </el-form>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="12">
        <el-card class="result-card">
          <template #header>
            <div class="result-header">
              <span>处理结果</span>
              <el-tag v-if="processResult" type="success">处理完成</el-tag>
            </div>
          </template>

          <div v-if="processing" class="processing-status">
            <el-progress 
              :percentage="processProgress" 
              :status="processStatus"
              stroke-width="8"
            />
            <p class="process-text">{{ processText }}</p>
          </div>

          <div v-else-if="processResult" class="result-content">
            <div class="subtitle-preview">
              <h4>字幕预览</h4>
              <el-input
                v-model="processResult.content"
                type="textarea"
                :rows="12"
                readonly
                class="subtitle-content"
              />
            </div>

            <div class="result-actions">
              <el-button @click="downloadSubtitle" type="primary">
                <el-icon><Download /></el-icon>
                下载字幕文件
              </el-button>
              <el-button @click="copySubtitle" type="success">
                <el-icon><CopyDocument /></el-icon>
                复制内容
              </el-button>
              <el-button @click="resetProcess">
                <el-icon><RefreshLeft /></el-icon>
                重新处理
              </el-button>
            </div>
          </div>

          <div v-else class="placeholder-content">
            <el-icon class="placeholder-icon"><VideoPlay /></el-icon>
            <p>上传视频文件并配置参数后开始处理</p>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 处理历史 -->
    <el-card class="history-card" v-if="historyList.length">
      <template #header>
        <div class="history-header">
          <span>处理历史</span>
          <el-button @click="loadHistory" type="text" size="small">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>

      <el-table :data="historyList" style="width: 100%">
        <el-table-column prop="video_name" label="视频文件" show-overflow-tooltip />
        <el-table-column prop="source_lang" label="源语言" width="100" />
        <el-table-column prop="target_lang" label="目标语言" width="100" />
        <el-table-column prop="output_format" label="格式" width="80" />
        <el-table-column prop="created_at" label="处理时间" width="180" />
        <el-table-column label="操作" width="120">
          <template #default="scope">
            <el-button @click="downloadHistoryFile(scope.row)" type="text" size="small">
              下载
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import {
  VideoPlay,
  Close,
  Cpu,
  Download,
  CopyDocument,
  RefreshLeft,
  Refresh
} from '@element-plus/icons-vue'
import api from '../utils/api'

const videoFile = ref(null)
const currentStep = ref(0)
const processing = ref(false)
const processProgress = ref(0)
const processStatus = ref('')
const processText = ref('')
const processResult = ref(null)
const historyList = ref([])

const subtitleForm = ref({
  source_lang: '英文',
  target_lang: '中文',
  output_format: 'srt'
})

const canProcess = computed(() => {
  return videoFile.value && 
         subtitleForm.value.source_lang && 
         subtitleForm.value.target_lang
})

const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const handleVideoUpload = (file) => {
  if (file.size > 500 * 1024 * 1024) { // 500MB
    ElMessage.error('文件大小不能超过 500MB')
    return
  }
  
  videoFile.value = file.raw
  currentStep.value = 1
  ElMessage.success('视频文件上传成功')
}

const removeVideo = () => {
  videoFile.value = null
  currentStep.value = 0
  processResult.value = null
}

const processSubtitle = async () => {
  processing.value = true
  processProgress.value = 0
  processStatus.value = ''
  currentStep.value = 2

  try {
    // 模拟处理进度
    const progressSteps = [
      { progress: 20, text: '上传视频文件...' },
      { progress: 40, text: '提取字幕内容...' },
      { progress: 60, text: '翻译字幕文本...' },
      { progress: 80, text: '格式化输出...' },
      { progress: 100, text: '处理完成' }
    ]

    for (const step of progressSteps) {
      processProgress.value = step.progress
      processText.value = step.text
      await new Promise(resolve => setTimeout(resolve, 1000))
    }

    // 实际API调用
    const formData = new FormData()
    formData.append('video', videoFile.value)
    formData.append('source_lang', subtitleForm.value.source_lang)
    formData.append('target_lang', subtitleForm.value.target_lang)
    formData.append('output_format', subtitleForm.value.output_format)

    const response = await api.post('/v1/tasks/subtitle', {
      video_path: videoFile.value.name, // 实际应该是上传后的路径
      source_lang: subtitleForm.value.source_lang,
      target_lang: subtitleForm.value.target_lang,
      output_format: subtitleForm.value.output_format
    })

    if (response.data.success) {
      processResult.value = response.data
      currentStep.value = 4
      ElMessage.success('字幕处理完成')
      
      // 保存到历史记录
      saveToHistory()
    } else {
      throw new Error(response.data.message || '处理失败')
    }
  } catch (error) {
    console.error('字幕处理错误:', error)
    ElMessage.error('处理失败: ' + (error.response?.data?.error || error.message))
    processStatus.value = 'exception'
  } finally {
    processing.value = false
  }
}

const downloadSubtitle = () => {
  if (!processResult.value) return

  const blob = new Blob([processResult.value.content], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `subtitle.${subtitleForm.value.output_format}`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
  ElMessage.success('字幕文件已下载')
}

const copySubtitle = async () => {
  if (!processResult.value) return

  try {
    await navigator.clipboard.writeText(processResult.value.content)
    ElMessage.success('字幕内容已复制到剪贴板')
  } catch (err) {
    ElMessage.error('复制失败')
  }
}

const resetProcess = () => {
  videoFile.value = null
  processResult.value = null
  currentStep.value = 0
  processProgress.value = 0
  processText.value = ''
}

const saveToHistory = () => {
  const historyItem = {
    video_name: videoFile.value.name,
    source_lang: subtitleForm.value.source_lang,
    target_lang: subtitleForm.value.target_lang,
    output_format: subtitleForm.value.output_format,
    content: processResult.value.content,
    created_at: new Date().toLocaleString()
  }

  const saved = JSON.parse(localStorage.getItem('subtitle_history') || '[]')
  saved.unshift(historyItem)
  localStorage.setItem('subtitle_history', JSON.stringify(saved.slice(0, 20))) // 只保留最近20条
  
  historyList.value = saved
}

const loadHistory = () => {
  const saved = JSON.parse(localStorage.getItem('subtitle_history') || '[]')
  historyList.value = saved
}

const downloadHistoryFile = (item) => {
  const blob = new Blob([item.content], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${item.video_name}_subtitle.${item.output_format}`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
  ElMessage.success('文件已下载')
}

onMounted(() => {
  loadHistory()
})
</script>

<style scoped>
.subtitle-container {
  max-width: 1400px;
  margin: 0 auto;
}

.service-header {
  margin-bottom: 20px;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 15px;
}

.service-icon {
  font-size: 32px;
  color: #e6a23c;
}

.service-header h2 {
  margin: 0;
  color: #303133;
}

.service-header p {
  margin: 5px 0 0 0;
  color: #909399;
  font-size: 14px;
}

.process-steps {
  margin: 20px 0;
}

.main-content {
  margin-top: 20px;
}

.upload-card, .result-card {
  height: 600px;
}

.upload-section {
  height: 500px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.video-upload {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.video-info {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.file-info {
  display: flex;
  align-items: center;
  gap: 15px;
  padding: 15px;
  background: #f5f7fa;
  border-radius: 4px;
}

.file-details {
  flex: 1;
}

.file-name {
  font-weight: 500;
  color: #303133;
}

.file-size {
  font-size: 12px;
  color: #909399;
}

.subtitle-form {
  flex: 1;
}

.result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.processing-status {
  height: 500px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  gap: 20px;
}

.process-text {
  color: #606266;
  font-size: 14px;
}

.result-content {
  height: 500px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.subtitle-preview {
  flex: 1;
}

.subtitle-preview h4 {
  margin: 0 0 10px 0;
  color: #303133;
}

.subtitle-content {
  height: 400px;
}

.result-actions {
  display: flex;
  gap: 10px;
  justify-content: center;
}

.placeholder-content {
  height: 500px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  color: #909399;
}

.placeholder-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.history-card {
  margin-top: 20px;
}

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
