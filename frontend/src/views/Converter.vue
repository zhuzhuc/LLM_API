<template>
  <div class="converter-container">
    <el-card class="service-header">
      <div class="header-content">
        <div class="header-left">
          <el-icon class="service-icon"><DocumentCopy /></el-icon>
          <div>
            <h2>文件格式转换</h2>
            <p>支持多种文件格式之间的智能转换</p>
          </div>
        </div>
        <el-button @click="showFormats" type="info" plain>
          <el-icon><InfoFilled /></el-icon>
          支持格式
        </el-button>
      </div>
    </el-card>

    <el-row :gutter="20">
      <el-col :span="12">
        <el-card class="input-card">
          <template #header>
            <div class="card-header">
              <span>输入内容</span>
              <el-select v-model="sourceFormat" placeholder="选择源格式" style="width: 120px">
                <el-option
                  v-for="format in availableFormats"
                  :key="format"
                  :label="format.toUpperCase()"
                  :value="format"
                />
              </el-select>
            </div>
          </template>
          
          <div class="input-section">
            <el-upload
              v-if="!inputContent"
              class="upload-demo"
              drag
              :auto-upload="false"
              :on-change="handleFileUpload"
              :show-file-list="false"
            >
              <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
              <div class="el-upload__text">
                将文件拖到此处，或<em>点击上传</em>
              </div>
              <template #tip>
                <div class="el-upload__tip">
                  支持 txt, json, xml, csv, yaml 等格式
                </div>
              </template>
            </el-upload>

            <div v-else class="input-content">
              <el-input
                v-model="inputContent"
                type="textarea"
                :rows="15"
                placeholder="请输入要转换的内容..."
                show-word-limit
                :maxlength="10000"
              />
              <div class="input-actions">
                <el-button @click="clearInput" size="small">清空</el-button>
                <el-button @click="pasteFromClipboard" size="small" type="primary">
                  <el-icon><DocumentCopy /></el-icon>
                  粘贴
                </el-button>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="12">
        <el-card class="output-card">
          <template #header>
            <div class="card-header">
              <span>转换结果</span>
              <div class="output-actions">
                <el-select v-model="targetFormat" placeholder="选择目标格式" style="width: 120px">
                  <el-option
                    v-for="format in availableFormats"
                    :key="format"
                    :label="format.toUpperCase()"
                    :value="format"
                  />
                </el-select>
                <el-button 
                  @click="convertFormat" 
                  type="primary" 
                  :loading="converting"
                  :disabled="!inputContent || !sourceFormat || !targetFormat"
                >
                  <el-icon><Refresh /></el-icon>
                  转换
                </el-button>
              </div>
            </div>
          </template>

          <div class="output-section">
            <el-input
              v-model="outputContent"
              type="textarea"
              :rows="15"
              placeholder="转换结果将显示在这里..."
              readonly
            />
            <div class="output-actions" v-if="outputContent">
              <el-button @click="copyToClipboard" size="small" type="success">
                <el-icon><CopyDocument /></el-icon>
                复制
              </el-button>
              <el-button @click="downloadResult" size="small" type="primary">
                <el-icon><Download /></el-icon>
                下载
              </el-button>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 支持格式对话框 -->
    <el-dialog v-model="formatsDialogVisible" title="支持的格式" width="600px">
      <div class="formats-content">
        <div v-for="(formats, category) in supportedFormats" :key="category" class="format-category">
          <h4>{{ getCategoryName(category) }}</h4>
          <el-tag v-for="format in formats" :key="format" class="format-tag">
            {{ format.toUpperCase() }}
          </el-tag>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import {
  DocumentCopy,
  InfoFilled,
  UploadFilled,
  Refresh,
  CopyDocument,
  Download
} from '@element-plus/icons-vue'
import api from '../utils/api'

const inputContent = ref('')
const outputContent = ref('')
const sourceFormat = ref('json')
const targetFormat = ref('yaml')
const converting = ref(false)
const formatsDialogVisible = ref(false)

const supportedFormats = ref({
  document: ['txt', 'md', 'html', 'json', 'xml', 'csv', 'yaml'],
  code: ['py', 'js', 'go', 'java', 'cpp', 'c', 'php', 'rb', 'rs'],
  data: ['json', 'xml', 'csv', 'yaml', 'toml', 'ini'],
  markup: ['html', 'xml', 'md', 'rst', 'tex']
})

const availableFormats = computed(() => {
  const allFormats = new Set()
  Object.values(supportedFormats.value).forEach(formats => {
    formats.forEach(format => allFormats.add(format))
  })
  return Array.from(allFormats).sort()
})

const getCategoryName = (category) => {
  const names = {
    document: '文档格式',
    code: '代码格式',
    data: '数据格式',
    markup: '标记语言'
  }
  return names[category] || category
}

const handleFileUpload = (file) => {
  const reader = new FileReader()
  reader.onload = (e) => {
    inputContent.value = e.target.result
    // 根据文件扩展名自动设置源格式
    const extension = file.name.split('.').pop().toLowerCase()
    if (availableFormats.value.includes(extension)) {
      sourceFormat.value = extension
    }
  }
  reader.readAsText(file.raw)
}

const clearInput = () => {
  inputContent.value = ''
  outputContent.value = ''
}

const pasteFromClipboard = async () => {
  try {
    const text = await navigator.clipboard.readText()
    inputContent.value = text
    ElMessage.success('已从剪贴板粘贴内容')
  } catch (err) {
    ElMessage.error('无法访问剪贴板')
  }
}

const convertFormat = async () => {
  if (!inputContent.value.trim()) {
    ElMessage.warning('请输入要转换的内容')
    return
  }

  converting.value = true
  try {
    const response = await api.post('/v1/tasks/convert', {
      source_format: sourceFormat.value,
      target_format: targetFormat.value,
      content: inputContent.value
    })

    if (response.data.success) {
      outputContent.value = response.data.converted_content
      ElMessage.success('转换成功')
    } else {
      ElMessage.error(response.data.message || '转换失败')
    }
  } catch (error) {
    console.error('转换错误:', error)
    ElMessage.error('转换失败: ' + (error.response?.data?.error || error.message))
  } finally {
    converting.value = false
  }
}

const copyToClipboard = async () => {
  try {
    await navigator.clipboard.writeText(outputContent.value)
    ElMessage.success('已复制到剪贴板')
  } catch (err) {
    ElMessage.error('复制失败')
  }
}

const downloadResult = () => {
  const blob = new Blob([outputContent.value], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `converted.${targetFormat.value}`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
  ElMessage.success('文件已下载')
}

const showFormats = () => {
  formatsDialogVisible.value = true
}

const loadSupportedFormats = async () => {
  try {
    const response = await api.get('/v1/tasks/formats')
    if (response.data.supported_formats) {
      supportedFormats.value = response.data.supported_formats
    }
  } catch (error) {
    console.error('加载支持格式失败:', error)
  }
}

onMounted(() => {
  loadSupportedFormats()
})
</script>

<style scoped>
.converter-container {
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
  color: #409eff;
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

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.input-card, .output-card {
  height: 600px;
}

.input-section, .output-section {
  height: 500px;
  display: flex;
  flex-direction: column;
}

.upload-demo {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.input-content {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.input-actions, .output-actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
  margin-top: 10px;
}

.formats-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.format-category h4 {
  margin: 0 0 10px 0;
  color: #303133;
}

.format-tag {
  margin: 0 8px 8px 0;
}
</style>
