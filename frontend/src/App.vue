<template>
  <router-view />
</template>

<script setup>
</script>

<style>
/* =========================================================
   设计令牌 — 来源：UI/UX Pro Max
   风格：Minimalism & Swiss（极简 · 栅格 · 高对比 · 功能优先）
   叙事：信任蓝（可信存证）+ 成就金（技能成果）
   ========================================================= */
:root {
  /* 主色：信任蓝 */
  --brand-primary: #0369a1;
  --brand-primary-dark: #075985;
  --brand-primary-deep: #0c4a6e;
  --brand-secondary: #0ea5e9;
  --brand-accent: #a16207;      /* 成就金 / CTA 点缀 */
  --brand-accent-soft: #fef9c3;

  /* 语义色 */
  --success: #15803d;
  --warning: #b45309;
  --danger: #dc2626;
  --info: #475569;

  /* 中性 */
  --bg-page: #f0f9ff;
  --bg-card: #ffffff;
  --bg-muted: #e7eff5;
  --text-1: #0c4a6e;
  --text-2: #475569;
  --text-3: #64748b;
  --border: #bae6fd;
  --ring: #0369a1;

  /* 间距（density 6/10 · Standard） */
  --space-1: 4px;
  --space-2: 8px;
  --space-3: 12px;
  --space-4: 16px;
  --space-5: 24px;
  --space-6: 32px;
  --space-7: 48px;

  /* 动效（subtle：200-250ms） */
  --ease: cubic-bezier(0.22, 1, 0.36, 1);
  --dur-fast: 180ms;
  --dur: 220ms;
  --dur-slow: 280ms;

  /* 字体：Fira Sans（拉丁）+ 系统 CJK；哈希/等宽用 Fira Code */
  --font-sans: 'Fira Sans', 'PingFang SC', 'Microsoft YaHei', 'Helvetica Neue', sans-serif;
  --font-mono: 'Fira Code', ui-monospace, Menlo, Consolas, monospace;

  /* 兼容旧命名（页面仍引用 --brand-700 等时逐步迁移） */
  --brand-900: var(--brand-primary-deep);
  --brand-800: var(--brand-primary-dark);
  --brand-700: var(--brand-primary);
  --brand-600: var(--brand-secondary);
  --accent: var(--brand-primary);
  --accent-hover: var(--brand-secondary);
  --accent-soft: #e0f2fe;
}

/* ===== Element Plus 主色对齐信任蓝 ===== */
:root {
  --el-color-primary: #0369a1;
  --el-color-primary-light-3: #3887b8;
  --el-color-primary-light-5: #6aa4c8;
  --el-color-primary-light-7: #9cc2d9;
  --el-color-primary-light-8: #b5d2e2;
  --el-color-primary-light-9: #e7f2f8;
  --el-color-primary-dark-2: #025480;
  --el-color-success: #15803d;
  --el-color-warning: #b45309;
  --el-color-danger: #dc2626;
  --el-color-error: #dc2626;
  --el-border-radius-base: 6px;
  --el-border-radius-small: 4px;
  --el-font-family: var(--font-sans);
}

* { margin: 0; padding: 0; box-sizing: border-box; }

/* 字体平滑：根节点，macOS 文字更干净 */
html {
  scroll-padding-top: 24px;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-rendering: optimizeLegibility;
}

body {
  font-family: var(--font-sans);
  font-size: 14px;
  line-height: 1.6;
  background: var(--bg-page);
  color: var(--text-1);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

/* 排版：标题防孤字，正文防末行孤词 */
h1, h2, h3, h4 {
  text-wrap: balance;
}
p, li, figcaption, .subtitle, .stat-label, .foot-tip {
  text-wrap: pretty;
}

/* ===== 焦点可见（键盘导航，勿移除 outline 而不替代）===== */
:focus-visible {
  outline: 2px solid var(--ring);
  outline-offset: 2px;
  border-radius: 4px;
}
.el-button:focus-visible,
.el-input__wrapper:focus-within {
  outline: 2px solid var(--ring);
  outline-offset: 2px;
}

/* ===== 可点击元素：指针 + 精确属性过渡（禁止 transition:all）===== */
button,
.el-button,
.el-menu-item,
.el-table__row,
a,
.copyable,
.demo-chip,
[role='button'] {
  cursor: pointer;
}
button,
.el-button,
.el-menu-item,
a,
.copyable,
.demo-chip {
  transition:
    color var(--dur) var(--ease),
    background-color var(--dur) var(--ease),
    border-color var(--dur) var(--ease),
    box-shadow var(--dur) var(--ease),
    transform var(--dur) var(--ease),
    opacity var(--dur) var(--ease);
}

/* 按下反馈：scale 0.96（不低于 0.95，避免夸张） */
.el-button:not(.is-disabled):not(.is-loading):active,
button:not(:disabled):active {
  transform: scale(0.96);
}

/* 图标按钮光学对齐：图标侧内边距比文字侧少 2px */
.el-button [class*='el-icon'] + span,
.el-button span + [class*='el-icon'] {
  margin-inline-start: 4px;
}

/* 卡片标题：与正文区分层级 */
.el-card__header {
  font-weight: 600;
  color: var(--brand-primary-deep);
  letter-spacing: 0.01em;
}

/* 紧凑「复制」等文字按钮：补足最小点击热区（密集桌面 ≥40px） */
.el-button.is-link,
.el-button.is-text {
  min-height: 40px;
  min-width: 40px;
}

/* ===== 全站统计卡片（Spotlight hover）===== */
.stat {
  position: relative;
  overflow: hidden;
  text-align: center;
  padding: var(--space-5);
  border-radius: 12px;
  border: 1px solid var(--border);
  background: var(--bg-card);
  box-shadow: 0 1px 2px rgba(12, 74, 110, 0.04);
  transition: transform var(--dur) var(--ease), box-shadow var(--dur) var(--ease);
}
/* 跟随光标的局部光斑（21st Spotlight Card 模式，用 CSS 变量定位） */
.stat::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  background: radial-gradient(
    280px circle at var(--mx, 50%) var(--my, 50%),
    rgba(14, 165, 233, 0.16),
    transparent 55%
  );
  opacity: 0;
  transition: opacity var(--dur) var(--ease);
  pointer-events: none;
}
.stat:hover {
  transform: translateY(-2px);
  box-shadow:
    0 1px 2px rgba(12, 74, 110, 0.04),
    0 8px 24px rgba(3, 105, 161, 0.1);
}
.stat:hover::after {
  opacity: 1;
}
.stat-num {
  position: relative;
  z-index: 1;
  font-family: var(--font-mono);
  font-size: 30px;
  font-weight: 600;
  font-variant-numeric: tabular-nums; /* 动态数字不跳动 */
  letter-spacing: -0.5px;
  color: var(--brand-primary);
}
.stat-label {
  position: relative;
  z-index: 1;
  margin-top: var(--space-2);
  color: var(--text-2);
  font-size: 13px;
  letter-spacing: 0.02em;
}

