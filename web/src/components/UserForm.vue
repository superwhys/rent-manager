<script setup lang="ts">
import { reactive, ref, computed } from 'vue'
import type { CreateUserForm, User, UserRole, PropertyOption } from '../types'
import { ElMessage } from 'element-plus'

/**
 * 用户表单组件
 * 用于创建和编辑用户信息，支持权限分配
 */

// Props定义
interface Props {
  visible: boolean
  user?: User | null
  propertyOptions: PropertyOption[]
  isEdit?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  user: null,
  isEdit: false
})

// Emits定义
interface Emits {
  'update:visible': [value: boolean]
  'confirm': [userData: CreateUserForm, isEdit: boolean, editUserId?: string]
}

const emit = defineEmits<Emits>()

// 表单数据
const formData = reactive<CreateUserForm>({
  username: '',
  email: '',
  phone: '',
  realName: '',
  password: '',
  confirmPassword: '',
  role: 'employee',
  managedProperties: []
})

// 表单引用
const formRef = ref()

// 加载状态
const loading = ref(false)

// 角色选项
const roleOptions = [
  { label: '主房东', value: 'landlord' },
  { label: '员工', value: 'employee' }
]

// 表单验证规则
const rules = computed(() => ({
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度在 3 到 20 个字符', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9_]+$/, message: '用户名只能包含字母、数字和下划线', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ],
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号', trigger: 'blur' }
  ],
  realName: [
    { required: true, message: '请输入真实姓名', trigger: 'blur' },
    { min: 2, max: 10, message: '姓名长度在 2 到 10 个字符', trigger: 'blur' }
  ],
  password: props.isEdit ? [] : [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度在 6 到 20 个字符', trigger: 'blur' }
  ],
  confirmPassword: props.isEdit ? [] : [
    { required: true, message: '请确认密码', trigger: 'blur' },
    {
      validator: (rule: any, value: string, callback: Function) => {
        if (value !== formData.password) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ],
  role: [
    { required: true, message: '请选择用户角色', trigger: 'change' }
  ],
  managedProperties: [
    {
      validator: (rule: any, value: number[], callback: Function) => {
        if (formData.role === 'employee' && (!value || value.length === 0)) {
          callback(new Error('员工必须分配至少一个物业权限'))
        } else {
          callback()
        }
      },
      trigger: 'change'
    }
  ]
}))

// 计算属性
const dialogTitle = computed(() => props.isEdit ? '编辑用户' : '创建用户')
const submitButtonText = computed(() => props.isEdit ? '保存' : '创建')

// 是否显示物业权限选择
const showPropertySelection = computed(() => formData.role === 'employee')

// 监听弹窗显示状态，初始化表单数据
const handleDialogOpen = () => {
  if (props.isEdit && props.user) {
    // 编辑模式，填充现有数据
    Object.assign(formData, {
      username: props.user.username,
      email: props.user.email,
      phone: props.user.phone,
      realName: props.user.realName,
      password: '',
      confirmPassword: '',
      role: props.user.role,
      managedProperties: [...props.user.managedProperties]
    })
  } else {
    // 创建模式，重置表单
    resetForm()
  }
}

/**
 * 重置表单
 */
const resetForm = () => {
  Object.assign(formData, {
    username: '',
    email: '',
    phone: '',
    realName: '',
    password: '',
    confirmPassword: '',
    role: 'employee',
    managedProperties: []
  })
  formRef.value?.clearValidate()
}

/**
 * 角色变化处理
 */
const handleRoleChange = (role: UserRole) => {
  if (role === 'landlord') {
    // 主房东自动拥有所有物业权限
    formData.managedProperties = props.propertyOptions.map(p => p.value)
  } else {
    // 员工需要手动分配权限
    formData.managedProperties = []
  }
}

/**
 * 提交表单
 */
const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    
    loading.value = true
    
    // 发送数据给父组件
    emit('confirm', { ...formData }, props.isEdit, props.user?.id)
    
    ElMessage.success(props.isEdit ? '用户信息更新成功' : '用户创建成功')
    handleClose()
  } catch (error) {
    console.error('Form validation failed:', error)
  } finally {
    loading.value = false
  }
}

/**
 * 关闭弹窗
 */
const handleClose = () => {
  resetForm()
  emit('update:visible', false)
}
</script>

<template>
  <el-dialog
    :model-value="visible"
    :title="dialogTitle"
    width="600px"
    :before-close="handleClose"
    @open="handleDialogOpen"
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="rules"
      label-width="100px"
      label-position="left"
    >
      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="用户名" prop="username">
            <el-input
              v-model="formData.username"
              placeholder="请输入用户名"
              :disabled="isEdit"
            />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="真实姓名" prop="realName">
            <el-input
              v-model="formData.realName"
              placeholder="请输入真实姓名"
            />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="邮箱" prop="email">
            <el-input
              v-model="formData.email"
              placeholder="请输入邮箱"
              type="email"
            />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="手机号" prop="phone">
            <el-input
              v-model="formData.phone"
              placeholder="请输入手机号"
            />
          </el-form-item>
        </el-col>
      </el-row>

      <el-form-item label="用户角色" prop="role">
        <el-select
          v-model="formData.role"
          placeholder="请选择用户角色"
          style="width: 100%"
          @change="handleRoleChange"
        >
          <el-option
            v-for="option in roleOptions"
            :key="option.value"
            :label="option.label"
            :value="option.value"
          />
        </el-select>
      </el-form-item>

      <el-form-item
        v-if="showPropertySelection"
        label="物业权限"
        prop="managedProperties"
      >
        <el-select
          v-model="formData.managedProperties"
          placeholder="请选择可管理的物业"
          multiple
          style="width: 100%"
        >
          <el-option
            v-for="property in propertyOptions"
            :key="property.value"
            :label="property.label"
            :value="property.value"
          />
        </el-select>
      </el-form-item>

      <div v-if="!isEdit">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="密码" prop="password">
              <el-input
                v-model="formData.password"
                type="password"
                placeholder="请输入密码"
                show-password
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="确认密码" prop="confirmPassword">
              <el-input
                v-model="formData.confirmPassword"
                type="password"
                placeholder="请确认密码"
                show-password
              />
            </el-form-item>
          </el-col>
        </el-row>
      </div>
    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button
          type="primary"
          :loading="loading"
          @click="handleSubmit"
        >
          {{ submitButtonText }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<style scoped>
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

:deep(.el-dialog__body) {
  padding: 20px 20px 10px 20px;
}

:deep(.el-form-item__label) {
  font-weight: 500;
}
</style>