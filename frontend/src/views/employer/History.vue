<template>
  <el-card>
    <template #header>
      <div style="display:flex;justify-content:space-between;align-items:center">
        <span>核验历史记录</span>
        <el-button size="small" @click="loadData">刷新</el-button>
      </div>
    </template>

    <div class="table-shell">
      <el-table :data="paged" stripe v-loading="loadingList" aria-label="核验历史">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="credential_no" label="存证编号" min-width="200">
          <template #default="{ row }"><span class="mono">{{ row.credential_no }}</span></template>
        </el-table-column>
        <el-table-column label="核验方" min-width="140">
          <template #default="{ row }">
            <div class="face-cell">
              <FaceMark :seed="row.employer_name || String(row.id)" :name="row.employer_name" :size="28" />
              <div class="name">{{ row.employer_name }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="input_hash" label="传入哈希" width="170">
          <template #default="{ row }">
            <el-tooltip :content="row.input_hash">
              <span class="mono">{{ shortHex(row.input_hash, 14) }}</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column prop="chain_result" label="结果" width="120">
          <template #default="{ row }">
            <span class="status-pill" :class="row.chain_result === 'pass' ? 'status-pill--ok' : 'status-pill--danger'">
              {{ row.chain_result === 'pass' ? '核验通过' : '核验不通过' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="核验时间" width="170" />
      </el-table>
    </div>
    <el-empty v-if="!loadingList && !list.length" description="暂无核验记录" />

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

async function loadData() {
  loadingList.value = true
  try {
    const res = await api.get('/verify/history')
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
