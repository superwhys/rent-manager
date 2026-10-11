<script setup lang="ts">
import { computed } from 'vue'
import AppIcon from '@/components/common/AppIcon.vue'

const props = withDefaults(
  defineProps<{
    loading?: boolean
    loadingLayout?: 'list' | 'cards' | 'dashboard' | 'form'
    hasContent?: boolean
    error?: string
    empty?: boolean
    emptyTitle?: string
    emptyDescription?: string
  }>(),
  {
    loadingLayout: 'list',
    hasContent: false,
    emptyTitle: '暂时没有数据',
    emptyDescription: '添加第一条记录，从这里开始。',
  },
)
const isForbidden = computed(() => {
  const error = props.error?.trim() ?? ''
  return ['无权访问', '你没有执行此操作的权限'].some(
    (message) => error === message || error.endsWith(`：${message}`),
  )
})
const showContent = computed(
  () =>
    !isForbidden.value &&
    (props.loading || props.error ? props.hasContent : !props.empty),
)
const skeletonCount = computed(() => (props.loadingLayout === 'list' ? 5 : 6))

function finishLeave(_element: Element, done: () => void) {
  done()
}

defineEmits<{ retry: [] }>()
</script>

<template>
  <div class="async-state" :aria-busy="loading">
    <Transition name="async-state" appear @leave="finishLeave">
      <div v-if="showContent" key="content" class="async-content">
        <Transition name="async-status" @leave="finishLeave">
          <div v-if="loading" class="refresh-status" role="status" aria-live="polite">
            <span class="refresh-dot" aria-hidden="true" />
            正在更新…
          </div>
          <div v-else-if="error" class="retained-error" role="alert">
            <span>{{ error }}</span>
            <el-button text size="small" @click="$emit('retry')">重新加载</el-button>
          </div>
        </Transition>
        <slot />
      </div>
      <div
        v-else-if="error && isForbidden"
        key="forbidden"
        class="permission-state"
        role="alert"
      >
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
      <div
        v-else-if="loading"
        key="loading"
        class="loading-state"
        role="status"
        aria-live="polite"
      >
        <p class="loading-label">正在加载…</p>
        <div class="skeleton-grid" :class="`skeleton-${loadingLayout}`" aria-hidden="true">
          <div
            v-for="index in skeletonCount"
            :key="index"
            class="skeleton-block"
            :class="{ 'skeleton-wide': loadingLayout === 'dashboard' && index > 4 }"
          >
            <span class="skeleton-mark" />
            <div class="skeleton-body">
              <span class="skeleton-line skeleton-title" />
              <span class="skeleton-line" />
              <span class="skeleton-line skeleton-short" />
            </div>
            <span class="skeleton-tail" />
          </div>
        </div>
      </div>
      <div v-else-if="error" key="error" class="state-panel" role="alert">
        <span class="state-symbol">!</span>
        <h3>暂时无法加载</h3>
        <p>{{ error }}</p>
        <el-button @click="$emit('retry')">重新加载</el-button>
      </div>
      <div v-else-if="empty" key="empty" class="state-panel">
        <span class="state-symbol" aria-hidden="true">—</span>
        <h3>{{ emptyTitle }}</h3>
        <p>{{ emptyDescription }}</p>
        <slot name="empty" />
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.async-state {
  position: relative;
  min-width: 0;
}
.async-content {
  min-width: 0;
}
.async-state-enter-active {
  transition: opacity var(--motion-enter) var(--ease-out);
}
.async-status-enter-active {
  transition: opacity var(--motion-fast) var(--ease-out);
}
.async-state-enter-from,
.async-status-enter-from {
  opacity: 0;
}
.refresh-status {
  position: absolute;
  z-index: 1;
  top: -0.75rem;
  right: 0.75rem;
  display: flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.3rem 0.65rem;
  border: 1px solid var(--border);
  border-radius: 20px;
  background: var(--surface);
  color: var(--text-secondary);
  font-size: 0.8rem;
  pointer-events: none;
}
.refresh-dot {
  width: 0.4rem;
  height: 0.4rem;
  border-radius: 50%;
  background: var(--accent);
  animation: loading-pulse 1.6s var(--ease-out) infinite;
}
.retained-error {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.55rem 0.85rem;
  margin-bottom: 0.75rem;
  border: 1px solid var(--border);
  border-radius: 9px;
  background: var(--surface-secondary);
  color: var(--danger);
  overflow-wrap: anywhere;
}
.retained-error .el-button {
  flex-shrink: 0;
}
.loading-state {
  padding: 0.5rem 0;
}
.loading-label {
  margin-bottom: 0.85rem;
  color: var(--text-secondary);
  font-size: 0.9rem;
}
.skeleton-grid {
  display: grid;
  gap: 0.85rem;
  animation: loading-pulse 1.6s var(--ease-out) infinite;
}
.skeleton-block {
  display: flex;
  align-items: center;
  gap: 1rem;
  min-width: 0;
  padding: 1.25rem;
  border: 1px solid var(--border);
  border-radius: 12px;
  background: var(--surface);
}
.skeleton-mark,
.skeleton-line,
.skeleton-tail {
  display: block;
  border-radius: 6px;
  background: var(--border);
}
.skeleton-mark {
  width: 2.5rem;
  height: 2.5rem;
  flex-shrink: 0;
}
.skeleton-body {
  display: grid;
  flex: 1;
  gap: 0.65rem;
  min-width: 0;
}
.skeleton-line {
  width: 85%;
  height: 0.7rem;
}
.skeleton-title {
  width: 50%;
  height: 0.85rem;
}
.skeleton-short {
  width: 65%;
}
.skeleton-tail {
  width: 3rem;
  height: 1.5rem;
  flex-shrink: 0;
}
.skeleton-list .skeleton-short {
  display: none;
}
.skeleton-cards {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}
.skeleton-cards .skeleton-block {
  align-items: flex-start;
  flex-direction: column;
  min-height: 15rem;
}
.skeleton-cards .skeleton-body {
  width: 100%;
  align-content: start;
}
.skeleton-cards .skeleton-tail {
  width: 100%;
}
.skeleton-dashboard {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}
.skeleton-dashboard .skeleton-block {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 2.5rem;
  align-content: center;
  min-height: 10rem;
}
.skeleton-dashboard .skeleton-mark {
  grid-column: 2;
  grid-row: 1;
}
.skeleton-dashboard .skeleton-body {
  grid-column: 1;
  grid-row: 1;
}
.skeleton-dashboard .skeleton-tail {
  grid-column: 1 / -1;
  width: 60%;
  height: 1.75rem;
}
.skeleton-dashboard .skeleton-wide {
  grid-column: span 2;
  min-height: 16rem;
}
.skeleton-form {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}
.skeleton-form .skeleton-block {
  border: 0;
  padding: 0;
  background: transparent;
}
.skeleton-form .skeleton-mark,
.skeleton-form .skeleton-short,
.skeleton-form .skeleton-tail {
  display: none;
}
.skeleton-form .skeleton-line:not(.skeleton-title) {
  width: 100%;
  height: 2.8rem;
}
@keyframes loading-pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}
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
@media (max-width: 900px) {
  .skeleton-dashboard {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
@media (max-width: 1200px) {
  .skeleton-cards {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
@media (max-width: 680px) {
  .skeleton-cards {
    grid-template-columns: minmax(0, 1fr);
  }
}
@media (max-width: 600px) {
  .skeleton-dashboard {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .skeleton-form {
    grid-template-columns: minmax(0, 1fr);
  }
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
  .skeleton-grid,
  .refresh-dot {
    animation: none;
  }
  .async-state-enter-active,
  .async-status-enter-active {
    transition-duration: 100ms !important;
  }
}
</style>