/* ===== 等宽（哈希 / 地址 / 编号）=====
   tabular-nums：数字变化时布局稳定 */
.mono {
  font-family: var(--font-mono);
  font-size: 0.92em;
  font-variant-numeric: tabular-nums;
  letter-spacing: -0.01em;
}

/* 表格中的动态数字统一等宽数字 */
.el-table .cell {
  font-variant-numeric: tabular-nums;
}

/* ===== 列表行视觉锚点 / 状态 pill（UserTable 模式）===== */
.face-cell {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}
.face-cell .name {
  font-weight: 500;
  color: var(--text-1);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.face-cell .sub {
  font-size: 12px;
  color: var(--text-3);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 状态 pill：圆角胶囊 + 描边，元数据感 */
.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border-radius: 999px;
  border: 1px solid var(--border);
  padding: 2px 10px;
  font-size: 12px;
  line-height: 1.6;
  color: var(--text-2);
  background: var(--bg-card);
  white-space: nowrap;
}
.status-pill--ok {
  color: var(--success);
  border-color: #bbf7d0;
  background: #f0fdf4;
}
.status-pill--warn {
  color: var(--warning);
  border-color: #fde68a;
  background: #fffbeb;
}
.status-pill--danger {
  color: var(--danger);
  border-color: #fecaca;
  background: #fef2f2;
}

/* 表格滚动容器：圆角在外层，避免 overflow 破坏圆角 */
.table-shell {
  width: 100%;
  overflow-x: auto;
  border: 1px solid var(--border);
  border-radius: 12px;
  background: var(--bg-card);
}
.table-shell .el-table {
  --el-table-border-color: transparent;
  min-width: 34rem;
}
.table-shell .el-table th.el-table__cell {
  color: var(--text-2);
  font-weight: 500;
  background: var(--bg-muted);
}
.table-shell .el-table .el-table__row {
  transition: background-color 150ms var(--ease);
}
.table-shell .el-table .el-table__row:hover > td.el-table__cell {
  background: #e0f2fe; /* accent hover */
}

/* ===== 品牌页背景（登录 / 注册 / 公开核验）=====
   技术网格 + 柔光晕（源自 background-snippets 视觉）：
   白底 + 6rem×4rem 浅灰网格 + 中上方浅蓝径向光 */
.brand-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-5);
  position: relative;
  overflow: hidden;
  background-color: #ffffff;
  background-image:
    linear-gradient(to right, #f0f0f0 1px, transparent 1px),
    linear-gradient(to bottom, #f0f0f0 1px, transparent 1px);
  background-size: 6rem 4rem;
  background-position: center top;
}
/* 光晕层：纯装饰，不挡交互 */
.brand-page::before {
  content: '';
  position: absolute;
  inset: 0;
  background: radial-gradient(circle 500px at 50% 200px, #c9ebff, transparent);
  pointer-events: none;
  z-index: 0;
}
/* 外层 16px = 内层块 12px + 内边距 ≥4px（区块间距大时按独立表面处理） */
.brand-card {
  position: relative;
  z-index: 1;
  width: 100%;
  padding: var(--space-7) var(--space-6);
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 16px;
  box-shadow:
    0 1px 2px rgba(12, 74, 110, 0.04),
    0 12px 40px rgba(12, 74, 110, 0.08);
}
.brand-card h1 {
  text-align: center;
  font-size: 24px;
  font-weight: 600;
  color: var(--brand-primary-deep);
  margin-bottom: var(--space-2);
  letter-spacing: 0.01em;
}
.brand-card .subtitle {
  text-align: center;
  color: var(--text-2);
  font-size: 13px;
  margin-bottom: var(--space-6);
  line-height: 1.5;
}

/* 卡片内嵌套圆角块：外 16 = 内 12 + 间距 4+ */
.nested-panel {
  border-radius: 12px;
}

/* ===== 图片描边（浅色模式用纯黑低透明，避免染色脏边）===== */
.img-outline {
  outline: 1px solid oklch(0 0 0 / 0.1);
  outline-offset: -1px;
}

/* ===== 成就金徽标 ===== */
.gold-tag {
  color: var(--brand-accent);
  background: var(--brand-accent-soft);
  border-color: #fde68a;
}

/* ===== 减少动效偏好 ===== */
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
    scroll-behavior: auto !important;
  }
  .stat:hover {
    transform: none;
  }
  .el-button:not(.is-disabled):not(.is-loading):active,
  button:not(:disabled):active {
    transform: none;
  }
}
</style>
