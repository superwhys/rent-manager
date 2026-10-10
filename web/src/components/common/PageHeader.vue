<script setup lang="ts">
defineProps<{ title: string; description?: string }>()
</script>

<template>
  <header class="page-heading">
    <div>
      <h1>{{ title }}</h1>
      <p v-if="description">{{ description }}</p>
    </div>
    <div v-if="$slots.actions" class="page-actions">
      <slot name="actions" />
    </div>
  </header>
</template>

<style scoped>
.page-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1.5rem;
  margin-bottom: 2rem;
}
h1 {
  font-size: 1.9rem;
  font-weight: 650;
  line-height: 1.3;
  letter-spacing: -0.04em;
}
p {
  color: var(--text-secondary);
  margin-top: 0.5rem;
  font-size: 0.875rem;
}
.page-actions {
  display: flex;
  gap: 0.65rem;
  flex-wrap: wrap;
  flex-shrink: 0;
}
@media (max-width: 640px) {
  .page-heading {
    align-items: flex-start;
    flex-direction: column;
    gap: 1rem;
  }
  h1 {
    font-size: 1.65rem;
  }
}
</style>
