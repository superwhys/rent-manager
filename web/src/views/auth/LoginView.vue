<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import AuthShell from '@/components/auth/AuthShell.vue'
import EmailCodeField from '@/components/auth/EmailCodeField.vue'
import AppIcon from '@/components/common/AppIcon.vue'
import { useAuthPage } from '@/composables/useAuthPage'
import { useEmailCode } from '@/composables/useEmailCode'
import {
  emailError,
  loginPasswordError,
  verificationCodeError,
} from '@/utils/authValidation'
import { errorMessage } from '@/utils/format'

const {
  auth,
  route,
  router,
  redirect,
  restoring,
  restoreError,
  restore,
  signal,
} = useAuthPage()
const method = ref<'password' | 'code'>('password')
const methods = [
  { value: 'password', label: '密码登录' },
  { value: 'code', label: '验证码登录' },
] as const
const email = ref('')
const password = ref('')
const code = ref('')
const busy = ref(false)
const error = ref('')
const touched = reactive({ email: false, password: false, code: false })
const emailValidation = computed(() =>
  touched.email ? emailError(email.value) : '',
)
const passwordValidation = computed(() =>
  touched.password ? loginPasswordError(password.value) : '',
)
const codeValidation = computed(() =>
  touched.code ? verificationCodeError(code.value) : '',
)
const emailCode = useEmailCode(email)
const { sending, remaining, message, error: sendError } = emailCode
const disabled = computed(() => busy.value || restoring.value)

watch([email, method], () => {
  code.value = ''
  touched.code = false
  error.value = ''
})
watch(method, () => {
  password.value = ''
  touched.password = false
})

async function sendCode() {
  if (disabled.value) return
  touched.email = true
  await emailCode.send()
}

async function submit() {
  if (disabled.value || sending.value) return
  touched.email = true
  touched.password = method.value === 'password'
  touched.code = method.value === 'code'
  if (emailValidation.value || passwordValidation.value || codeValidation.value)
    return
  busy.value = true
  error.value = ''
  try {
    if (method.value === 'password') {
      await auth.loginWithPassword(
        { email: email.value.trim(), password: password.value },
        signal,
      )
    } else {
      await auth.loginWithCode(
        { email: email.value.trim(), code: code.value },
        signal,
      )
    }
    if (signal.aborted) return
    password.value = ''
    code.value = ''
    await router.replace(redirect.value)
  } catch (cause) {
    if (!signal.aborted) error.value = errorMessage(cause)
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <AuthShell
    title="欢迎回到租序"
    description="登录你的账号，继续打理每一处安心。"
    :restoring="restoring"
    :restore-error="restoreError"
    :expired="Boolean(route.query.expired)"
    @restore="restore"
  >
    <div class="login-methods" role="group" aria-label="登录方式">
      <button
        v-for="option in methods"
        :key="option.value"
        type="button"
        :aria-pressed="method === option.value"
        :disabled="disabled || sending"
        @click="method = option.value"
      >
        {{ option.label }}
      </button>
    </div>
    <form class="auth-form" novalidate @submit.prevent="submit">
      <div class="field">
        <label for="login-email">邮箱地址</label>
        <input
          id="login-email"
          v-model="email"
          type="email"
          autocomplete="username"
          autocapitalize="none"
          spellcheck="false"
          placeholder="name@example.com"
          :disabled="disabled"
          :aria-invalid="Boolean(emailValidation)"
          :aria-describedby="emailValidation ? 'login-email-error' : undefined"
          @blur="touched.email = true"
        />
        <small
          v-if="emailValidation"
          id="login-email-error"
          class="auth-error"
          role="alert"
        >
          {{ emailValidation }}
        </small>
      </div>
      <div v-if="method === 'password'" class="field">
        <label for="login-password">密码</label>
        <input
          id="login-password"
          v-model="password"
          type="password"
          autocomplete="current-password"
          placeholder="输入账号密码"
          :disabled="disabled"
          :aria-invalid="Boolean(passwordValidation)"
          :aria-describedby="
            passwordValidation ? 'login-password-error' : undefined
          "
          @blur="touched.password = true"
        />
        <small
          v-if="passwordValidation"
          id="login-password-error"
          class="auth-error"
          role="alert"
        >
          {{ passwordValidation }}
        </small>
      </div>
      <EmailCodeField
        v-else
        v-model="code"
        :disabled="disabled"
        :sending="sending"
        :remaining="remaining"
        :message="message"
        :send-error="sendError"
        :validation-error="codeValidation"
        @send="sendCode"
        @blur="touched.code = true"
      />
      <p v-if="error" class="auth-error" role="alert">{{ error }}</p>
      <el-button
        native-type="submit"
        type="primary"
        :loading="busy"
        :disabled="disabled || sending"
        class="auth-submit"
      >
        登录工作台 <AppIcon name="arrow-right" :size="17" />
      </el-button>
    </form>
    <p class="auth-switch">
      还没有账号？
      <RouterLink
        :to="{ path: '/register', query: { redirect } }"
        :aria-disabled="disabled"
        :tabindex="disabled ? -1 : undefined"
        @click="disabled && $event.preventDefault()"
        >创建账号</RouterLink
      >
    </p>
  </AuthShell>
</template>

<style scoped>
.login-methods {
  display: flex;
  gap: 0.2rem;
  padding: 0.25rem;
  margin-bottom: 1.4rem;
  border-radius: 0.75rem;
  background: #eef2ef;
}
.login-methods button {
  flex: 1;
  min-height: 3.15rem;
  padding: 0.5rem;
  border: 0;
  border-radius: 0.55rem;
  background: transparent;
  color: #697b6e;
  font-size: 0.82rem;
}
.login-methods button[aria-pressed='true'] {
  background: white;
  color: #285e43;
  font-weight: 600;
  box-shadow: 0 1px 4px #254b3712;
}
</style>
