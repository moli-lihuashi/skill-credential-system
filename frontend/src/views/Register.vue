<template>
  <div class="brand-page">
    <div class="brand-card" style="max-width: 460px">
      <h1>注册账号</h1>
      <p class="subtitle">注册后自动分配区块链链账户</p>

      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" placeholder="3-64位" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="form.password" type="password" placeholder="至少6位" show-password />
        </el-form-item>
        <el-form-item label="真实姓名" prop="real_name">
          <el-input v-model="form.real_name" />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="form.role" placeholder="选择角色" style="width:100%">
            <el-option label="学生" value="student" />
            <el-option label="院校管理员" value="school_admin" />
            <el-option label="实训企业" value="company" />
            <el-option label="招聘企业" value="employer" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="form.role === 'student'" label="学号" prop="student_no">
          <el-input v-model="form.student_no" placeholder="学生必填" />
        </el-form-item>
        <el-form-item v-if="form.role !== 'student'" label="单位名称" prop="org_name">
          <el-input v-model="form.org_name" placeholder="学校/企业名称" />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" style="width:100%" :loading="loading" @click="handleRegister">
            注 册
          </el-button>
        </el-form-item>
      </el-form>

      <div class="links">
        <router-link to="/login">已有账号？去登录</router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../api'

const router = useRouter()
const formRef = ref()
const loading = ref(false)

const form = reactive({
  username: '', password: '', real_name: '',
  role: 'student', student_no: '', org_name: ''
})

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, min: 6, message: '密码至少6位', trigger: 'blur' }],
  real_name: [{ required: true, message: '请输入真实姓名', trigger: 'blur' }],
  role: [{ required: true, message: '请选择角色', trigger: 'change' }],
  student_no: [{
    validator: (rule, value, callback) => {
      if (form.role === 'student' && !value) {
        callback(new Error('学生角色必须填写学号'))
      } else {
        callback()
      }
    },
    trigger: 'blur'
  }],
  org_name: [{
    validator: (rule, value, callback) => {
      if (form.role !== 'student' && !value) {
        callback(new Error('请填写单位名称'))
      } else {
        callback()
      }
    },
    trigger: 'blur'
  }]
}

async function handleRegister() {
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  loading.value = true
  try {
    const res = await api.post('/auth/register', form)
    await ElMessageBox.alert(
      `链账户地址：${res.data.eth_address}${res.data.pending ? '\n\n该角色需管理员审核激活后才能登录。' : ''}`,
      '注册成功（已自动生成区块链账户）',
      { confirmButtonText: '去登录' }
    )
    router.push('/login')
  } catch (e) {
    // alert 的取消/关闭也会 reject，不能当成注册失败
    if (e !== 'cancel' && e !== 'close') ElMessage.error(e.msg || '注册失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.links {
  text-align: center;
  margin-top: 20px;
}
.links a {
  color: var(--brand-primary);
  text-decoration: none;
  font-size: 13px;
}
.links a:hover { text-decoration: underline; }
</style>
