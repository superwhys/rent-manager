<script setup lang="ts">
import { watch } from 'vue'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import { useSession } from '@/stores/session'
import { safeAuthRedirect } from '@/utils/authValidation'

const route = useRoute()
const router = useRouter()
const { expired, session } = useSession()
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
    <RouterView v-if="route.meta.public || !session" />
    <AppLayout v-else><RouterView /></AppLayout>
  </el-config-provider>
</template>
