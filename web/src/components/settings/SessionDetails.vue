<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useSession } from '@/stores/session'
import { accountNameError } from '@/utils/authValidation'
import { errorMessage } from '@/utils/format'
import AppIcon from '@/components/common/AppIcon.vue'

const { session, isSuperAdmin, updateAccountName } = useSession()
const feedback = ref('')
const editing = ref(false)
const draft = ref('')
const saving = ref(false)
const saveError = ref('')
const nameInput = ref<HTMLInputElement>()
const validationError = computed(() => accountNameError(draft.value))
const canSave = computed(
  () =>
    !saving.value &&
    !validationError.value &&
    draft.value.trim() !== session.value?.account.Username,
)

function editName() {
  draft.value = session.value?.account.Username ?? ''
  saveError.value = ''
  editing.value = true
}
async function saveName() {
  if (!canSave.value) return
  saving.value = true
  saveError.value = ''
  try {
    await updateAccountName(draft.value.trim())
    editing.value = false
    ElMessage.success('账号名称已更新')
  } catch (error) {
    saveError.value = errorMessage(error)
  } finally {
    saving.value = false
  }
}
async function copyId() {
  if (!session.value) return
  try {
    await navigator.clipboard.writeText(session.value.account.ID)
    feedback.value = '账号 ID 已复制'
  } catch {
    feedback.value = '无法访问剪贴板，请选择账号 ID 手动复制'
  }
}
</script>

<template>
  <section v-if="session" class="panel session-details">
    <div class="session-heading">
      <span class="session-icon"><AppIcon name="users" :size="19" /></span>
      <h2>我的账号</h2>
    </div>
    <dl class="session-grid">
      <div class="session-row">
        <dt>账号名称</dt>
        <dd class="session-name">
          <strong>{{ session.account.Username || '未设置' }}</strong>
          <span v-if="isSuperAdmin" class="badge admin-badge">超级管理员</span>
          <el-button text type="primary" @click="editName">
            修改名称
          </el-button>
        </dd>
      </div>
      <div class="session-row">
        <dt>账号 ID</dt>
        <dd class="session-id">
          <code>{{ session.account.ID }}</code>
          <el-button text type="primary" @click="copyId">复制 ID</el-button>
        </dd>
      </div>
    </dl>
    <p class="session-hint">将账号 ID 提供给房源所有者，即可接受协作授权。</p>
    <p v-if="feedback" role="status" class="copy-feedback">{{ feedback }}</p>
  </section>
  <el-dialog
    v-if="session"
    v-model="editing"
    title="修改账号名称"
    width="420px"
    :close-on-click-modal="false"
    :show-close="!saving"
    :close-on-press-escape="!saving"
    @opened="nameInput?.focus()"
  >
    <form class="name-form" @submit.prevent="saveName">
      <label class="field">
        <span>账号名称</span>
        <input
          ref="nameInput"
          v-model="draft"
          name="username"
          autocomplete="nickname"
          placeholder="输入账号名称"
          required
          :disabled="saving"
          :aria-invalid="Boolean(validationError)"
          aria-describedby="account-name-hint account-name-error"
          @input="saveError = ''"
        />
      </label>
      <p id="account-name-hint" class="muted">1–64 个字符，可使用中文、字母或表情。</p>
      <p id="account-name-error" class="form-error" role="alert">
        {{ validationError || saveError }}
      </p>
      <div class="dialog-actions">
        <button
          type="button"
          class="btn btn-secondary"
          :disabled="saving"
          @click="editing = false"
        >
          取消
        </button>
        <button type="submit" class="btn btn-primary" :disabled="!canSave">
          {{ saving ? '正在保存…' : '保存名称' }}
        </button>
      </div>
    </form>
  </el-dialog>
</template>

<style scoped>
.session-details {
  padding: 1rem 1.5rem;
}
.session-heading {
  display: flex;
  gap: 0.6rem;
  align-items: center;
}
.session-icon {
  width: 32px;
  height: 32px;
  display: grid;
  place-items: center;
  border-radius: 9px;
  background: var(--accent-soft);
  color: var(--accent);
  flex-shrink: 0;
}
.session-grid {
  display: grid;
  gap: 0.15rem;
  margin: 0.5rem 0 0;
}
.session-row {
  display: grid;
  grid-template-columns: 5rem minmax(0, 1fr);
  align-items: baseline;
  gap: 0.75rem;
}
dt {
  font-size: 0.8rem;
  color: var(--text-secondary);
}
dd {
  margin: 0;
  font-size: 0.9rem;
  overflow-wrap: anywhere;
}
.session-name,
.session-id {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
  min-width: 0;
}
.session-name strong {
  font-weight: 500;
  min-width: 0;
}
.admin-badge {
  padding: 0.15rem 0.4rem;
  font-size: 0.7rem;
  flex-shrink: 0;
}
code {
  font-size: 0.8rem;
  color: var(--text-secondary);
  user-select: all;
  min-width: 0;
}
.session-hint,
.copy-feedback {
  margin-top: 0.35rem;
  font-size: 0.78rem;
}
.session-hint {
  color: var(--text-secondary);
}
.copy-feedback {
  color: var(--accent);
}
.name-form {
  display: grid;
  gap: 0.75rem;
}
.name-form p {
  font-size: 0.85rem;
}
.form-error {
  color: var(--danger);
}
.form-error:empty {
  display: none;
}
.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  margin-top: 0.5rem;
}
@media (max-width: 640px) {
  .session-details {
    padding: 1rem;
  }
  .session-row {
    grid-template-columns: 1fr;
    gap: 0;
  }
  .session-id {
    column-gap: 0;
  }
  .session-grid {
    gap: 0.35rem;
  }
}
</style>
