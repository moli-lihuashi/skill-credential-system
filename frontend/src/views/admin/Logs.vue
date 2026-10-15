<template>
  <el-card>
    <template #header>
      <div style="display:flex;justify-content:space-between;align-items:center">
        <span>系统操作日志（全部操作哈希上链留痕）</span>
        <el-button size="small" @click="loadData">刷新</el-button>
      </div>
    </template>

    <div class="table-shell">
      <el-table :data="paged" stripe v-loading="loadingList" aria-label="系统操作日志">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column label="操作人" min-width="130">
          <template #default="{ row }">
            <div class="face-cell">
              <FaceMark :seed="row.username" :name="row.username" :size="28" />
              <div class="name">{{ row.username }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="action" label="操作类型" width="150">
          <template #default="{ row }">
            <el-tag :type="actionType(row.action)" size="small">{{ row.action }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="detail" label="详情" min-width="220" show-overflow-tooltip />
        <el-table-column prop="log_hash" label="日志哈希" width="170">
          <template #default="{ row }">
            <el-tooltip :content="row.log_hash">
              <span class="mono">{{ shortHex(row.log_hash, 16) }}</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column prop="onchain" label="上链" width="90">
          <template #default="{ row }">
            <span class="status-pill" :class="row.onchain ? 'status-pill--ok' : 'status-pill--warn'">
              {{ row.onchain ? '已上链' : '待上链' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="时间" width="165" />
      </el-table>
    </div>

    <el-empty v-if="!loadingList && !list.length" description="暂无操作日志" />

    <el-pagination
      v-if="list.length > pageSize"
      class="pager"
      layout="total, prev, pager, next"
      :total="list.length"
      :page-size="pageSize"
      v-model:current-page="page"
    />
  </el-card>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import api from '../../api'
import { shortHex } from '../../utils'
import FaceMark from '../../components/FaceMark.vue'

const list = ref([])
const loadingList = ref(false)
const page = ref(1)
const pageSize = 10

const paged = computed(() => {
  const start = (page.value - 1) * pageSize
  return list.value.slice(start, start + pageSize)
})

function actionType(action) {
  const a = action || ''
  if (a.includes('FAIL')) return 'danger'
  if (a.includes('STORE') || a.includes('VERIFY_pass')) return 'success'
  return 'info'
}

async function loadData() {
  loadingList.value = true
  try {
    const res = await api.get('/logs')
    list.value = res.data || []
  } finally {
    loadingList.value = false
  }
}

onMounted(loadData)
</script>

<style scoped>
.mono {
  font-family: var(--font-mono);
  font-variant-numeric: tabular-nums;
}
.pager { margin-top: 16px; justify-content: flex-end; }
</style>
