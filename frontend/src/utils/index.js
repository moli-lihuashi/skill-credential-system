/**
 * 全站共用常量与工具 —— 消除各页面重复的 typeLabel / 复制逻辑
 */

// 成果类型
export const ACH_TYPES = {
  certificate: '技能证书',
  competition: '竞赛获奖',
  internship: '岗位实训',
  project: '项目实践'
}

// 角色
export const ROLE_LABELS = {
  student: '学生',
  school_admin: '院校管理员',
  company: '实训企业',
  employer: '招聘企业'
}

// 登录后各角色首页
export const ROLE_HOME = {
  student: '/student/home',
  school_admin: '/admin/home',
  company: '/company/home',
  employer: '/employer/verify'
}

// 演示数据（答辩演示用，见 docs/演示操作指南.md）
export const DEMO_CREDENTIAL = {
  credential_no: 'CC-20260927120316-A3D0B3',
  file_hash: 'aa193a9593bf51cff846bc74f8a01d9a5ccc8103ce042ec4f44df72fd8520bb8',
  student_name: '张伟',
  student_no: '2023010101',
  title: '全国职业院校技能大赛一等奖'
}

export function typeLabel(t) {
  return ACH_TYPES[t] || t
}

/** 安全读取 localStorage 中的 user 对象，损坏时返回 {} */
export function getStoredUser() {
  try {
    return JSON.parse(localStorage.getItem('user') || '{}') || {}
  } catch {
    return {}
  }
}

/** 复制文本到剪贴板（含降级方案） */
export async function copyText(text) {
  if (!text) return false
  try {
    await navigator.clipboard.writeText(text)
    return true
  } catch {
    const ta = document.createElement('textarea')
    ta.value = text
    ta.style.position = 'fixed'
    ta.style.opacity = '0'
    document.body.appendChild(ta)
    ta.select()
    try {
      // execCommand 失败时返回 false 而不是抛异常，需如实返回
      return document.execCommand('copy')
    } catch {
      return false
    } finally {
      document.body.removeChild(ta)
    }
  }
}

/** 把哈希末位十六进制取反，模拟「文件被篡改」的错误哈希 */
export function tamperHash(hash) {
  if (!hash) return hash
  const last = hash.slice(-1).toLowerCase()
  const flipped = { 0: '1', f: 'e' }[last] || '0'
  return hash.slice(0, -1) + flipped
}

/** 截断显示长哈希/地址 */
export function shortHex(hex, head = 16, tail = 0) {
  if (!hex) return ''
  if (hex.length <= head + tail + 1) return hex
  return tail
    ? `${hex.slice(0, head)}…${hex.slice(-tail)}`
    : `${hex.slice(0, head)}…`
}

/**
 * 统计数字 count-up（21st Stats 模式）
 * 尊重 prefers-reduced-motion：用户要求减少动效时直接跳到终值
 */
export function countUp(from, to, ms = 700, onUpdate) {
  const reduce =
    typeof window !== 'undefined' &&
    window.matchMedia &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches
  if (reduce || from === to) {
    onUpdate(to)
    return () => {}
  }
  let raf = 0
  const start = performance.now()
  const tick = (now) => {
    const t = Math.min(1, (now - start) / ms)
    // ease-out
    const eased = 1 - Math.pow(1 - t, 3)
    onUpdate(Math.round(from + (to - from) * eased))
    if (t < 1) raf = requestAnimationFrame(tick)
  }
  raf = requestAnimationFrame(tick)
  return () => cancelAnimationFrame(raf)
}

/** 给 .stat 元素绑定 spotlight 光标跟随（mouseenter/mousemove） */
export function bindSpotlight(el) {
  if (!el) return () => {}
  const move = (e) => {
    const r = el.getBoundingClientRect()
    el.style.setProperty('--mx', `${e.clientX - r.left}px`)
    el.style.setProperty('--my', `${e.clientY - r.top}px`)
  }
  el.addEventListener('mousemove', move)
  return () => el.removeEventListener('mousemove', move)
}
