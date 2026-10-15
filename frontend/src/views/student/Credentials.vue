<template>
  <el-card>
    <template #header>
      <div class="card-head">
        <span>我的链上存证凭证</span>
        <div class="head-actions">
          <el-radio-group v-model="viewMode" size="small">
            <el-radio-button value="table">表格</el-radio-button>
            <el-radio-button value="timeline">时间线</el-radio-button>
          </el-radio-group>
          <el-button size="small" @click="loadData">刷新</el-button>
        </div>
      </div>
    </template>

    <!-- 表格视图 -->
    <template v-if="viewMode === 'table'">
      <el-table :data="paged" stripe v-loading="loadingList">
        <el-table-column prop="credential_no" label="存证编号" width="230">
          <template #default="{ row }">
            <span class="copyable" @click="doCopy(row.credential_no)">
              <span class="mono">{{ row.credential_no }}</span>
              <el-icon style="margin-left:4px"><DocumentCopy /></el-icon>
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="成果名称" min-width="150" />
        <el-table-column prop="ach_type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag>{{ typeLabel(row.ach_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="file_hash" label="文件哈希" width="170">
          <template #default="{ row }">
            <el-tooltip :content="row.file_hash">
              <span class="mono">{{ shortHex(row.file_hash, 16) }}</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column prop="onchain_time" label="上链时间" width="165" />
        <el-table-column prop="verify_count" label="核验" width="70" />
        <el-table-column label="操作" width="200">
          <template #default="{ row }">
            <el-button size="small" type="primary" plain @click="showQR(row)">二维码</el-button>
            <el-button size="small" plain @click="shareLink(row)">分享链接</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!list.length" description="暂无存证记录，联系院校管理员或实训企业录入成果" />

      <el-pagination
        v-if="list.length > pageSize"
        class="pager"
        layout="total, prev, pager, next"
        :total="list.length"
        :page-size="pageSize"
        v-model:current-page="page"
      />
    </template>

    <!-- 时间线视图：存证 = 链上锚定事件 -->
    <div v-else v-loading="loadingList" class="timeline-wrap">
      <div v-if="!list.length" class="timeline-empty">
        <el-empty description="暂无链上存证，录入并上链后将在这里形成时间线" />
      </div>

      <ol v-else class="cred-timeline">
        <li v-for="(row, idx) in list" :key="row.credential_no" class="tl-item">
          <div class="tl-rail" aria-hidden="true">
            <span class="tl-node" :class="{ 'tl-node--latest': idx === 0 }" />
            <span v-if="idx < list.length - 1" class="tl-line" />
          </div>

          <div class="tl-card">
            <div class="tl-top">
              <div class="tl-title">{{ row.title }}</div>
              <el-tag size="small" effect="plain">{{ typeLabel(row.ach_type) }}</el-tag>
            </div>

            <div class="tl-meta">
              <span class="tl-time">{{ row.onchain_time }}</span>
              <span class="tl-dot">·</span>
              <span class="mono">{{ row.credential_no }}</span>
            </div>

            <div class="tl-hash mono">{{ shortHex(row.file_hash, 28) }}</div>

            <div class="tl-foot">
              <span class="tl-badge">
                <span class="tl-pulse" /> 已上链锚定
              </span>
              <span class="tl-verify">核验 {{ row.verify_count }} 次</span>
              <div class="tl-actions">
                <el-button size="small" text type="primary" @click="showQR(row)">二维码</el-button>
                <el-button size="small" text @click="shareLink(row)">分享</el-button>
                <el-button size="small" text @click="doCopy(row.credential_no)">复制编号</el-button>
              </div>
            </div>
          </div>
        </li>
      </ol>
    </div>

    <!-- 二维码弹窗 -->
    <el-dialog v-model="qrVisible" title="存证凭证二维码" width="420px">
      <div style="text-align:center">
        <img
          v-if="qrDataUrl"
          :src="qrDataUrl"
          alt="存证凭证核验二维码"
          class="qr-img img-outline"
        />
        <div v-else class="qr-placeholder">二维码生成中…</div>
        <div style="margin-top:16px;font-size:13px;color:#666;text-align:left">
          <p><b>存证编号：</b><span class="mono">{{ currentCred?.credential_no }}</span></p>
          <p><b>成果名称：</b>{{ currentCred?.title }}</p>
          <p><b>文件哈希：</b><span class="mono">{{ shortHex(currentCred?.file_hash, 24) }}</span></p>
        </div>
        <el-divider />
        <p style="font-size:12px;color:#999">企业扫码后上传收到的证书文件，即可自动比对链上哈希完成核验</p>
        <div style="margin-top:12px;display:flex;gap:10px;justify-content:center">
          <el-button type="primary" plain size="small" @click="doCopy(currentCred?.credential_no)">
            复制存证编号
          </el-button>
          <el-button plain size="small" @click="shareLink(currentCred)">复制核验链接</el-button>
        </div>
      </div>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { DocumentCopy } from '@element-plus/icons-vue'
