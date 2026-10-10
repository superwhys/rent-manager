import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useSession } from '@/stores/session'
import { safeAuthRedirect } from '@/utils/authValidation'
import { errorMessage } from '@/utils/format'

export function useAuthPage() {
  const route = useRoute()
  const router = useRouter()
  const auth = useSession()
  const restoring = ref(true)
  const restoreError = ref('')
  const controller = new AbortController()
  const signal = controller.signal
  const redirect = computed(() => safeAuthRedirect(route.query.redirect))

  async function restore() {
    restoring.value = true
    restoreError.value = ''
    try {
      await auth.initialize()
      if (!signal.aborted && auth.isAuthenticated.value)
        await router.replace(redirect.value)
    } catch (cause) {
      if (!signal.aborted) restoreError.value = errorMessage(cause)
    } finally {
      restoring.value = false
    }
  }

  onMounted(restore)
  onUnmounted(() => controller.abort())

  return {
    auth,
    route,
    router,
    redirect,
    restoring,
    restoreError,
    restore,
    signal,
  }
}
