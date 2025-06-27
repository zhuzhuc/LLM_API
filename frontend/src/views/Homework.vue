<template>
  <div class="homework-container">
    <el-card class="service-header">
      <div class="header-content">
        <div class="header-left">
          <el-icon class="service-icon"><EditPen /></el-icon>
          <div>
            <h2>AI 作业批改</h2>
            <p>智能批改作业，提供详细反馈和改进建议</p>
          </div>
        </div>
        <div class="header-stats">
          <el-statistic title="今日批改" :value="todayCount" />
          <el-statistic title="总计批改" :value="totalCount" />
        </div>
      </div>
    </el-card>

    <el-row :gutter="20">
      <el-col :span="14">
        <el-card class="homework-form">
          <template #header>
            <span>作业信息</span>
          </template>

          <el-form :model="homeworkForm" label-width="100px" class="homework-form-content">
            <el-form-item label="科目">
              <el-select v-model="homeworkForm.subject" placeholder="请选择科目" style="width: 200px">
                <el-option label="数学" value="数学" />
                <el-option label="语文" value="语文" />
                <el-option label="英语" value="英语" />
                <el-option label="物理" value="物理" />
                <el-option label="化学" value="化学" />
                <el-option label="生物" value="生物" />
                <el-option label="历史" value="历史" />
                <el-option label="地理" value="地理" />
                <el-option label="政治" value="政治" />
                <el-option label="其他" value="其他" />
              </el-select>
            </el-form-item>

            <el-form-item label="年级">
              <el-select v-model="homeworkForm.grade_level" placeholder="请选择年级" style="width: 200px">
                <el-option label="小学" value="小学" />
                <el-option label="初中" value="初中" />
                <el-option label="高中" value="高中" />
                <el-option label="大学" value="大学" />
              </el-select>
            </el-form-item>

            <el-form-item label="语言">
              <el-radio-group v-model="homeworkForm.language">
                <el-radio value="中文">中文</el-radio>
                <el-radio value="英文">英文</el-radio>
              </el-radio-group>
            </el-form-item>

            <el-form-item label="题目">
              <el-input
                v-model="homeworkForm.question"
                type="textarea"
                :rows="4"
                placeholder="请输入题目内容..."
                show-word-limit
                :maxlength="1000"
              />
            </el-form-item>

            <el-form-item label="学生答案">
              <el-input
                v-model="homeworkForm.answer"
                type="textarea"
                :rows="6"
                placeholder="请输入学生的答案..."
                show-word-limit
                :maxlength="2000"
              />
            </el-form-item>

            <el-form-item>
              <el-button 
                type="primary" 
                @click="gradeHomework"
                :loading="grading"
                :disabled="!canSubmit"
                size="large"
              >
                <el-icon><Check /></el-icon>
                开始批改
              </el-button>
              <el-button @click="resetForm" size="large">
                <el-icon><RefreshLeft /></el-icon>
                重置
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>

      <el-col :span="10">
        <el-card class="result-card" v-if="gradingResult">
          <template #header>
            <div class="result-header">
              <span>批改结果</span>
              <el-tag :type="getScoreType(gradingResult.score)" size="large">
                {{ gradingResult.score }} 分
              </el-tag>
            </div>
          </template>

          <div class="result-content">
            <div class="score-section">
              <el-progress 
                :percentage="gradingResult.score" 
                :color="getProgressColor(gradingResult.score)"
                :stroke-width="20"
                text-inside
              />
            </div>

            <div class="feedback-section">
              <h4>详细反馈</h4>
              <div class="feedback-content">
                {{ gradingResult.feedback }}
              </div>
            </div>

            <div class="suggestions-section" v-if="gradingResult.suggestions?.length">
              <h4>改进建议</h4>
              <ul class="suggestions-list">
                <li v-for="(suggestion, index) in gradingResult.suggestions" :key="index">
                  {{ suggestion }}
                </li>
              </ul>
            </div>

            <div class="result-actions">
              <el-button @click="saveResult" type="success" size="small">
                <el-icon><FolderAdd /></el-icon>
                保存结果
              </el-button>
              <el-button @click="exportResult" type="primary" size="small">
                <el-icon><Download /></el-icon>
                导出报告
              </el-button>
            </div>
          </div>
        </el-card>

        <el-card v-else class="placeholder-card">
          <div class="placeholder-content">
            <el-icon class="placeholder-icon"><EditPen /></el-icon>
            <p>填写作业信息后，点击"开始批改"查看结果</p>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 历史记录 -->
    <el-card class="history-card" v-if="historyList.length">
      <template #header>
        <div class="history-header">
          <span>批改历史</span>
          <el-button @click="loadHistory" type="text" size="small">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>

      <el-table :data="historyList" style="width: 100%">
        <el-table-column prop="subject" label="科目" width="80" />
        <el-table-column prop="grade_level" label="年级" width="80" />
        <el-table-column prop="question" label="题目" show-overflow-tooltip />
        <el-table-column prop="score" label="分数" width="80">
          <template #default="scope">
            <el-tag :type="getScoreType(scope.row.score)">
              {{ scope.row.score }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="批改时间" width="180" />
        <el-table-column label="操作" width="120">
          <template #default="scope">
            <el-button @click="viewDetail(scope.row)" type="text" size="small">
              查看详情
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
  EditPen,
  Check,
  RefreshLeft,
  FolderAdd,
  Download,
  Refresh
} from '@element-plus/icons-vue'
import api from '../utils/api'

