<template>
  <div>
    <el-card>
      <template #header>
        <div style="display:flex;justify-content:space-between;align-items:center">
          <span>实训企业工作台</span>
          <el-tag type="success" effect="plain">联盟链授权节点</el-tag>
        </div>
      </template>
      <el-alert type="info" :closable="false" show-icon
        title="实训企业权限：录入学生岗位实训评价、实训项目成果，并将成果上链存证。" />

      <el-row :gutter="16" style="margin-top:20px">
        <el-col :span="8">
          <div class="stat">
            <div class="stat-num">{{ display.total }}</div>
            <div class="stat-label">录入成果总数</div>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="stat">
            <div class="stat-num" style="color:var(--success)">{{ display.onchain }}</div>
            <div class="stat-label">已上链存证</div>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="stat">
            <div class="stat-num" style="color:var(--brand-accent)">{{ display.pending }}</div>
            <div class="stat-label">待上链</div>
          </div>
        </el-col>
      </el-row>

      <div style="margin-top:24px;display:flex;gap:12px">
        <el-button type="primary" @click="$router.push('/company/achievements')">去录入实训成果</el-button>
        <el-button plain @click="$router.push('/company/achievements')">上链存证</el-button>
      </div>
    </el-card>

    <el-card style="margin-top:20px">
      <template #header>最近录入</template>
      <el-table :data="recent" stripe v-loading="loading">
        <el-table-column prop="title" label="成果名称" min-width="140" />
        <el-table-column prop="ach_type" label="类型" width="110">
          <template #default="{ row }"><el-tag size="small">{{ typeLabel(row.ach_type) }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="issuer" label="评价单位" width="140" />
        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 'onchain' ? 'success' : 'warning'" size="small">
              {{ row.status === 'onchain' ? '已上链' : '待上链' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && !recent.length" description="暂无录入记录" />
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onBeforeUnmount } from 'vue'
import api from '../../api'
import { typeLabel, countUp, bindSpotlight } from '../../utils'

const stats = reactive({ total: 0, onchain: 0, pending: 0 })
const display = reactive({ total: 0, onchain: 0, pending: 0 })
const recent = ref([])
const loading = ref(false)
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

  loading.value = true
  try {
    const res = await api.get('/achievement')
    const list = res.data || []
    stats.total = list.length
    stats.onchain = list.filter(a => a.status === 'onchain').length
    stats.pending = list.filter(a => a.status !== 'onchain').length
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
