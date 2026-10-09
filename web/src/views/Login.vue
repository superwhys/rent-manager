<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { LoginForm } from '../types'
import { useAuth } from '../composables/useAuth'

/**
 * 用户登录页面
 * 提供用户名密码登录功能，支持表单验证
 */

const router = useRouter()
const { login } = useAuth()

// 登录表单数据
const loginForm = reactive<LoginForm>({
  username: '',
  password: ''
})

// 表单加载状态
const loading = ref(false)

// 表单引用
const loginFormRef = ref()

// 表单验证规则
const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度在 3 到 20 个字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度在 6 到 20 个字符', trigger: 'blur' }
  ]
}

/**
 * 处理登录
 */
const handleLogin = async () => {
  try {
    // 验证表单
    await loginFormRef.value.validate()

    loading.value = true

    // 调用登录接口
    const result = await login(loginForm)

    if (result.success) {
      ElMessage.success(result.message)
      // 登录成功，跳转到首页
      router.push('/dashboard')
    } else {
      ElMessage.error(result.message)
    }
  } catch (error) {
    console.error('Login error:', error)
    ElMessage.error('登录过程中发生错误')
  } finally {
    loading.value = false
  }
}

/**
 * 测试账户快速填充
 */
const fillTestAccount = (type: 'landlord' | 'employee' | 'readonly' | 'finance') => {
  if (type === 'landlord') {
    loginForm.username = 'landlord001'
    loginForm.password = '123456'
  } else if (type === 'employee') {
    loginForm.username = 'employee001'
    loginForm.password = '123456'
  } else if (type === 'readonly') {
    loginForm.username = 'employee002'
    loginForm.password = '123456'
  } else if (type === 'finance') {
    loginForm.username = 'employee003'
    loginForm.password = '123456'
  }
}
</script>

<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <div class="logo">🏠</div>
        <h1 class="title">智能房租管理系统</h1>
        <p class="subtitle">欢迎登录</p>
      </div>

      <el-form ref="loginFormRef" :model="loginForm" :rules="rules" class="login-form" @submit.prevent="handleLogin">
        <el-form-item prop="username">
          <el-input v-model="loginForm.username" placeholder="请输入用户名" size="large" prefix-icon="User" clearable />
        </el-form-item>

        <el-form-item prop="password">
          <el-input v-model="loginForm.password" type="password" placeholder="请输入密码" size="large" prefix-icon="Lock"
            show-password @keyup.enter="handleLogin" />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" size="large" :loading="loading" @click="handleLogin" class="login-button">
            {{ loading ? '登录中...' : '登录' }}
          </el-button>
        </el-form-item>
      </el-form>

      <!-- 测试账户 -->
      <div class="test-accounts">
        <div class="test-title">测试账户：</div>
        <div class="test-buttons">
          <el-button size="small" type="info" plain @click="fillTestAccount('landlord')">
            主房东账户
          </el-button>
          <el-button size="small" type="warning" plain @click="fillTestAccount('employee')">
            管理员员工
          </el-button>
          <el-button size="small" type="success" plain @click="fillTestAccount('readonly')">
            只读员工
          </el-button>
          <el-button size="small" type="primary" plain @click="fillTestAccount('finance')">
            财务员工
          </el-button>
        </div>
        <div class="test-hint">
          <strong>主房东：</strong> landlord001 / 123456 （所有权限）<br>
          <strong>管理员：</strong> employee001 / 123456 （管理权限）<br>
          <strong>只读员工：</strong> employee002 / 123456 （查看权限）<br>
          <strong>财务员工：</strong> employee003 / 123456 （财务权限）
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.login-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 20px;
  padding: 40px;
  width: 100%;
  max-width: 400px;
  box-shadow: 0 15px 35px rgba(0, 0, 0, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.18);
}

.login-header {
  text-align: center;
  margin-bottom: 30px;
}

.logo {
  font-size: 60px;
  margin-bottom: 10px;
}

.title {
  font-size: 28px;
  font-weight: bold;
  color: #4f46e5;
  margin: 0 0 8px 0;
}

.subtitle {
  color: #6b7280;
  margin: 0;
  font-size: 16px;
}

.login-form {
  margin-bottom: 20px;
}

.login-button {
  width: 100%;
  height: 50px;
  font-size: 16px;
  font-weight: 600;
}

.test-accounts {
  background: rgba(243, 244, 246, 0.8);
  border-radius: 12px;
  padding: 16px;
  text-align: center;
}

.test-title {
  font-size: 14px;
  color: #6b7280;
  margin-bottom: 12px;
  font-weight: 500;
}

.test-buttons {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
  justify-content: center;
  margin-bottom: 12px;
}

.test-hint {
  font-size: 12px;
  color: #9ca3af;
  line-height: 1.4;
}

/* 响应式设计 */
@media (max-width: 480px) {
  .login-card {
    padding: 30px 20px;
    margin: 10px;
  }

  .title {
    font-size: 24px;
  }

  .logo {
    font-size: 50px;
  }

  .test-buttons {
    grid-template-columns: 1fr;
  }
}

/* Element Plus 样式覆盖 */
:deep(.el-input__wrapper) {
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

:deep(.el-button) {
  border-radius: 12px;
}

:deep(.el-form-item__error) {
  font-size: 12px;
}
</style>