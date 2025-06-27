import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '../utils/api'

export const useUserStore = defineStore('user', () => {
  const user = ref(null)
  const token = ref(localStorage.getItem('token') || '')
  
  const isLoggedIn = computed(() => !!token.value)
  
  // 登录
  const login = async (credentials) => {
    try {
      const response = await api.post('/v1/auth/login', credentials)
      console.log('登录响应:', response.data)
      const { token: newToken, user: userData } = response.data

      console.log('提取的 token:', newToken)
      console.log('提取的 user:', userData)

      token.value = newToken
      user.value = userData
      localStorage.setItem('token', newToken)

      console.log('存储后的状态 - token:', token.value)
      console.log('存储后的状态 - user:', user.value)
      console.log('isLoggedIn:', !!token.value)

      return { success: true }
    } catch (error) {
      console.error('登录错误:', error)
      return {
        success: false,
        message: error.response?.data?.error || '登录失败'
      }
    }
  }

  // 注册
  const register = async (userData) => {
    try {
      const response = await api.post('/v1/auth/register', userData)
      return { success: true, message: '注册成功' }
    } catch (error) {
      return {
        success: false,
        message: error.response?.data?.error || '注册失败'
      }
    }
  }

  // 获取用户信息
  const fetchProfile = async () => {
    try {
      const response = await api.get('/v1/auth/profile')
      user.value = response.data
      return { success: true }
    } catch (error) {
      return {
        success: false,
        message: error.response?.data?.error || '获取用户信息失败'
      }
    }
  }
  
  // 退出登录
  const logout = () => {
    user.value = null
    token.value = ''
    localStorage.removeItem('token')
  }
  
  // 更新用户token数量
  const updateTokens = (newTokenCount) => {
    if (user.value) {
      user.value.tokens = newTokenCount
    }
  }
  
  // 初始化时如果有token则获取用户信息
  const init = async () => {
    if (token.value) {
      await fetchProfile()
    }
  }
  
  return {
    user,
    token,
    isLoggedIn,
    login,
    register,
    fetchProfile,
    logout,
    updateTokens,
    init
  }
})