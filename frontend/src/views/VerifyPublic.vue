<template>
  <div class="brand-page">
    <div class="brand-card verify-card">
      <h1>成果凭证核验</h1>
      <p class="subtitle">输入存证凭证编号与文件哈希，核验成果真实性</p>

      <div class="demo-bar">
        <el-button size="small" plain type="primary" @click="fillDemo">填入演示数据</el-button>
        <el-button size="small" plain type="warning" @click="simulateTamper">一键篡改演示</el-button>
        <el-button size="small" text @click="copyShareLink">复制核验链接</el-button>
      </div>

      <el-form :model="form" label-width="100px">
        <el-form-item label="存证编号">
          <el-input v-model="form.credential_no" placeholder="如 CC-20240601120000-A1B2C3">
            <template #append>
              <el-button @click="doCopy(form.credential_no)">复制</el-button>
            </template>
          </el-input>
        </el-form-item>

        <!-- 企业拿到证书文件后直接上传，系统自动计算 SHA-256 并填入下方 -->
        <el-form-item label="证书文件">
          <el-upload
            :auto-upload="false"
            :show-file-list="false"
            :on-change="handleFileChange"
            :on-exceed="handleFileExceed"
            :limit="1"
            style="width:100%"
          >
            <el-button :loading="hashing" style="width:100%">上传文件，自动计算哈希</el-button>
          </el-upload>
          <div v-if="hashHint" class="hash-hint" :class="`hash-hint--${hashHintType}`">{{ hashHint }}</div>
        </el-form-item>

        <el-form-item label="文件哈希">
          <el-input
            v-model="form.file_hash"
            type="textarea"
            :rows="2"
            placeholder="上传文件自动填入，也可手工粘贴 64 位十六进制"
          />
        </el-form-item>
        <el-form-item label="学生姓名">
          <el-input v-model="form.student_name" placeholder="可选：辅助校验" />
        </el-form-item>
        <el-form-item label="学号">
          <el-input v-model="form.student_no" placeholder="可选：辅助校验" />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" size="large" style="width:100%" :loading="loading" @click="handleVerify">
            开始核验
          </el-button>
        </el-form-item>
      </el-form>

      <!-- 核验结果：pass / fail(篡改) / error(服务异常) 三态区分，避免把网络故障误判为篡改 -->
      <el-result
        v-if="result"
        :icon="resultIcon"
        :title="resultTitle"
        :sub-title="resultMsg"
      >
        <template #extra>
          <el-descriptions v-if="result.result === 'pass'" :column="1" border class="result-table">
            <el-descriptions-item label="存证编号">
              <span class="mono">{{ result.credential_no }}</span>
              <el-button link type="primary" size="small" @click="doCopy(result.credential_no)">复制</el-button>
            </el-descriptions-item>
            <el-descriptions-item label="学生姓名">{{ result.student_name }}</el-descriptions-item>
            <el-descriptions-item label="学号">{{ result.student_no }}</el-descriptions-item>
            <el-descriptions-item label="成果类型">{{ typeLabel(result.ach_type) }}</el-descriptions-item>
            <el-descriptions-item label="成果名称">{{ result.title }}</el-descriptions-item>
            <el-descriptions-item label="上链时间">{{ result.onchain_time }}</el-descriptions-item>
            <el-descriptions-item label="交易哈希">
              <span class="mono">{{ shortHex(result.chain_tx_hash, 18, 6) }}</span>
              <el-button link type="primary" size="small" @click="doCopy(result.chain_tx_hash)">复制</el-button>
            </el-descriptions-item>
          </el-descriptions>
          <div class="result-actions">
            <el-button @click="reset">继续核验</el-button>
            <el-button type="primary" plain @click="printResult">打印 / 导出 PDF</el-button>
          </div>
        </template>
      </el-result>

      <p class="foot-tip">本页面无需登录 · 核验过程调用联盟链智能合约比对链上哈希</p>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { sha256 } from 'js-sha256'
import api from '../api'
import { typeLabel, shortHex, copyText, tamperHash, DEMO_CREDENTIAL } from '../utils'

const route = useRoute()

const form = reactive({
  credential_no: '', file_hash: '', student_name: '', student_no: ''
})

// 创意功能：扫码/链接跳转后自动预填核验参数
onMounted(() => {
  if (route.query.credential_no) form.credential_no = route.query.credential_no
  if (route.query.file_hash) form.file_hash = route.query.file_hash
  if (route.query.student_no) form.student_no = route.query.student_no
  if (route.query.credential_no) {
    ElMessage.info('已自动预填存证信息，请确认后点击核验')
  }
})
const result = ref(null)
const resultMsg = ref('')
const loading = ref(false)
const hashing = ref(false)
const hashHint = ref('')
const hashHintType = ref('success') // success | warning

const resultIcon = computed(() => {
  const r = result.value?.result
  if (r === 'pass') return 'success'
  if (r === 'error') return 'warning'
  return 'error'
})
const resultTitle = computed(() => {
  const r = result.value?.result
  if (r === 'pass') return '核验通过（数据未篡改）'
  if (r === 'error') return '核验服务暂不可用'
  return '核验不通过（信息存在篡改风险）'
})

