<template>
  <div class="anim-login">
    <!-- 左侧：品牌 + 动画角色 -->
    <div class="left-panel">
      <div class="left-top">
        <div class="brand-mark">
          <svg width="28" height="28" viewBox="0 0 28 28" fill="none">
            <rect width="28" height="28" rx="7" fill="white" fill-opacity="0.15" />
            <path d="M7 14L12 9L17 14L12 19L7 14Z" fill="white" fill-opacity="0.9" />
            <path d="M13 14L18 9L21 12V16L18 19L13 14Z" fill="white" fill-opacity="0.5" />
          </svg>
        </div>
        <span class="brand-name">技能成果存证</span>
      </div>

      <div class="characters-area">
        <AnimatedCharacters
          :is-typing="isTyping"
          :show-password="showPassword"
          :password-length="form.password.length"
          :is-password-focused="isPasswordFocused"
        />
      </div>

      <div class="left-footer">
        <span>联盟链 · 不可篡改</span>
        <span>SHA-256 上链存证</span>
      </div>

      <div class="decor-blur1" />
      <div class="decor-blur2" />
      <div class="decor-grid" />
    </div>

    <!-- 右侧：蓝白底 + 登录表单 -->
    <div class="right-panel">
      <div class="form-wrapper">
        <div class="form-header">
          <h1 class="form-title">登录到工作台</h1>
          <p class="form-subtitle">技能成果可信系统</p>
        </div>

        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          class="form"
          @submit.prevent="handleLogin"
        >
          <div class="field-label">账号</div>
          <el-form-item prop="username">
            <el-input
              v-model="form.username"
              placeholder="输入您的账号"
              size="large"
              class="anim-input"
              @focus="isTyping = true"
              @blur="isTyping = false"
            >
              <template #prefix>
                <el-icon class="prefix-icon"><User /></el-icon>
              </template>
            </el-input>
          </el-form-item>

          <div class="field-label">密码</div>
          <el-form-item prop="password">
            <el-input
              v-model="form.password"
              :type="showPassword ? 'text' : 'password'"
              placeholder="输入您的密码"
              size="large"
              class="anim-input"
              @focus="isPasswordFocused = true"
              @blur="isPasswordFocused = false"
              @keyup.enter="handleLogin"
            >
              <template #prefix>
                <el-icon class="prefix-icon"><Lock /></el-icon>
              </template>
              <template #suffix>
                <span class="eye-toggle" @click="showPassword = !showPassword">
                  <el-icon><View v-if="showPassword" /><Hide v-else /></el-icon>
                </span>
              </template>
            </el-input>
          </el-form-item>

          <el-form-item style="margin-bottom: 0">
            <el-button
              type="primary"
              class="submit-btn"
              :loading="loading"
              style="width: 100%"
              @click="handleLogin"
            >
              {{ loading ? '登录中...' : '登录' }}
            </el-button>
          </el-form-item>
        </el-form>

        <div class="divider"><span>或</span></div>

        <!-- 测试账号一键登录（替换原飞书入口） -->
        <div class="demo-grid">
          <el-button
            v-for="d in demoAccounts"
            :key="d.username"
            class="demo-btn"
            block
            @click="fillDemo(d)"
          >
            <el-icon class="demo-icon"><UserFilled /></el-icon>
            {{ d.label }}一键登录
          </el-button>
        </div>

        <div class="extra-links">
          <router-link to="/register">注册账号</router-link>
          <router-link to="/verify">外部企业核验入口</router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock, View, Hide, UserFilled } from '@element-plus/icons-vue'
import AnimatedCharacters from '../components/animated-characters/Index.vue'
import api from '../api'
import { ROLE_HOME } from '../utils'

const router = useRouter()
const formRef = ref()
const loading = ref(false)
const showPassword = ref(false)
const isTyping = ref(false)
const isPasswordFocused = ref(false)

const form = reactive({ username: '', password: '' })

const rules = {
  username: [
    { required: true, message: '请输入账号', trigger: 'blur' },
    { min: 3, message: '账号长度不能少于 3 个字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于 6 个字符', trigger: 'blur' }
  ]
}

// 演示账号口令不写入源码：构建时通过 VITE_DEMO_PASSWORD 注入（见 frontend/.env，未提交）
const demoPassword = import.meta.env.VITE_DEMO_PASSWORD || ''

const demoAccounts = [
  { label: '学生', username: 'demo_student' },
  { label: '院校管理员', username: 'admin01' },
  { label: '实训企业', username: 'company01' },
  { label: '招聘企业', username: 'employer01' }
]

function fillDemo(d) {
  form.username = d.username
  form.password = demoPassword
  ElMessage.success(demoPassword
    ? `已填充「${d.label}」演示账号`
    : `已填充「${d.label}」用户名，请手动输入口令`)
}

