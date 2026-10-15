<template>
  <el-card>
    <template #header>
      <div style="display:flex;justify-content:space-between;align-items:center">
        <span>凭证真实性核验</span>
        <el-button size="small" @click="fillDemo">填入演示数据</el-button>
      </div>
    </template>

    <el-alert type="info" :closable="false" show-icon
      title="输入存证凭证编号 + 文件哈希（或学生信息），系统调用智能合约比对链上数据，输出核验结果。" />

    <el-form :model="form" label-width="100px" style="max-width:560px;margin-top:24px">
      <el-form-item label="存证编号">
        <el-input v-model="form.credential_no" placeholder="CC-YYYYMMDDHHmmss-XXXXXX" />
      </el-form-item>

      <el-form-item label="证书文件">
        <el-upload
          :auto-upload="false"
          :show-file-list="false"
          :on-change="handleFileChange"
          :on-exceed="handleFileExceed"
          :limit="1"
          style="width:100%"
        >
          <el-button :loading="hashing">上传文件，自动计算哈希</el-button>
        </el-upload>
        <div v-if="hashHint" class="hash-hint" :class="`hash-hint--${hashHintType}`">{{ hashHint }}</div>
      </el-form-item>

      <el-form-item label="文件哈希">
        <el-input v-model="form.file_hash" placeholder="上传文件自动填入，也可手工粘贴 64 位十六进制">
          <template #append>
            <el-button @click="doCopy(form.file_hash)">复制</el-button>
          </template>
        </el-input>
      </el-form-item>
      <el-form-item label="学生姓名">
        <el-input v-model="form.student_name" placeholder="可选" />
      </el-form-item>
      <el-form-item label="学号">
        <el-input v-model="form.student_no" placeholder="可选" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" size="large" :loading="loading" @click="handleVerify">
          调用合约核验
        </el-button>
        <el-button size="large" @click="simulateTamper">模拟篡改核验</el-button>
      </el-form-item>
    </el-form>

    <el-result
      v-if="result"
      :icon="resultIcon"
      :title="resultTitle"
      :sub-title="msg"
    >
      <template #extra>
        <el-descriptions v-if="result.result === 'pass'" :column="2" border style="max-width:560px;margin:0 auto">
          <el-descriptions-item label="存证编号">
            <span class="mono">{{ result.credential_no }}</span>
            <el-button link type="primary" size="small" @click="doCopy(result.credential_no)">复制</el-button>
          </el-descriptions-item>
          <el-descriptions-item label="学生">{{ result.student_name }}（{{ result.student_no }}）</el-descriptions-item>
          <el-descriptions-item label="成果类型">{{ typeLabel(result.ach_type) }}</el-descriptions-item>
          <el-descriptions-item label="成果名称">{{ result.title }}</el-descriptions-item>
          <el-descriptions-item label="上链时间">{{ result.onchain_time }}</el-descriptions-item>
          <el-descriptions-item label="交易哈希">
            <span class="mono">{{ shortHex(result.chain_tx_hash, 18, 6) }}</span>
            <el-button link type="primary" size="small" @click="doCopy(result.chain_tx_hash)">复制</el-button>
          </el-descriptions-item>
        </el-descriptions>
      </template>
    </el-result>
  </el-card>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { sha256 } from 'js-sha256'
import api from '../../api'
import { typeLabel, shortHex, copyText, tamperHash, DEMO_CREDENTIAL } from '../../utils'

const form = reactive({ credential_no: '', file_hash: '', student_name: '', student_no: '' })
const result = ref(null)
const msg = ref('')
const loading = ref(false)
const hashing = ref(false)
const hashHint = ref('')
const hashHintType = ref('success')

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
  hashHint.value = '已填入演示数据，可直接点击「调用合约核验」'
  ElMessage.success('已填入演示存证数据')
}

function simulateTamper() {
  if (!form.file_hash) {
    form.credential_no = form.credential_no || DEMO_CREDENTIAL.credential_no
    form.file_hash = tamperHash(DEMO_CREDENTIAL.file_hash)
  } else {
    form.file_hash = tamperHash(form.file_hash)
  }
  hashHintType.value = 'warning'
  hashHint.value = '已篡改文件哈希末位，点击核验可看到「不通过」结果'
  ElMessage.warning('已篡改哈希末位，提交核验将演示防篡改能力')
}

function handleFileExceed(files) {
  if (files && files[0]) handleFileChange({ raw: files[0] })
}

async function handleFileChange(file) {
  const raw = file && file.raw
  if (!raw) return
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
    msg.value = res.msg
  } catch (e) {
    result.value = e.data || { result: 'error' }
    msg.value = e.msg || '核验失败（网络或服务异常，请稍后重试）'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.hash-hint {
  margin-top: 6px;
  font-size: 12px;
  word-break: break-all;
}
.hash-hint--success { color: var(--success); }
.hash-hint--warning { color: var(--warning); }
.mono { font-family: var(--font-mono); }
</style>
