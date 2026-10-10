<script setup lang="ts">
import { computed } from 'vue'
import AppIcon from '@/components/common/AppIcon.vue'

const props = withDefaults(
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
const isForbidden = computed(() =>
  ['无权访问', '你没有执行此操作的权限'].includes(props.error?.trim() ?? ''),
)
defineEmits<{ retry: [] }>()
</script>

<template>
  <div v-if="loading" class="state-panel" role="status" aria-live="polite">
    <span class="loading-indicator" aria-hidden="true" />
    <p>正在加载…</p>
  </div>
  <div v-else-if="error && isForbidden" class="permission-state" role="alert">
    <div class="permission-card">
      <div class="permission-illustration" aria-hidden="true">
        <div class="permission-document">
          <span />
          <span />
          <span />
        </div>
        <div class="permission-lock">
          <AppIcon name="lock" :size="32" />
        </div>
      </div>
      <span class="permission-label">访问受限</span>
      <h3>暂时没有访问权限</h3>
      <p class="permission-description">当前账号尚未获得查看此内容的权限。</p>
      <p class="permission-help">请联系所有者或管理员调整授权，开通后即可查看。</p>
      <el-button type="primary" class="permission-retry" @click="$emit('retry')">
        <AppIcon name="refresh" :size="16" />
        <span>重新检查权限</span>
      </el-button>
    </div>
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
.permission-state {
  display: grid;
  place-items: center;
  padding: 1.5rem 0;
}
.permission-card {
  width: min(100%, 42rem);
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 2.5rem 2rem;
  border: 1px solid var(--border);
  border-radius: 20px;
  background: linear-gradient(180deg, var(--accent-soft), var(--surface) 48%);
  box-shadow: 0 8px 32px rgb(35 43 49 / 3%);
  text-align: center;
}
.permission-illustration {
  position: relative;
  width: 8rem;
  height: 7rem;
  margin-bottom: 1.25rem;
  border-radius: 50%;
  background: var(--accent-soft);
}
.permission-document {
  position: absolute;
  top: 0.5rem;
  left: 1.75rem;
  width: 4.5rem;
  height: 5.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
  padding: 1.25rem 0.8rem;
  border: 1px solid var(--el-color-primary-light-8);
  border-radius: 10px;
  background: var(--surface);
  transform: rotate(-9deg);
}
.permission-document span {
  height: 3px;
  border-radius: 2px;
  background: var(--el-color-primary-light-8);
}
.permission-document span:last-child {
  width: 60%;
}
.permission-lock {
  position: absolute;
  right: 0.25rem;
  bottom: 0.25rem;
  width: 4rem;
  height: 4rem;
  display: grid;
  place-items: center;
  border: 4px solid var(--surface);
  border-radius: 18px;
  background: var(--accent);
  color: var(--surface);
}
.permission-label {
  margin-bottom: 0.75rem;
  color: var(--accent);
  font-size: 0.8rem;
  font-weight: 500;
  letter-spacing: 0.12em;
}
.permission-card h3 {
  font-size: 1.5rem;
  letter-spacing: -0.02em;
}
.permission-description {
  margin-top: 0.75rem;
  color: var(--text-secondary);
}
.permission-help {
  margin-top: 0.35rem;
  color: var(--text-secondary);
  font-size: 0.9rem;
}
.permission-retry {
  margin-top: 1.75rem;
  min-height: 44px;
}
.permission-retry :deep(> span) {
  gap: 0.5rem;
}
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
@media (max-width: 600px) {
  .permission-state {
    padding: 0.5rem 0;
  }
  .permission-card {
    padding: 2rem 1.25rem;
    border-radius: 16px;
  }
  .permission-card h3 {
    font-size: 1.35rem;
  }
}
@media (prefers-reduced-motion: reduce) {
  .loading-indicator {
    animation: none;
  }
}
</style>