const homeworkForm = ref({
  subject: '',
  grade_level: '',
  language: '中文',
  question: '',
  answer: ''
})

const gradingResult = ref(null)
const grading = ref(false)
const todayCount = ref(0)
const totalCount = ref(0)
const historyList = ref([])

const canSubmit = computed(() => {
  return homeworkForm.value.subject && 
         homeworkForm.value.grade_level && 
         homeworkForm.value.question.trim() && 
         homeworkForm.value.answer.trim()
})

const getScoreType = (score) => {
  if (score >= 90) return 'success'
  if (score >= 80) return 'primary'
  if (score >= 70) return 'warning'
  if (score >= 60) return ''
  return 'danger'
}

const getProgressColor = (score) => {
  if (score >= 90) return '#67c23a'
  if (score >= 80) return '#409eff'
  if (score >= 70) return '#e6a23c'
  if (score >= 60) return '#f56c6c'
  return '#f56c6c'
}

const gradeHomework = async () => {
  grading.value = true
  try {
    const response = await api.post('/v1/tasks/homework', homeworkForm.value)
    
    if (response.data.success) {
      gradingResult.value = response.data
      ElMessage.success('批改完成')
      todayCount.value++
      totalCount.value++
    } else {
      ElMessage.error(response.data.message || '批改失败')
    }
  } catch (error) {
    console.error('批改错误:', error)
    ElMessage.error('批改失败: ' + (error.response?.data?.error || error.message))
  } finally {
    grading.value = false
  }
}

const resetForm = () => {
  homeworkForm.value = {
    subject: '',
    grade_level: '',
    language: '中文',
    question: '',
    answer: ''
  }
  gradingResult.value = null
}

const saveResult = () => {
  // 保存到本地存储或发送到服务器
  const result = {
    ...homeworkForm.value,
    ...gradingResult.value,
    created_at: new Date().toLocaleString()
  }
  
  const saved = JSON.parse(localStorage.getItem('homework_history') || '[]')
  saved.unshift(result)
  localStorage.setItem('homework_history', JSON.stringify(saved.slice(0, 50))) // 只保留最近50条
  
  historyList.value = saved
  ElMessage.success('结果已保存')
}

const exportResult = () => {
  const content = `
作业批改报告
================

科目: ${homeworkForm.value.subject}
年级: ${homeworkForm.value.grade_level}
语言: ${homeworkForm.value.language}

题目:
${homeworkForm.value.question}

学生答案:
${homeworkForm.value.answer}

批改结果:
分数: ${gradingResult.value.score}/100

详细反馈:
${gradingResult.value.feedback}

改进建议:
${gradingResult.value.suggestions?.map((s, i) => `${i + 1}. ${s}`).join('\n') || '无'}

批改时间: ${new Date().toLocaleString()}
  `.trim()

  const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `作业批改报告_${homeworkForm.value.subject}_${new Date().toLocaleDateString()}.txt`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
  ElMessage.success('报告已导出')
}

const loadHistory = () => {
  const saved = JSON.parse(localStorage.getItem('homework_history') || '[]')
  historyList.value = saved
}

const viewDetail = (row) => {
  homeworkForm.value = {
    subject: row.subject,
    grade_level: row.grade_level,
    language: row.language,
    question: row.question,
    answer: row.answer
  }
  gradingResult.value = {
    score: row.score,
    feedback: row.feedback,
    suggestions: row.suggestions
  }
}

onMounted(() => {
  loadHistory()
  // 加载统计数据
  const saved = JSON.parse(localStorage.getItem('homework_history') || '[]')
  totalCount.value = saved.length
  
  const today = new Date().toDateString()
  todayCount.value = saved.filter(item => 
    new Date(item.created_at).toDateString() === today
  ).length
})
</script>

<style scoped>
.homework-container {
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

.header-stats {
  display: flex;
  gap: 40px;
}

.service-icon {
  font-size: 32px;
  color: #67c23a;
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

.homework-form {
  height: fit-content;
}

.homework-form-content {
  padding: 20px 0;
}

.result-card {
  height: fit-content;
  min-height: 400px;
}

.placeholder-card {
  height: 400px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.placeholder-content {
  text-align: center;
  color: #909399;
}

.placeholder-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.result-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.score-section {
  text-align: center;
}

.feedback-section h4,
.suggestions-section h4 {
  margin: 0 0 10px 0;
  color: #303133;
}

.feedback-content {
  background: #f5f7fa;
  padding: 15px;
  border-radius: 4px;
  line-height: 1.6;
}

.suggestions-list {
  margin: 0;
  padding-left: 20px;
}

.suggestions-list li {
  margin-bottom: 8px;
  line-height: 1.5;
}

.result-actions {
  display: flex;
  gap: 10px;
  justify-content: center;
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
