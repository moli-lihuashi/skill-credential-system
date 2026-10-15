<template>
  <div>
    <el-card>
      <template #header>
        <div style="display:flex;justify-content:space-between;align-items:center">
          <span>个人中心</span>
          <el-tag type="success" effect="plain">链上身份已激活</el-tag>
        </div>
      </template>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="姓名">{{ user.real_name }}</el-descriptions-item>
        <el-descriptions-item label="学号">{{ user.student_no }}</el-descriptions-item>
        <el-descriptions-item label="用户名">{{ user.username }}</el-descriptions-item>
        <el-descriptions-item label="角色">学生</el-descriptions-item>
        <el-descriptions-item label="链账户地址" :span="2">
          <span class="mono">{{ user.eth_address }}</span>
          <el-button link type="primary" size="small" @click="copyAddr">复制</el-button>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <el-card style="margin-top:20px">
      <template #header>
        <div style="display:flex;justify-content:space-between;align-items:center">
          <span>存证概览</span>
          <el-button size="small" text type="primary" @click="$router.push('/student/credentials')">
            查看全部存证 →
          </el-button>
        </div>
      </template>
      <el-row :gutter="16">
        <el-col :span="6">
          <div class="stat">
            <div class="stat-num">{{ display.total }}</div>
            <div class="stat-label">链上存证总数</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="stat">
            <div class="stat-num" style="color:var(--success)">{{ display.certificate }}</div>
            <div class="stat-label">技能证书</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="stat">
            <div class="stat-num" style="color:var(--brand-accent)">{{ display.competition }}</div>
            <div class="stat-label">竞赛获奖</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="stat">
            <div class="stat-num" style="color:var(--brand-secondary)">{{ display.internship + display.project }}</div>
            <div class="stat-label">实训 / 项目</div>
          </div>
        </el-col>
      </el-row>
    </el-card>

    <el-card style="margin-top:20px">
      <template #header>最近存证</template>
      <el-table :data="recent" stripe v-loading="loading">
        <el-table-column prop="credential_no" label="存证编号" width="220">
          <template #default="{ row }"><span class="mono">{{ row.credential_no }}</span></template>
        </el-table-column>
        <el-table-column prop="title" label="成果名称" min-width="140" />
        <el-table-column prop="ach_type" label="类型" width="100">
          <template #default="{ row }"><el-tag size="small">{{ typeLabel(row.ach_type) }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="onchain_time" label="上链时间" width="165" />
        <el-table-column label="操作" width="150">
          <template #default="{ row }">
            <el-button size="small" text type="primary" @click="goDetail">查看全部存证</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && !recent.length" description="暂无存证，录入上链后将在这里展示" />
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import api from '../../api'
import { typeLabel, copyText, countUp, bindSpotlight } from '../../utils'

const router = useRouter()
const user = JSON.parse(localStorage.getItem('user') || '{}')
const stats = reactive({ total: 0, certificate: 0, competition: 0, internship: 0, project: 0 })
// count-up 展示值
const display = reactive({ total: 0, certificate: 0, competition: 0, internship: 0, project: 0 })
const recent = ref([])
const loading = ref(false)
let stopCountUps = []
let unbindSpot = () => {}

async function copyAddr() {
  const ok = await copyText(user.eth_address)
  ElMessage[ok ? 'success' : 'error'](ok ? '链账户地址已复制' : '复制失败')
}

function goDetail() {
  router.push('/student/credentials')
}

function animateStats() {
  stopCountUps.forEach((s) => s && s())
  stopCountUps = Object.keys(stats).map((k) =>
    countUp(display[k] || 0, stats[k], 700, (v) => { display[k] = v })
  )
}

onMounted(async () => {
  // spotlight：对所有 .stat 卡绑定
  const cards = document.querySelectorAll('.stat')
  const unsubs = []
  cards.forEach((el) => unsubs.push(bindSpotlight(el)))
  unbindSpot = () => unsubs.forEach((u) => u())

  loading.value = true
  try {
    const res = await api.get('/credential/my')
    const list = res.data || []
    stats.total = list.length
    list.forEach(item => {
      if (stats[item.ach_type] !== undefined) stats[item.ach_type]++
    })
    recent.value = list.slice(0, 5)
    animateStats()
  } catch (e) { /* ignore */ }
  finally { loading.value = false }
})

onBeforeUnmount(() => {
  stopCountUps.forEach((s) => s && s())
  unbindSpot()
})
</script>

<style scoped>
.mono { font-family: var(--font-mono); }
</style>
