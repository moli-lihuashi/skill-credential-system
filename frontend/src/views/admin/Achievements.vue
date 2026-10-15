<template>
  <el-card>
    <template #header>
      <div style="display:flex;justify-content:space-between;align-items:center">
        <span>成果录入（技能证书 / 竞赛获奖）</span>
        <el-button type="primary" @click="openDialog">+ 录入成果</el-button>
      </div>
    </template>

    <el-table :data="paged" stripe v-loading="loadingList">
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="title" label="成果名称" min-width="140" />
      <el-table-column prop="ach_type" label="类型" width="100">
        <template #default="{ row }"><el-tag>{{ typeLabel(row.ach_type) }}</el-tag></template>
      </el-table-column>
      <el-table-column prop="issuer" label="颁发单位" width="140" />
      <el-table-column prop="file_hash" label="文件哈希" width="160">
        <template #default="{ row }">
          <el-tooltip :content="row.file_hash" placement="top">
            <span class="mono">{{ shortHex(row.file_hash, 14) }}</span>
          </el-tooltip>
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.status === 'onchain' ? 'success' : 'warning'">
            {{ row.status === 'onchain' ? '已上链' : '待上链' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="110">
        <template #default="{ row }">
          <el-button
            v-if="row.status !== 'onchain'"
            type="primary" size="small"
            @click="storeOnChain(row)"
          >上链存证</el-button>
          <el-tag v-else type="success" size="small">已完成</el-tag>
        </template>
      </el-table-column>
    </el-table>

    <el-empty v-if="!list.length" description="暂无成果，点击右上角「+ 录入成果」开始" />

    <el-pagination
      v-if="list.length > pageSize"
      class="pager"
      layout="total, prev, pager, next"
      :total="list.length"
      :page-size="pageSize"
      v-model:current-page="page"
    />
  </el-card>

  <!-- 录入对话框 -->
  <el-dialog v-model="showDialog" title="录入成果" width="540px" destroy-on-close>
    <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
      <el-form-item label="学生学号" prop="student_no">
        <el-input
          v-model="form.student_no"
          placeholder="输入学号后自动关联学生"
          @blur="lookupStudent"
        >
          <template #append>
            <el-button @click="lookupStudent" :loading="looking">查询</el-button>
          </template>
        </el-input>
        <!-- 学号查询成功：FaceMark 预览 + 姓名，减少「挂错学生」的焦虑 -->
        <div v-if="studentHint" class="student-hint">
          <FaceMark :seed="form.student_no" :name="studentName" :size="28" />
          <div class="student-meta">
            <div class="student-name">{{ studentName }}</div>
            <div class="student-sub">学号 {{ form.student_no }} · 已关联</div>
          </div>
        </div>
      </el-form-item>

      <el-form-item label="成果名称" prop="title">
        <el-input v-model="form.title" placeholder="如：全国职业院校技能大赛一等奖" />
      </el-form-item>
      <el-form-item label="成果类型" prop="ach_type">
        <el-select v-model="form.ach_type" style="width:100%">
          <el-option label="技能证书" value="certificate" />
          <el-option label="竞赛获奖" value="competition" />
        </el-select>
      </el-form-item>
      <el-form-item label="颁发单位">
        <el-input v-model="form.issuer" placeholder="证书/获奖颁发单位" />
      </el-form-item>
      <el-form-item label="描述">
        <el-input v-model="form.description" type="textarea" :rows="2" />
      </el-form-item>
      <el-form-item label="附件" prop="file_hash">
        <el-upload
          :action="''"
          :auto-upload="false"
          :on-change="handleFileChange"
          :on-exceed="handleFileExceed"
          :limit="1"
          accept=".pdf,.png,.jpg,.jpeg,.doc,.docx"
        >
          <el-button :loading="uploading">选择文件（上传后自动计算哈希）</el-button>
        </el-upload>
        <div v-if="form.file_hash" class="hash-ok">
          <el-icon><CircleCheckFilled /></el-icon>
          SHA-256：<span class="mono">{{ form.file_hash }}</span>
        </div>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="showDialog = false">取消</el-button>
      <el-button type="primary" :loading="loading" @click="handleSubmit">提交</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { UserFilled, CircleCheckFilled } from '@element-plus/icons-vue'
import api from '../../api'
import { typeLabel, shortHex } from '../../utils'
import FaceMark from '../../components/FaceMark.vue'

const list = ref([])
const showDialog = ref(false)
const loading = ref(false)
const loadingList = ref(false)
const looking = ref(false)
const uploading = ref(false)
const formRef = ref()
const studentHint = ref('')
const studentName = ref('')
const page = ref(1)
const pageSize = 8

const form = reactive({
  student_id: '', student_no: '', title: '',
  ach_type: 'certificate', issuer: '', description: '',
  file_hash: '', file_path: ''
})

const rules = {
  student_no: [{ required: true, message: '请填写学号并查询关联学生', trigger: 'blur' }],
  title: [{ required: true, message: '请填写成果名称', trigger: 'blur' }],
  ach_type: [{ required: true, message: '请选择成果类型', trigger: 'change' }],
  file_hash: [{ required: true, message: '请上传附件以计算哈希', trigger: 'change' }]
}

const paged = computed(() => {
  const start = (page.value - 1) * pageSize
  return list.value.slice(start, start + pageSize)
})

function openDialog() {
  form.student_id = ''
  form.student_no = ''
  form.title = ''
  form.ach_type = 'certificate'
  form.issuer = ''
  form.description = ''
  form.file_hash = ''
  form.file_path = ''
  studentHint.value = ''
  studentName.value = ''
  showDialog.value = true
}

async function lookupStudent() {
  if (!form.student_no) return
  looking.value = true
  studentHint.value = ''
  studentName.value = ''
  try {
    const res = await api.get('/user/lookup-student', { params: { student_no: form.student_no } })
    form.student_id = res.data.user_id
    studentName.value = res.data.real_name
    studentHint.value = `已关联：${res.data.real_name}（学号 ${res.data.student_no}）`
    ElMessage.success('学生信息已自动填入')
  } catch (e) {
    form.student_id = ''
    studentName.value = ''
    studentHint.value = ''
    ElMessage.error(e.msg || '未找到该学号对应的学生')
  } finally {
    looking.value = false
  }
}

async function loadData() {
  loadingList.value = true
  try {
    const res = await api.get('/achievement')
    list.value = res.data || []
  } finally {
    loadingList.value = false
  }
}

function handleFileExceed(files) {
  if (files && files[0]) handleFileChange({ raw: files[0] })
}

async function handleFileChange(file) {
  uploading.value = true
  try {
    const fd = new FormData()
    fd.append('file', file.raw)
    const res = await api.post('/upload', fd)
    form.file_hash = res.data.file_hash
    form.file_path = res.data.file_path
    ElMessage.success('文件已上传，哈希已计算')
  } catch (e) {
    ElMessage.error(e.msg || '文件上传失败')
  } finally {
    uploading.value = false
  }
}

async function handleSubmit() {
  await formRef.value.validate()
  if (!form.student_id) {
    ElMessage.warning('请先通过学号查询关联学生')
    return
  }
  loading.value = true
  try {
    await api.post('/achievement', form)
    ElMessage.success('录入成功')
    showDialog.value = false
    loadData()
  } catch (e) {
    ElMessage.error(e.msg || '录入失败')
  } finally {
    loading.value = false
  }
}

async function storeOnChain(row) {
  try {
    await ElMessageBox.confirm(
      `确认将「${row.title}」上链存证？上链后不可修改删除。`,
      '上链确认', { type: 'warning' }
    )
    const res = await api.post('/credential/store', { achievement_id: row.id })
    await ElMessageBox.alert(
      `存证编号：${res.data.credential_no}\n交易哈希：${res.data.chain_tx_hash}\n区块高度：${res.data.chain_block_no}`,
      '上链成功', { type: 'success' }
    )
    loadData()
  } catch (e) {
    if (e !== 'cancel' && e !== 'close') ElMessage.error(e.msg || '上链失败')
  }
}

onMounted(loadData)
</script>

<style scoped>
.mono { font-family: var(--font-mono); }
.student-hint {
  margin-top: 8px;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  background: var(--accent-soft, #e0f2fe);
  border: 1px solid var(--border);
  border-radius: 12px;
  width: 100%;
}
.student-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-1);
}
.student-sub {
  font-size: 12px;
  color: var(--text-2);
}
.hash-ok {
  margin-top: 6px;
  font-size: 12px;
  color: var(--success);
  word-break: break-all;
}
.pager { margin-top: 16px; justify-content: flex-end; }
</style>