import QRCode from 'qrcode'
import api from '../../api'
import { typeLabel, shortHex, copyText } from '../../utils'

const list = ref([])
const viewMode = ref('timeline')
const qrVisible = ref(false)
const currentCred = ref(null)
const qrDataUrl = ref('')
const loadingList = ref(false)
const page = ref(1)
const pageSize = 8

const paged = computed(() => {
  const start = (page.value - 1) * pageSize
  return list.value.slice(start, start + pageSize)
})

async function loadData() {
  loadingList.value = true
  try {
    const res = await api.get('/credential/my')
    list.value = res.data || []
  } finally {
    loadingList.value = false
  }
}

async function doCopy(no) {
  if (!no) return
  const ok = await copyText(no)
  ElMessage[ok ? 'success' : 'error'](ok ? '已复制到剪贴板' : '复制失败')
}

function buildVerifyUrl(row) {
  const query = new URLSearchParams({
    credential_no: row.credential_no || '',
    student_no: row.student_no || ''
  }).toString()
  return `${window.location.origin}/verify?${query}`
}

async function shareLink(row) {
  if (!row) return
  const ok = await copyText(buildVerifyUrl(row))
  ElMessage[ok ? 'success' : 'error'](ok ? '核验链接已复制，可发给招聘企业' : '复制失败')
}

async function showQR(row) {
  currentCred.value = row
  qrDataUrl.value = ''
  qrVisible.value = true
  try {
    qrDataUrl.value = await QRCode.toDataURL(buildVerifyUrl(row), {
      width: 220,
      margin: 1,
      errorCorrectionLevel: 'M',
      color: { dark: '#0C4A6E', light: '#FFFFFF' }
    })
  } catch (e) {
    ElMessage.error('二维码生成失败：' + (e && e.message ? e.message : e))
  }
}

onMounted(loadData)
</script>

<style scoped>
.card-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.head-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.mono {
  font-family: var(--font-mono);
  font-variant-numeric: tabular-nums;
}
.copyable {
  cursor: pointer;
  color: var(--brand-primary);
}
.copyable:hover { text-decoration: underline; }
.pager { margin-top: 16px; justify-content: flex-end; }

/* ===== 时间线 ===== */
.timeline-wrap {
  min-height: 200px;
  padding: 8px 4px 4px;
}
.cred-timeline {
  list-style: none;
  margin: 0;
  padding: 0;
}
.tl-item {
  display: grid;
  grid-template-columns: 28px 1fr;
  gap: 12px;
}
.tl-rail {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 18px;
}
.tl-node {
  width: 12px;
  height: 12px;
  border-radius: 999px;
  background: var(--brand-primary);
  box-shadow: 0 0 0 3px rgba(3, 105, 161, 0.18);
  flex-shrink: 0;
  z-index: 1;
}
.tl-node--latest {
  background: var(--brand-accent);
  box-shadow: 0 0 0 4px rgba(161, 98, 7, 0.2);
}
.tl-line {
  width: 2px;
  flex: 1;
  margin-top: 4px;
  background: linear-gradient(180deg, var(--border), transparent);
  border-radius: 2px;
}

.tl-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 16px 18px;
  margin-bottom: 14px;
  transition:
    box-shadow var(--dur, 220ms) var(--ease, cubic-bezier(0.22, 1, 0.36, 1)),
    border-color var(--dur, 220ms) var(--ease, cubic-bezier(0.22, 1, 0.36, 1));
}
.tl-card:hover {
  border-color: #7dd3fc;
  box-shadow:
    0 1px 2px rgba(12, 74, 110, 0.04),
    0 8px 24px rgba(3, 105, 161, 0.08);
}

.tl-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.tl-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-1);
}
.tl-meta {
  margin-top: 8px;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--text-2);
}
.tl-dot { opacity: 0.5; }
.tl-hash {
  margin-top: 8px;
  font-size: 12px;
  color: var(--text-3);
  word-break: break-all;
}

.tl-foot {
  margin-top: 12px;
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.tl-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--success);
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  padding: 2px 10px;
  border-radius: 999px;
}
.tl-pulse {
  width: 6px;
  height: 6px;
  border-radius: 999px;
  background: var(--success);
}
.tl-verify {
  font-size: 12px;
  color: var(--text-3);
}
.tl-actions {
  margin-left: auto;
  display: flex;
  gap: 4px;
}

/* 二维码 */
.qr-img {
  width: 220px;
  height: 220px;
  padding: 8px;
  background: #fff;
  border-radius: 12px;
  display: block;
  margin: 0 auto;
}
.qr-placeholder {
  width: 220px;
  height: 220px;
  margin: 0 auto;
  line-height: 220px;
  color: var(--text-3);
  border: 1px dashed var(--border);
  border-radius: 12px;
  text-align: center;
}
</style>
