<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import { useSession } from '@/stores/session'
import { safeAuthRedirect } from '@/utils/authValidation'

const route = useRoute()
const router = useRouter()
const { expired, session } = useSession()
const animatePages = ref(true)

function trackNavigationInput(event: Event) {
  animatePages.value = event.type === 'pointerdown'
}

function finishPageLeave(_element: Element, done: () => void) {
  done()
}

onMounted(() => {
  document.addEventListener('pointerdown', trackNavigationInput, true)
  document.addEventListener('keydown', trackNavigationInput, true)
})
onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', trackNavigationInput, true)
  document.removeEventListener('keydown', trackNavigationInput, true)
})
watch(expired, (value) => {
  if (value && route.matched.length > 0)
    void router.replace({
      path: '/login',
      query: {
        redirect: route.meta.public
          ? safeAuthRedirect(route.query.redirect)
          : route.fullPath,
        expired: '1',
      },
    })
})
</script>

<template>
  <el-config-provider :locale="zhCn">
    <RouterView
      v-if="route.meta.public || !session"
      v-slot="{ Component, route: viewRoute }"
    >
      <Transition name="page" :css="animatePages" @leave="finishPageLeave">
        <div v-if="Component" :key="viewRoute.path" class="route-page">
          <component :is="Component" />
        </div>
      </Transition>
    </RouterView>
    <AppLayout v-else>
      <RouterView v-slot="{ Component, route: viewRoute }">
        <Transition name="page" :css="animatePages" @leave="finishPageLeave">
          <div v-if="Component" :key="viewRoute.path" class="route-page">
            <component :is="Component" />
          </div>
        </Transition>
      </RouterView>
    </AppLayout>
  </el-config-provider>
</template>
