<template>
  <el-card>
    <template #header>
      <div style="display:flex;justify-content:space-between;align-items:center">
        <span>存证凭证检索</span>
        <el-button size="small" @click="loadData">刷新</el-button>
      </div>
    </template>

    <!-- 搜索栏 -->
    <el-form inline>
      <el-form-item label="学生姓名">
        <el-input v-model="query.student_name" placeholder="姓名" clearable @keyup.enter="handleSearch" />
      </el-form-item>
      <el-form-item label="学号">
        <el-input v-model="query.student_no" placeholder="学号" clearable @keyup.enter="handleSearch" />
      </el-form-item>
      <el-form-item label="成果类型">
        <el-select v-model="query.ach_type" clearable placeholder="全部" style="width:140px">
          <el-option label="技能证书" value="certificate" />
          <el-option label="竞赛获奖" value="competition" />
          <el-option label="岗位实训" value="internship" />
          <el-option label="项目实践" value="project" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="handleSearch">检索</el-button>
      </el-form-item>
    </el-form>

    <div class="table-shell">
      <el-table :data="list" stripe v-loading="loadingList" aria-label="存证检索结果">
        <el-table-column prop="credential_no" label="存证编号" width="230">
          <template #default="{ row }">
            <span class="mono">{{ row.credential_no }}</span>
          </template>
        </el-table-column>
        <el-table-column label="学生" min-width="150">
          <template #default="{ row }">
            <div class="face-cell">
              <FaceMark :seed="row.student_no || row.student_name" :name="row.student_name" :size="32" />
              <div style="min-width:0">
                <div class="name">{{ row.student_name }}</div>
                <div class="sub mono">{{ row.student_no }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="成果名称" min-width="140" />
        <el-table-column prop="ach_type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ typeLabel(row.ach_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="chain_block_no" label="区块" width="80" />
        <el-table-column prop="onchain_time" label="上链时间" width="165" />
        <el-table-column prop="verify_count" label="核验" width="70">
          <template #default="{ row }">
            <span class="status-pill" :class="row.verify_count > 0 ? 'status-pill--ok' : ''">
              {{ row.verify_count }}
            </span>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-empty v-if="!loadingList && !list.length" description="未找到符合条件的存证记录" />

    <el-pagination
      class="pager"
      layout="total, prev, pager, next"
      :total="total"
      :page-size="pageSize"
      :current-page="page"
      @current-change="handlePageChange"
    />
  </el-card>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import api from '../../api'
import { typeLabel } from '../../utils'
import FaceMark from '../../components/FaceMark.vue'

const list = ref([])
const total = ref(0)
const loadingList = ref(false)
const query = reactive({ student_name: '', student_no: '', ach_type: '' })
const page = ref(1)
const pageSize = 10

async function loadData() {
  loadingList.value = true
  try {
    const params = { page: page.value, page_size: pageSize }
    Object.entries(query).forEach(([k, v]) => { if (v) params[k] = v })
    const res = await api.get('/credential/search', { params })
    list.value = res.data || []
    total.value = res.total || 0
  } finally {
    loadingList.value = false
  }
}

function handlePageChange(p) {
  page.value = p
  loadData()
}

function handleSearch() {
  page.value = 1
  loadData()
}

onMounted(loadData)
</script>

<style scoped>
.mono { font-family: var(--font-mono); }
.pager { margin-top: 16px; justify-content: flex-end; }
</style>
