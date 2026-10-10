<script setup lang="ts">
import { ref } from 'vue'
import { useSession } from '@/stores/session'
import { formatDate } from '@/utils/format'
import AppIcon from '@/components/common/AppIcon.vue'

const { session, isSuperAdmin } = useSession()
const feedback = ref('')
async function copyId() {
  if (!session.value) return
  try {
    await navigator.clipboard.writeText(session.value.account.ID)
    feedback.value = '账号 ID 已复制'
  } catch {
    feedback.value = '无法访问剪贴板，请选择下方账号 ID 手动复制'
  }
}
</script>

<template>
  <section v-if="session" class="panel session-details">
    <div class="session-heading">
      <span class="session-icon"><AppIcon name="shield" :size="23" /></span>
      <div>
        <h2>我的账号</h2>
        <p>将账号 ID 提供给房源所有者，即可接受协作授权。</p>
      </div>
    </div>
    <dl class="session-grid">
      <div>
        <dt>账号名称</dt>
        <dd>{{ session.account.Username || '未设置' }}</dd>
      </div>
      <div>
        <dt>邮箱</dt>
        <dd>{{ session.account.Email || '未设置' }}</dd>
      </div>
      <div>
        <dt>账号身份</dt>
        <dd>
          <span v-if="isSuperAdmin" class="badge">超级管理员</span>
          <span v-else>普通账号</span>
        </dd>
      </div>
      <div class="session-id">
        <dt>账号 ID</dt>
        <dd>
          <code>{{ session.account.ID }}</code
          ><el-button text type="primary" @click="copyId">复制 ID</el-button>
        </dd>
      </div>
      <div class="session-expiry">
        <dt>会话有效期至</dt>
        <dd>{{ formatDate(session.expires_at) }}</dd>
      </div>
    </dl>
    <p v-if="feedback" role="status" class="copy-feedback">{{ feedback }}</p>
  </section>
</template>

<style scoped>
.session-heading {
  display: flex;
  gap: 0.9rem;
  align-items: center;
}
.session-icon {
  width: 42px;
  height: 42px;
  display: grid;
  place-items: center;
  border-radius: 12px;
  background: var(--accent-soft);
  color: var(--accent);
  flex-shrink: 0;
}
.session-heading p {
  color: var(--text-secondary);
  margin-top: 0.35rem;
  font-size: 0.8rem;
}
.session-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.3rem;
  margin: 1.5rem 0 0;
}
dt {
  font-size: 0.76rem;
  color: var(--text-secondary);
  margin-bottom: 0.4rem;
}
dd {
  margin: 0;
  font-size: 0.9rem;
  overflow-wrap: anywhere;
}
.session-id,
.session-expiry {
  grid-column: 1 / -1;
}
.session-id dd {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.65rem;
}
code {
  font-size: 0.8rem;
  color: #66796e;
  user-select: all;
}
.copy-feedback {
  margin-top: 0.8rem;
  color: var(--accent);
  font-size: 0.78rem;
}
</style>