async function handleLogin() {
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  loading.value = true
  try {
    const res = await api.post('/auth/login', form)
    localStorage.setItem('token', res.data.token)
    localStorage.setItem('user', JSON.stringify(res.data))
    ElMessage.success('登录成功')
    router.push(ROLE_HOME[res.data.role] || '/login')
  } catch (e) {
    ElMessage.error(e.msg || '账号或密码有误，请重新输入')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
/* ===== 布局（源自 animated-login）===== */
.anim-login {
  min-height: 100vh;
  display: grid;
  grid-template-columns: 1fr 1fr;
}
@media (max-width: 1024px) {
  .anim-login { grid-template-columns: 1fr; }
  .left-panel { display: none; }
}

/* ===== 左侧品牌区 ===== */
.left-panel {
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 48px;
  background: linear-gradient(145deg, #0f172a 0%, #1e3a8a 50%, #1e40af 100%);
  overflow: hidden;
}
.left-top {
  position: relative;
  z-index: 20;
  display: flex;
  align-items: center;
  gap: 10px;
  color: #fff;
}
.brand-mark {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.12);
  border: 1px solid rgba(255, 255, 255, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  backdrop-filter: blur(8px);
}
.brand-name {
  font-size: 20px;
  font-weight: 700;
  letter-spacing: 1px;
}
.characters-area {
  position: relative;
  z-index: 20;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  height: 500px;
}
.left-footer {
  position: relative;
  z-index: 20;
  display: flex;
  gap: 24px;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.5);
}
.decor-blur1 {
  position: absolute;
  top: 15%;
  right: 10%;
  width: 300px;
  height: 300px;
  background: rgba(59, 130, 246, 0.25);
  border-radius: 50%;
  filter: blur(80px);
  pointer-events: none;
  z-index: 0;
}
.decor-blur2 {
  position: absolute;
  bottom: 10%;
  left: 5%;
  width: 400px;
  height: 400px;
  background: rgba(30, 64, 175, 0.3);
  border-radius: 50%;
  filter: blur(100px);
  pointer-events: none;
  z-index: 0;
}
.decor-grid {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.03) 1px, transparent 1px);
  background-size: 40px 40px;
  pointer-events: none;
  z-index: 1;
}

/* ===== 右侧：蓝白底（原品牌页背景）===== */
.right-panel {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 32px;
  background-color: #ffffff;
  background-image:
    linear-gradient(to right, #f0f0f0 1px, transparent 1px),
    linear-gradient(to bottom, #f0f0f0 1px, transparent 1px);
  background-size: 6rem 4rem;
  position: relative;
}
.right-panel::before {
  content: '';
  position: absolute;
  inset: 0;
  background: radial-gradient(circle 500px at 50% 200px, #c9ebff, transparent);
  pointer-events: none;
}
.form-wrapper {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 400px;
}

/* ===== 表单（animated-login 样式）===== */
.form-header {
  text-align: center;
  margin-bottom: 40px;
}
.form-title {
  font-size: 26px;
  font-weight: 700;
  letter-spacing: -0.02em;
  color: #0f172a;
  margin: 0 0 10px;
  line-height: 1.3;
  text-wrap: balance;
}
.form-subtitle {
  font-size: 14px;
  color: #6b7280;
  margin: 0;
  line-height: 1.6;
}

.field-label {
  font-size: 13px;
  font-weight: 500;
  color: #374151;
  margin-bottom: 6px;
  letter-spacing: 0.2px;
}

.anim-input :deep(.el-input__wrapper) {
  height: 48px;
  background: #fafafa;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  box-shadow: none;
  transition: border-color 0.2s, box-shadow 0.2s;
}
.anim-input :deep(.el-input__wrapper:hover) {
  border-color: #3b82f6;
}
.anim-input :deep(.el-input__wrapper.is-focus) {
  border-color: #1e40af;
  box-shadow: 0 0 0 3px rgba(30, 64, 175, 0.08);
  background: #fff;
}
.anim-input :deep(.el-input__inner) {
  font-size: 14px;
  color: #111827;
}
.anim-input :deep(.el-input__inner::placeholder) {
  color: #c0c4cc;
}
.prefix-icon {
  color: #b0b7c3;
  font-size: 15px;
}
.eye-toggle {
  color: #6b7280;
  cursor: pointer;
  font-size: 16px;
  display: flex;
  align-items: center;
  transition: color 0.2s;
}
.eye-toggle:hover { color: #374151; }

.submit-btn {
  height: 48px;
  font-size: 15px;
  font-weight: 600;
  border-radius: 10px;
  background: #1e40af;
  border-color: #1e40af;
  letter-spacing: 1px;
}
.submit-btn:hover {
  background: #1d4ed8;
  border-color: #1d4ed8;
}

.divider {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 20px 0 0;
  font-size: 13px;
}
.divider::before,
.divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background: #e5e7eb;
}
.divider span {
  color: #9ca3af;
  white-space: nowrap;
}

/* 测试账号一键登录 */
.demo-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  margin-top: 12px;
}
.demo-btn {
  height: 44px;
  border-radius: 10px;
  border: 1px solid #e5e7eb;
  background: #fff;
  color: #374151;
  font-size: 13px;
  transition: background 0.2s, border-color 0.2s, color 0.2s;
}
.demo-btn:hover {
  background: #eff6ff;
  border-color: rgba(30, 64, 175, 0.25);
  color: #1e40af;
}
.demo-icon {
  margin-right: 4px;
  color: #1e40af;
}

.extra-links {
  display: flex;
  justify-content: space-between;
  margin-top: 24px;
  font-size: 13px;
}
.extra-links a {
  color: #1e40af;
  text-decoration: none;
}
.extra-links a:hover { text-decoration: underline; }
</style>
