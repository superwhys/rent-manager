<script setup lang="ts">
import AppIcon from './AppIcon.vue'
defineProps<{ title: string; description: string; icon: string }>()
</script>

<template>
  <section class="unavailable panel">
    <span class="feature-icon"><AppIcon :name="icon" :size="32" /></span>
    <span class="badge">暂未开放</span>
    <h2>{{ title }}</h2>
    <p>{{ description }}</p>
    <RouterLink to="/dashboard" class="btn"
      >返回经营概览 <AppIcon name="arrow-right" :size="16"
    /></RouterLink>
  </section>
</template>

<style scoped>
.unavailable {
  min-height: 440px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1.25rem;
  text-align: center;
}
.feature-icon {
  color: #85a391;
  background: #edf4ef;
  width: 80px;
  height: 80px;
  border-radius: 24px;
  display: grid;
  place-items: center;
}
h2 {
  font-size: 1.3rem;
}
.unavailable p {
  max-width: 25rem;
  color: var(--text-secondary);
  line-height: 1.8;
}
.unavailable .btn {
  margin-top: 0.75rem;
}
</style>
