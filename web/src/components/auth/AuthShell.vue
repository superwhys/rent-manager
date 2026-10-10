<script setup lang="ts">
import AppIcon from '@/components/common/AppIcon.vue'

defineProps<{
  title: string
  description: string
  restoring: boolean
  restoreError: string
  expired?: boolean
}>()
defineEmits<{ restore: [] }>()
</script>

<template>
  <main class="auth-page">
    <div class="auth-brand">
      <span><AppIcon name="building" :size="24" /></span>
      栖居 <small>RENT MANAGER</small>
    </div>
    <section class="auth-card" aria-labelledby="auth-title">
      <div class="welcome-icon"><AppIcon name="home" :size="28" /></div>
      <p class="eyebrow">一处工作台，每一份安心</p>
      <h1 id="auth-title">{{ title }}</h1>
      <p class="intro">{{ description }}</p>
      <el-alert
        v-if="expired"
        title="登录已失效，请重新登录。"
        type="warning"
        :closable="false"
        class="auth-notice"
      />
      <div v-if="restoreError" class="restore-notice" role="alert">
        <p>{{ restoreError }}</p>
        <button type="button" :disabled="restoring" @click="$emit('restore')">
          {{ restoring ? '正在验证…' : '重新验证当前会话' }}
        </button>
      </div>
      <p v-if="restoring" class="restore-status" role="status">
        正在验证当前会话…
      </p>
      <slot />
      <p class="session-note">
        <AppIcon name="shield" :size="14" /> 登录状态仅保留在当前浏览器标签页
      </p>
    </section>
    <footer>栖居 · 用心打理，安心生活</footer>
  </main>
</template>

<style scoped>
.auth-page {
  min-height: 100dvh;
  padding: 2rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2rem;
  background: radial-gradient(ellipse at 50% 35%, #e8f0eb 0, #f6f8f7 65%);
}
.auth-brand {
  display: flex;
  align-items: center;
  gap: 0.7rem;
  align-self: flex-start;
  font-size: 1.3rem;
  font-weight: 600;
  color: #3d6650;
}
.auth-brand > span {
  width: 2.5rem;
  height: 2.5rem;
  border-radius: 0.75rem;
  background: #287a60;
  color: white;
  display: grid;
  place-items: center;
}
.auth-brand small {
  font-size: 0.65rem;
  letter-spacing: 0.12em;
  font-weight: 500;
  color: #6f8778;
  margin-left: 0.25rem;
}
.auth-card {
  width: 100%;
  max-width: 29rem;
  margin: auto;
  padding: 2.5rem;
  background: rgba(255, 255, 255, 0.9);
  border: 1px solid #e3ebe5;
  border-radius: 1.4rem;
  box-shadow: 0 1.1rem 4.4rem #35554409;
  backdrop-filter: blur(20px);
}
.welcome-icon {
  background: #edf5f0;
  color: #59866a;
  width: 3.75rem;
  height: 3.75rem;
  border-radius: 1.1rem;
  display: grid;
  place-items: center;
  margin-bottom: 1.5rem;
}
.eyebrow {
  font-size: 0.75rem;
  color: #637f6d;
  letter-spacing: 0.08em;
}
h1 {
  font-size: 1.85rem;
  line-height: 1.25;
  letter-spacing: -0.03em;
  margin: 0.5rem 0 0.65rem;
  font-weight: 600;
}
.intro {
  color: #697c70;
  font-size: 0.82rem;
  line-height: 1.7;
  margin-bottom: 1.8rem;
}
.auth-notice,
.restore-notice {
  margin-bottom: 1rem;
}
.restore-notice,
.restore-status {
  color: #687b6e;
  font-size: 0.8rem;
  line-height: 1.6;
}
.restore-notice button {
  border: 0;
  background: none;
  color: var(--accent);
  padding: 0.75rem 0;
}
.auth-card :deep(.auth-form) {
  display: grid;
  gap: 1rem;
}
.auth-card :deep(.field) {
  margin: 0;
}
.auth-card :deep(.field input) {
  width: 100%;
  min-width: 0;
  min-height: 3.15rem;
}
.auth-card :deep(.field small) {
  color: #6d7f72;
  font-size: 0.74rem;
  line-height: 1.6;
}
.auth-card :deep(.auth-error),
.auth-card :deep(.field .auth-error) {
  color: var(--danger);
  font-size: 0.8rem;
  line-height: 1.6;
  margin: 0;
}
.auth-card :deep(.field input[aria-invalid='true']) {
  border-color: var(--danger);
}
.auth-card :deep(.auth-submit) {
  width: 100%;
  min-height: 3.15rem;
  margin-top: 0.35rem;
}
.auth-card :deep(.auth-submit svg) {
  margin-left: 0.5rem;
}
.auth-card :deep(.auth-switch) {
  color: #687b6e;
  font-size: 0.82rem;
  text-align: center;
  margin: 1.2rem 0 0;
}
.auth-card :deep(.auth-switch a) {
  display: inline-flex;
  align-items: center;
  min-height: 2.75rem;
  margin-left: 0.4rem;
  color: var(--accent);
  font-weight: 600;
}
.auth-card :deep(.auth-switch a[aria-disabled='true']) {
  color: #718476;
  cursor: not-allowed;
}
.session-note {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  justify-content: center;
  margin-top: 1.3rem;
  color: #718476;
  font-size: 0.7rem;
  line-height: 1.6;
}
.session-note svg {
  flex-shrink: 0;
}
footer {
  font-size: 0.75rem;
  color: #718476;
}
@media (max-width: 500px) {
  .auth-page {
    padding: 1.25rem;
    gap: 1.75rem;
  }
  .auth-card {
    padding: 1.5rem;
  }
  .auth-brand small {
    display: none;
  }
}
@media (prefers-reduced-transparency: reduce) {
  .auth-card {
    background: white;
    backdrop-filter: none;
  }
}
@media (prefers-contrast: more) {
  .auth-card {
    background: white;
    border-color: #526b5b;
  }
}
</style>
