import { createRouter, createWebHistory } from 'vue-router'
import { ROLE_HOME, getStoredUser } from '../utils'

const routes = [
  {
    path: '/',
    redirect: '/login'
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue')
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('../views/Register.vue')
  },
  // 公开核验页（招聘企业/游客）
  {
    path: '/verify',
    name: 'PublicVerify',
    component: () => import('../views/VerifyPublic.vue')
  },
  // 学生端
  {
    path: '/student',
    component: () => import('../layouts/MainLayout.vue'),
    meta: { roles: ['student'] },
    children: [
      { path: '', redirect: '/student/home' },
      { path: 'home', component: () => import('../views/student/Home.vue') },
      { path: 'credentials', component: () => import('../views/student/Credentials.vue') }
    ]
  },
  // 院校管理端
  {
    path: '/admin',
    component: () => import('../layouts/MainLayout.vue'),
    meta: { roles: ['school_admin'] },
    children: [
      { path: '', redirect: '/admin/home' },
      { path: 'home', component: () => import('../views/admin/Home.vue') },
      { path: 'achievements', component: () => import('../views/admin/Achievements.vue') },
      { path: 'credentials', component: () => import('../views/admin/Credentials.vue') },
      { path: 'logs', component: () => import('../views/admin/Logs.vue') }
    ]
  },
  // 实训企业端
  {
    path: '/company',
    component: () => import('../layouts/MainLayout.vue'),
    meta: { roles: ['company'] },
    children: [
      { path: '', redirect: '/company/home' },
      { path: 'home', component: () => import('../views/company/Home.vue') },
      { path: 'achievements', component: () => import('../views/company/Achievements.vue') }
    ]
  },
  // 招聘企业核验端
  {
    path: '/employer',
    component: () => import('../layouts/MainLayout.vue'),
    meta: { roles: ['employer'] },
    children: [
      { path: '', redirect: '/employer/verify' },
      { path: 'verify', component: () => import('../views/employer/Verify.vue') },
      { path: 'history', component: () => import('../views/employer/History.vue') }
    ]
  },
  // 404 兜底：未匹配路径跳回登录，避免空白页
  {
    path: '/:pathMatch(.*)*',
    redirect: '/login'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫：登录与角色权限
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  const user = getStoredUser()

  if (to.path === '/login' || to.path === '/register' || to.path === '/verify') {
    return next()
  }

  if (!token) {
    return next('/login')
  }

  if (to.meta?.roles && !to.meta.roles.includes(user.role)) {
    // 角色不匹配则跳到对应首页
    return next(ROLE_HOME[user.role] || '/login')
  }

  next()
})

export default router
