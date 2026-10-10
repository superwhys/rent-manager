<script setup lang="ts">
withDefaults(
  defineProps<{
    loading?: boolean
    error?: string
    empty?: boolean
    emptyTitle?: string
    emptyDescription?: string
  }>(),
  {
    emptyTitle: '暂时没有数据',
    emptyDescription: '添加第一条记录，从这里开始。',
  },
)
defineEmits<{ retry: [] }>()
</script>

<template>
  <div v-if="loading" class="state-panel" role="status" aria-live="polite">
    <span class="loading-indicator" aria-hidden="true" />
    <p>正在加载…</p>
  </div>
  <div v-else-if="error" class="state-panel" role="alert">
    <span class="state-symbol">!</span>
    <h3>暂时无法加载</h3>
    <p>{{ error }}</p>
    <el-button @click="$emit('retry')">重新加载</el-button>
  </div>
  <div v-else-if="empty" class="state-panel">
    <span class="state-symbol" aria-hidden="true">—</span>
    <h3>{{ emptyTitle }}</h3>
    <p>{{ emptyDescription }}</p>
    <slot name="empty" />
  </div>
  <slot v-else />
</template>

<style scoped>
.state-panel {
  min-height: 16rem;
  padding: 3rem 1.5rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.85rem;
  text-align: center;
}
.state-panel p {
  color: var(--text-secondary);
  max-width: 30rem;
  overflow-wrap: anywhere;
}
.state-symbol {
  width: 3rem;
  height: 3rem;
  display: grid;
  place-items: center;
  border-radius: 50%;
  background: var(--surface-secondary);
  color: var(--text-secondary);
  font-size: 1.5rem;
}
.loading-indicator {
  width: 1.5rem;
  height: 1.5rem;
  border: 2px solid var(--border);
  border-top-color: var(--accent);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
@media (prefers-reduced-motion: reduce) {
  .loading-indicator {
    animation: none;
  }
}
</style>