async function doCopy(text) {
  const ok = await copyText(text)
  ElMessage[ok ? 'success' : 'error'](ok ? '已复制到剪贴板' : '复制失败')
}

function fillDemo() {
  form.credential_no = DEMO_CREDENTIAL.credential_no
  form.file_hash = DEMO_CREDENTIAL.file_hash
  form.student_name = DEMO_CREDENTIAL.student_name
  form.student_no = DEMO_CREDENTIAL.student_no
  hashHintType.value = 'success'
  hashHint.value = '已填入演示数据，点击「开始核验」应显示通过'
  ElMessage.success('已填入演示存证数据')
}

function simulateTamper() {
  // 无论是否已有哈希，篡改末位，演示「哈希不一致 = 篡改」
  const base = form.file_hash || DEMO_CREDENTIAL.file_hash
  form.credential_no = form.credential_no || DEMO_CREDENTIAL.credential_no
  form.file_hash = tamperHash(base)
  // 篡改属警告语义，不可用成功绿
  hashHintType.value = 'warning'
  hashHint.value = '已篡改文件哈希末位，点击「开始核验」应显示不通过'
  ElMessage.warning('已篡改哈希末位，开始核验即可演示防篡改')
}

async function copyShareLink() {
  const query = new URLSearchParams()
  if (form.credential_no) query.set('credential_no', form.credential_no)
  if (form.student_no) query.set('student_no', form.student_no)
  const url = `${window.location.origin}/verify${query.toString() ? '?' + query.toString() : ''}`
  const ok = await copyText(url)
  ElMessage[ok ? 'success' : 'error'](ok ? '核验链接已复制，可发给招聘企业' : '复制失败')
}

function printResult() {
  window.print()
}

// 超出 limit=1 时用新文件覆盖旧文件，避免第二次选文件被静默丢弃
function handleFileExceed(files) {
  if (files && files[0]) handleFileChange({ raw: files[0] })
}

// 企业上传收到的证书文件，浏览器端直接计算 SHA-256 填入表单
// 说明：公开核验页无需登录，不能复用需要鉴权的 /api/upload，因此在前端计算
// 计算逻辑与后端保持一致（对文件原始字节做 SHA-256，输出 64 位小写十六进制）
async function handleFileChange(file) {
  const raw = file && file.raw
  if (!raw) return
  // 与后端上传限制保持一致（20MB），超限直接提示，避免主线程假死
  if (raw.size > 20 * 1024 * 1024) {
    ElMessage.error('文件超过 20MB 限制，请压缩后再上传')
    return
  }
  hashing.value = true
  hashHint.value = ''
  try {
    const buf = await raw.arrayBuffer()
    const hex = sha256(new Uint8Array(buf))
    form.file_hash = hex
    hashHintType.value = 'success'
    hashHint.value = `已计算 SHA-256：${hex.slice(0, 20)}…（${raw.name}，${(raw.size / 1024).toFixed(1)} KB）`
    ElMessage.success('文件哈希已自动填入')
  } catch (e) {
    ElMessage.error('文件读取失败：' + (e && e.message ? e.message : e))
  } finally {
    hashing.value = false
  }
}

async function handleVerify() {
  if (!form.credential_no || !form.file_hash) {
    ElMessage.warning('请填写存证编号和文件哈希')
    return
  }
  loading.value = true
  result.value = null
  try {
    const res = await api.post('/verify', form)
    result.value = res.data
    resultMsg.value = res.msg
  } catch (e) {
    // 网络/服务异常归为 error，不能与「篡改」混为一谈
    result.value = e.data || { result: 'error' }
    resultMsg.value = e.msg || '核验失败（网络或服务异常，请稍后重试）'
  } finally {
    loading.value = false
  }
}

function reset() {
  result.value = null
}
</script>

<style scoped>
.verify-card { max-width: 580px; }

/* 演示工具条：成就金区，与「信任蓝」主操作区分（独立表面，圆角 12） */
.demo-bar {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  justify-content: center;
  margin-bottom: var(--space-5);
  padding: var(--space-3);
  background: var(--brand-accent-soft);
  border-radius: 12px;
  border: 1px dashed #fde68a;
}

.hash-hint {
  margin-top: 6px;
  font-size: 12px;
  word-break: break-all;
}
.hash-hint--success { color: var(--success); }
.hash-hint--warning { color: var(--warning); }

.result-table {
  max-width: 420px;
  margin: 0 auto;
}
.result-actions {
  margin-top: var(--space-4);
  display: flex;
  gap: var(--space-3);
  justify-content: center;
}
.foot-tip {
  text-align: center;
  font-size: 12px;
  color: var(--text-3);
  margin-top: var(--space-5);
  letter-spacing: 0.02em;
}
.mono { font-family: var(--font-mono); }

@media print {
  .demo-bar, .result-actions, .foot-tip { display: none !important; }
}
</style>
