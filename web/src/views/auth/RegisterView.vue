<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import AuthShell from '@/components/auth/AuthShell.vue'
import EmailCodeField from '@/components/auth/EmailCodeField.vue'
import AppIcon from '@/components/common/AppIcon.vue'
import { useAuthPage } from '@/composables/useAuthPage'
import { useEmailCode } from '@/composables/useEmailCode'
import {
  emailError,
  registrationPasswordError,
  verificationCodeError,
} from '@/utils/authValidation'
import { errorMessage } from '@/utils/format'

const { auth, router, redirect, restoring, restoreError, restore, signal } =
  useAuthPage()
const email = ref('')
const code = ref('')
const password = ref('')
const confirmation = ref('')
const busy = ref(false)
const error = ref('')
const touched = reactive({
  email: false,
  code: false,
  password: false,
  confirmation: false,
})
const emailValidation = computed(() =>
  touched.email ? emailError(email.value) : '',
)
const codeValidation = computed(() =>
  touched.code ? verificationCodeError(code.value) : '',
)
const passwordValidation = computed(() =>
  touched.password
    ? registrationPasswordError(password.value, password.value)
    : '',
)
const confirmationValidation = computed(() =>
  touched.confirmation && password.value !== confirmation.value
    ? '两次输入的密码不一致'
    : '',
)
const emailCode = useEmailCode(email)
const { sending, remaining, message, error: sendError } = emailCode
const disabled = computed(() => busy.value || restoring.value)

watch(email, () => {
  code.value = ''
  touched.code = false
  error.value = ''
})

async function sendCode() {
  if (disabled.value) return
  touched.email = true
  await emailCode.send()
}

async function submit() {
  if (disabled.value || sending.value) return
  touched.email = true
  touched.code = true
  touched.password = true
  touched.confirmation = true
  if (
    emailValidation.value ||
    codeValidation.value ||
    passwordValidation.value ||
    confirmationValidation.value
  )
    return
  busy.value = true
  error.value = ''
  try {
    await auth.register(
      { email: email.value.trim(), code: code.value, password: password.value },
      signal,
    )
    if (signal.aborted) return
    password.value = ''
    confirmation.value = ''
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
    title="创建你的租序账号"
    description="验证邮箱并设置密码，开启有条理的租赁生活。"
    :restoring="restoring"
    :restore-error="restoreError"
    @restore="restore"
  >
    <form class="auth-form" novalidate @submit.prevent="submit">
      <div class="field">
        <label for="register-email">邮箱地址</label>
        <input
          id="register-email"
          v-model="email"
          type="email"
          autocomplete="username"
          autocapitalize="none"
          spellcheck="false"
          placeholder="name@example.com"
          :disabled="disabled"
          :aria-invalid="Boolean(emailValidation)"
          :aria-describedby="
            emailValidation ? 'register-email-error' : undefined
          "
          @blur="touched.email = true"
        />
        <small
          v-if="emailValidation"
          id="register-email-error"
          class="auth-error"
          role="alert"
        >
          {{ emailValidation }}
        </small>
      </div>
      <EmailCodeField
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
      <div class="field">
        <label for="register-password">设置密码</label>
        <input
          id="register-password"
          v-model="password"
          type="password"
          autocomplete="new-password"
          placeholder="设置 8–128 个字符的密码"
          :disabled="disabled"
          :aria-invalid="Boolean(passwordValidation)"
          aria-describedby="register-password-hint"
          @blur="touched.password = true"
        />
        <small
          id="register-password-hint"
          :class="{ 'auth-error': passwordValidation }"
        >
          {{ passwordValidation || '密码区分大小写，首尾空格也会保留。' }}
        </small>
      </div>
      <div class="field">
        <label for="register-confirmation">确认密码</label>
        <input
          id="register-confirmation"
          v-model="confirmation"
          type="password"
          autocomplete="new-password"
          placeholder="再次输入密码"
          :disabled="disabled"
          :aria-invalid="Boolean(confirmationValidation)"
          :aria-describedby="
            confirmationValidation ? 'register-confirmation-error' : undefined
          "
          @blur="touched.confirmation = true"
        />
        <small
          v-if="confirmationValidation"
          id="register-confirmation-error"
          class="auth-error"
          role="alert"
        >
          {{ confirmationValidation }}
        </small>
      </div>
      <p v-if="error" class="auth-error" role="alert">{{ error }}</p>
      <el-button
        native-type="submit"
        type="primary"
        :loading="busy"
        :disabled="disabled || sending"
        class="auth-submit"
      >
        创建账号并登录 <AppIcon name="arrow-right" :size="17" />
      </el-button>
    </form>
    <p class="auth-switch">
      已有账号？
      <RouterLink
        :to="{ path: '/login', query: { redirect } }"
        :aria-disabled="disabled"
        :tabindex="disabled ? -1 : undefined"
        @click="disabled && $event.preventDefault()"
        >返回登录</RouterLink
      >
    </p>
  </AuthShell>
</template>
