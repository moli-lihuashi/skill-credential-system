<template>
  <el-container style="min-height: 100vh">
    <!-- 侧边栏 -->
    <el-aside width="220px" class="sidebar">
      <div class="logo">
        <div class="logo-mark">链</div>
        <div class="logo-text">
          <h2>技能成果存证</h2>
          <p>联盟链可信存证平台</p>
        </div>
      </div>
      <el-menu
        :default-active="$route.path"
        router
        class="side-menu"
        background-color="transparent"
        text-color="rgba(255,255,255,0.72)"
        active-text-color="#ffffff"
      >
        <template v-for="item in menus" :key="item.path">
          <el-menu-item :index="item.path" class="side-item">
            <el-icon><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </el-menu-item>
        </template>
      </el-menu>

      <div class="sidebar-foot">
        <p>数据哈希上链 · 不可篡改</p>
      </div>
    </el-aside>

    <el-container>
      <!-- 顶栏 -->
      <el-header class="header">
        <div class="header-title">{{ roleLabel }}</div>
        <div class="header-right">
          <el-tag type="info" effect="plain" class="addr-tag" @click="copyAddr">
            链账户 {{ shortHex(user.eth_address, 10, 4) }}
          </el-tag>
          <span class="user-name">{{ user.real_name || user.username }}</span>
          <el-button type="danger" size="small" plain @click="logout">退出</el-button>
        </div>
      </el-header>

      <!-- 主内容 -->
      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  HomeFilled, Document, Medal, Search, List
} from '@element-plus/icons-vue'
import { ROLE_LABELS, copyText, shortHex, getStoredUser } from '../utils'

const route = useRoute()
const router = useRouter()
const user = getStoredUser()

const roleMenus = {
  student: [
    { path: '/student/home', label: '个人中心', icon: HomeFilled },
    { path: '/student/credentials', label: '我的存证', icon: Medal }
  ],
  school_admin: [
    { path: '/admin/home', label: '工作台', icon: HomeFilled },
    { path: '/admin/achievements', label: '成果录入', icon: Document },
    { path: '/admin/credentials', label: '存证管理', icon: Medal },
    { path: '/admin/logs', label: '操作日志', icon: List }
  ],
  company: [
    { path: '/company/home', label: '工作台', icon: HomeFilled },
    { path: '/company/achievements', label: '实训成果录入', icon: Document }
  ],
  employer: [
    { path: '/employer/verify', label: '凭证核验', icon: Search },
    { path: '/employer/history', label: '核验历史', icon: List }
  ]
}

const roleLabels = {
  student: '学生端',
  school_admin: '院校管理端',
  company: '实训企业端',
  employer: '企业核验端'
}

const menus = computed(() => roleMenus[user.role] || [])
const roleLabel = computed(() => roleLabels[user.role] || ROLE_LABELS[user.role] || '')

async function copyAddr() {
  const ok = await copyText(user.eth_address)
  ElMessage[ok ? 'success' : 'error'](ok ? '链账户地址已复制' : '复制失败')
}

function logout() {
  localStorage.removeItem('token')
  localStorage.removeItem('user')
  router.push('/login')
}
</script>

<style scoped>
/* 侧边栏：信任蓝深阶，几何极简 */
.sidebar {
  background: linear-gradient(180deg, var(--brand-primary) 0%, var(--brand-primary-deep) 100%);
  display: flex;
  flex-direction: column;
}
.logo {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 20px 16px;
  color: #fff;
  border-bottom: 1px solid rgba(255, 255, 255, .12);
}
.logo-mark {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: rgba(255, 255, 255, .16);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 16px;
  flex-shrink: 0;
  /* 光学对齐：汉字略偏上更稳 */
  line-height: 1;
  padding-top: 1px;
}
.logo-text {
  /* 文字块与图标光学居中 */
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-width: 0;
}
.logo h2 { font-size: 15px; font-weight: 600; margin-bottom: 2px; line-height: 1.3; }
.logo p { font-size: 11px; color: rgba(255, 255, 255, .7); letter-spacing: 0.02em; }

.side-menu {
  border-right: none;
  padding: 8px 0;
  flex: 1;
}
.side-item {
  margin: 4px 10px;
  border-radius: 8px;
  height: 44px;
  line-height: 44px;
}
.side-item.is-active {
  /* 成就金高亮：当前页身份感 */
  background: rgba(255, 255, 255, .18) !important;
  color: #fff !important;
  box-shadow: inset 3px 0 0 var(--brand-accent);
}

.sidebar-foot {
  padding: 16px;
  color: rgba(255, 255, 255, .55);
  font-size: 11px;
  text-align: center;
  letter-spacing: 0.04em;
  border-top: 1px solid rgba(255, 255, 255, .1);
}

.header {
  background: var(--bg-card);
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--border);
  box-shadow: none;
}
.header-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--brand-primary-deep);
  letter-spacing: 0.01em;
}
.header-right { display: flex; align-items: center; gap: 12px; }
.addr-tag {
  cursor: pointer;
  font-family: var(--font-mono);
  font-size: 12px;
}
.addr-tag:hover {
  border-color: var(--brand-primary);
  color: var(--brand-primary);
}
.user-name { color: var(--text-2); font-size: 14px; }
.main {
  padding: var(--space-5);
  background: var(--bg-page);
}
</style>
