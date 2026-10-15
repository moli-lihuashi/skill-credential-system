<template>
  <!--
    行内视觉锚点（源自 21st UserTable / Blobatar 思路）
    - squircle 底：所有人像同宽，姓名左缘对齐
    - 静态无动画：表格是高频扫视区，动效只会干扰
    - alt=""：姓名就在旁边，头像不重复播报
  -->
  <span
    class="face-mark"
    :style="{ background: faceBg }"
    aria-hidden="true"
  >
    {{ initials }}
  </span>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  /** 稳定种子：学号/邮箱/用户 ID，改名后头像不变 */
  seed: { type: String, default: '' },
  name: { type: String, default: '' },
  size: { type: Number, default: 32 }
})

const initials = computed(() => {
  const n = (props.name || props.seed || '?').trim()
  if (!n) return '?'
  // 中文取首字，拉丁取首字母
  if (/[一-鿿]/.test(n)) return n.slice(0, 1)
  const parts = n.split(/[\s@._-]+/).filter(Boolean)
  if (parts.length >= 2) return (parts[0][0] + parts[1][0]).toUpperCase()
  return n.slice(0, 2).toUpperCase()
})

/** 由种子生成稳定色相，信任蓝区间的浅色底 + 深字 */
const faceBg = computed(() => {
  const s = props.seed || props.name || 'x'
  let h = 0
  for (let i = 0; i < s.length; i++) h = (h * 31 + s.charCodeAt(i)) >>> 0
  const hue = 195 + (h % 40) // 195–235：青蓝到蓝
  return `hsl(${hue} 52% 88%)`
})
</script>

<style scoped>
.face-mark {
  width: v-bind(size + 'px');
  height: v-bind(size + 'px');
  border-radius: 10px; /* squircle：与卡片/表格的圆角同族，不用正圆 */
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: calc(v-bind(size + 'px') * 0.38);
  font-weight: 600;
  color: var(--brand-primary-deep);
  letter-spacing: 0.02em;
  flex-shrink: 0;
  user-select: none;
}
</style>
