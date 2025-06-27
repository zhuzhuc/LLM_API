<template>
  <div class="profile-container">
    <div class="profile-content">
      <el-card class="profile-card">
        <template #header>
          <div class="card-header">
            <h2>个人资料</h2>
          </div>
        </template>
        
        <div class="profile-info">
          <div class="avatar-section">
            <el-avatar :size="80" :icon="UserFilled" />
            <h3>{{ userStore.user?.username }}</h3>
          </div>
          
          <el-descriptions :column="1" border>
            <el-descriptions-item label="用户名">
              {{ userStore.user?.username }}
            </el-descriptions-item>
            <el-descriptions-item label="邮箱">
              {{ userStore.user?.email }}
            </el-descriptions-item>
            <el-descriptions-item label="当前Token">
              <el-tag type="success" size="large">
                <el-icon><Coin /></el-icon>
                {{ userStore.user?.tokens || 0 }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="注册时间">
              {{ formatDate(userStore.user?.created_at) }}
            </el-descriptions-item>
            <el-descriptions-item label="最后更新">
              {{ formatDate(userStore.user?.updated_at) }}
            </el-descriptions-item>
          </el-descriptions>
        </div>
      </el-card>
      
      <el-card class="token-card">
        <template #header>
          <div class="card-header">
            <h3>Token 管理</h3>
          </div>
        </template>
        
        <div class="token-info">
          <div class="token-display">
            <div class="token-amount">
              <el-icon size="24"><Coin /></el-icon>
              <span class="amount">{{ userStore.user?.tokens || 0 }}</span>
              <span class="label">可用 Tokens</span>
            </div>
          </div>
          
          <el-divider />
          
          <div class="token-actions">
            <el-alert
              title="Token 说明"
              type="info"
              :closable="false"
              show-icon
            >
              <p>• 每次对话会根据消息长度消耗相应的 Token</p>
              <p>• Token 不足时无法继续对话</p>
              <p>• 新用户注册赠送 1000 个 Token</p>
            </el-alert>
            
            <div class="recharge-section">
              <h4>Token 充值</h4>
              <p class="recharge-note">如需更多 Token，请联系管理员</p>
              <el-button type="primary" disabled>
                联系管理员充值
              </el-button>
            </div>
          </div>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useUserStore } from '../stores/user'
import { UserFilled, Coin } from '@element-plus/icons-vue'

const userStore = useUserStore()

// 格式化日期
const formatDate = (dateString) => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 刷新用户信息
onMounted(async () => {
  await userStore.fetchProfile()
})
</script>

<style scoped>
.profile-container {
  min-height: 100vh;
  background: #f5f7fa;
  padding: 20px;
}

.profile-content {
  max-width: 800px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.profile-card,
.token-card {
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h2,
.card-header h3 {
  margin: 0;
  color: #333;
}

.profile-info {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.avatar-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 20px 0;
}

.avatar-section h3 {
  margin: 0;
  color: #333;
  font-size: 20px;
}

.token-info {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.token-display {
  display: flex;
  justify-content: center;
  padding: 20px 0;
}

.token-amount {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 24px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-radius: 12px;
  min-width: 200px;
}

.amount {
  font-size: 32px;
  font-weight: 600;
}

.label {
  font-size: 14px;
  opacity: 0.9;
}

.token-actions {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.recharge-section {
  text-align: center;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
}

.recharge-section h4 {
  margin: 0 0 8px 0;
  color: #333;
}

.recharge-note {
  margin: 0 0 16px 0;
  color: #666;
  font-size: 14px;
}

:deep(.el-descriptions__label) {
  font-weight: 500;
  color: #333;
}

:deep(.el-descriptions__content) {
  color: #666;
}

:deep(.el-alert__content) {
  line-height: 1.6;
}

:deep(.el-alert__content p) {
  margin: 4px 0;
}
</style>