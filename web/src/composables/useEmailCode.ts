import { computed, onUnmounted, ref, watch, type Ref } from 'vue'
import { sendEmailCode } from '@/api/auth'
import { emailError } from '@/utils/authValidation'
import { errorMessage } from '@/utils/format'

export function useEmailCode(email: Ref<string>) {
  const sending = ref(false)
  const error = ref('')
  const now = ref(Date.now())
  const deliveries = ref<
    Record<string, { resendAt: number; expiresAt: number }>
  >({})
  const emailKey = computed(() => email.value.trim().toLowerCase())
  const delivery = computed(() => deliveries.value[emailKey.value])
  const remaining = computed(() =>
    Math.max(
      0,
      Math.ceil(((delivery.value?.resendAt ?? 0) - now.value) / 1000),
    ),
  )
  const message = computed(() => {
    if (!delivery.value) return ''
    return now.value < delivery.value.expiresAt
      ? '验证码已发送，请查看收件箱或垃圾邮件。'
      : '验证码已过期，请重新获取。'
  })
  let timer: ReturnType<typeof setInterval> | undefined
  let disposed = false
  let emailRevision = 0

  watch(
    emailKey,
    () => {
      error.value = ''
      emailRevision += 1
    },
    { flush: 'sync' },
  )

  async function send() {
    if (sending.value || remaining.value > 0) return
    error.value = emailError(email.value)
    if (error.value) return
    const key = emailKey.value
    const revision = emailRevision
    sending.value = true
    try {
      const result = await sendEmailCode(email.value.trim())
      if (disposed) return
      now.value = Date.now()
      deliveries.value[key] = {
        resendAt: now.value + result.resend_after_seconds * 1000,
        expiresAt: now.value + result.expires_in_seconds * 1000,
      }
      if (!timer)
        timer = setInterval(() => {
          now.value = Date.now()
        }, 1000)
    } catch (cause) {
      if (!disposed && revision === emailRevision)
        error.value = errorMessage(cause)
    } finally {
      if (!disposed) sending.value = false
    }
  }

  onUnmounted(() => {
    disposed = true
    clearInterval(timer)
  })

  return { sending, remaining, error, message, send }
}
