import { computed, ref } from 'vue'
import {
  getSession,
  loginEmail,
  loginPassword,
  registerEmail,
  revokeSession,
  updateAccount,
} from '@/api/auth'
import { ApiError, configureSession } from '@/api/request'
import type {
  EmailLoginRequest,
  PasswordLoginRequest,
  RegisterRequest,
  Session,
} from '@/types/auth'

const storageKey = 'rent-manager.session'
const session = ref<Session | null>(null)
const token = ref(localStorage.getItem(storageKey) ?? '')
const expired = ref(false)
let initialized = false
let initialization: Promise<void> | undefined

function clearSession() {
  // 其他标签页可能已登录新会话，旧请求的失败不能清除它。
  if (localStorage.getItem(storageKey) === token.value)
    localStorage.removeItem(storageKey)
  token.value = ''
  session.value = null
  configureSession('')
}

function syncSession(event?: StorageEvent) {
  if (
    event &&
    (event.storageArea !== localStorage ||
      (event.key !== storageKey && event.key !== null))
  ) return
  // 重新初始化页面和业务状态，避免切换账号后仍显示旧账号的数据。
  if ((localStorage.getItem(storageKey) ?? '') !== token.value)
    window.location.reload()
}

configureSession(token.value, () => {
  clearSession()
  expired.value = true
})

function acceptSession(value: Session, accessToken: string) {
  session.value = {
    account: value.account,
    expires_at: value.expires_at,
    is_super_admin: value.is_super_admin,
  }
  token.value = accessToken
  localStorage.setItem(storageKey, accessToken)
  configureSession(accessToken)
  initialized = true
  expired.value = false
}

async function initialize() {
  if (initialized) return
  if (!token.value) {
    initialized = true
    return
  }
  if (!initialization) {
    initialization = getSession()
      .then((value) => {
        session.value = value
        initialized = true
      })
      .catch((error) => {
        if (
          error instanceof ApiError &&
          (error.status === 401 || [400002, 400004].includes(error.code))
        ) {
          clearSession()
          initialized = true
          return
        }
        throw error
      })
      .finally(() => {
        initialization = undefined
      })
  }
  return initialization
}

export function useSession() {
  return {
    session,
    expired,
    isAuthenticated: computed(() => Boolean(session.value)),
    isSuperAdmin: computed(() => session.value?.is_super_admin === true),
    accountName: computed(() => session.value?.account.Username || '我的账号'),
    initialize,
    syncSession,
    async updateAccountName(username: string) {
      const currentToken = token.value
      const account = await updateAccount({ username })
      if (
        session.value &&
        token.value === currentToken &&
        session.value.account.ID === account.ID
      ) {
        session.value = { ...session.value, account }
      }
    },
    async loginWithPassword(body: PasswordLoginRequest, signal?: AbortSignal) {
      const result = await loginPassword(body, signal)
      signal?.throwIfAborted()
      acceptSession(result, result.access_token)
    },
    async loginWithCode(body: EmailLoginRequest, signal?: AbortSignal) {
      const result = await loginEmail(body, signal)
      signal?.throwIfAborted()
      acceptSession(result, result.access_token)
    },
    async register(body: RegisterRequest, signal?: AbortSignal) {
      const result = await registerEmail(body, signal)
      signal?.throwIfAborted()
      acceptSession(result, result.access_token)
    },
    async logout() {
      await revokeSession()
      clearSession()
    },
  }
}
