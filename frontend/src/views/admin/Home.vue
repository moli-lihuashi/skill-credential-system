<template>
  <div>
    <el-row :gutter="16">
      <el-col :span="6">
        <div class="stat">
          <div class="stat-num">{{ display.totalCred }}</div>
          <div class="stat-label">链上存证总数</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat">
          <div class="stat-num" style="color:var(--success)">{{ display.totalAch }}</div>
          <div class="stat-label">成果录入总数</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat">
          <div class="stat-num" style="color:var(--brand-accent)">{{ display.pending }}</div>
          <div class="stat-label">待上链</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat">
          <div class="stat-num" style="color:var(--danger)">{{ display.verifyFail }}</div>
          <div class="stat-label">核验不通过</div>
        </div>
      </el-col>
    </el-row>

    <el-card style="margin-top:20px">
      <template #header>快速操作</template>
      <el-space wrap>
        <el-button type="primary" @click="$router.push('/admin/achievements')">录入成果</el-button>
        <el-button type="success" plain @click="$router.push('/admin/credentials')">存证管理</el-button>
        <el-button @click="$router.push('/admin/logs')">查看日志</el-button>
        <el-button text type="primary" @click="$router.push('/verify')">外部核验入口</el-button>
      </el-space>
    </el-card>
  </div>
</template>

<script setup>
import { reactive, onMounted, onBeforeUnmount } from 'vue'
import api from '../../api'
import { countUp, bindSpotlight } from '../../utils'

const stats = reactive({ totalCred: 0, totalAch: 0, pending: 0, verifyFail: 0 })
const display = reactive({ totalCred: 0, totalAch: 0, pending: 0, verifyFail: 0 })
let stopCountUps = []
let unbindSpot = () => {}

function animateStats() {
  stopCountUps.forEach((s) => s && s())
  stopCountUps = Object.keys(stats).map((k) =>
    countUp(display[k] || 0, stats[k], 700, (v) => { display[k] = v })
  )
}

onMounted(async () => {
  const unsubs = []
  document.querySelectorAll('.stat').forEach((el) => unsubs.push(bindSpotlight(el)))
  unbindSpot = () => unsubs.forEach((u) => u())

  try {
    const [credRes, achRes, verifyRes] = await Promise.all([
      api.get('/credential/my'),
      api.get('/achievement'),
      api.get('/verify/history').catch(() => ({ data: [] }))
    ])
    stats.totalCred = (credRes.data || []).length
    const achs = achRes.data || []
    stats.totalAch = achs.length
    stats.pending = achs.filter(a => a.status !== 'onchain').length
    stats.verifyFail = (verifyRes.data || []).filter(v => v.chain_result !== 'pass').length
    animateStats()
  } catch (e) {
    console.error('工作台数据加载失败', e)
  }
})

onBeforeUnmount(() => {
  stopCountUps.forEach((s) => s && s())
  unbindSpot()
})
</script>
