import { computed, ref } from 'vue'
import {
  getSession,
  loginEmail,
  loginPassword,
  registerEmail,
  revokeSession,
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
const token = ref(sessionStorage.getItem(storageKey) ?? '')
const expired = ref(false)
let initialized = false
let initialization: Promise<void> | undefined

function clearSession() {
  token.value = ''
  session.value = null
  sessionStorage.removeItem(storageKey)
  configureSession('')
}

configureSession(token.value, () => {
  clearSession()
  expired.value = true
})

function acceptSession(value: Session, accessToken: string) {
  session.value = { account: value.account, expires_at: value.expires_at }
  token.value = accessToken
  sessionStorage.setItem(storageKey, accessToken)
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
    accountName: computed(() => session.value?.account.Username || '我的账号'),
    initialize,
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
