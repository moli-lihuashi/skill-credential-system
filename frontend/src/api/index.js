import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 30000
})

// 请求拦截：自动携带 Token
api.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截：统一错误处理
api.interceptors.response.use(
  resp => resp.data,
  err => {
    const url = err.config?.url || ''
    // 登录/注册接口本身的 401/400 是业务结果（密码错误等），不能触发强制登出跳转
    const isAuthCall = url.includes('/auth/login') || url.includes('/auth/register')
    if (err.response?.status === 401 && !isAuthCall) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }
    return Promise.reject(err.response?.data || err)
  }
)

export default api
