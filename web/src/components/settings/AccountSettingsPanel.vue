<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import AsyncState from '@/components/common/AsyncState.vue'
import type { AccountSettings, SettingInput } from '@/types/settings'
import { settingError } from '@/utils/settingsValidation'

const props = defineProps<{
  settings: AccountSettings
  loading: boolean
  error: string
  busy: boolean
  actionError: string
  savedVersion: number
}>()
const emit = defineEmits<{
  retry: []
  save: [value: SettingInput]
  clearError: []
}>()
const opened = ref(false)
const isEditing = ref(false)
const draft = reactive({ key: '', value: '' })
const entries = computed(() =>
  Object.entries(props.settings).sort(([a], [b]) => a.localeCompare(b)),
)
const validationError = computed(() => settingError(draft))
const canSave = computed(() => !props.busy && !validationError.value)
function edit(key?: string) {
  emit('clearError')
  draft.key = key ?? ''
  draft.value = key ? (props.settings[key] ?? '') : ''
  isEditing.value = key !== undefined
  opened.value = true
}
function save() {
  if (!canSave.value) return
  emit('save', { key: draft.key.trim(), value: draft.value.trim() })
}
watch(
  () => props.savedVersion,
  () => {
    opened.value = false
  },
)
</script>

<template>
  <section class="panel account-settings-panel">
    <div class="toolbar">
      <div>
        <h2>账号配置</h2>
        <p class="muted">查看并逐项更新当前账号的配置。</p>
      </div>
      <button class="btn btn-secondary" :disabled="busy" @click="edit()">
        添加配置
      </button>
    </div>
    <AsyncState
      :loading="loading"
      :error="error"
      :has-content="entries.length > 0"
      :empty="!entries.length"
      empty-title="暂无账号配置"
      empty-description="需要自定义配置时，可添加配置键与对应的值。"
      @retry="emit('retry')"
    >
      <div class="setting-list">
        <div v-for="[key, value] in entries" :key="key" class="setting-row">
          <div>
            <strong>{{ key }}</strong>
            <p class="muted">{{ value }}</p>
          </div>
          <button class="btn btn-secondary" :disabled="busy" @click="edit(key)">
            编辑
          </button>
        </div>
      </div>
    </AsyncState>
  </section>
  <el-dialog
    v-model="opened"
    :title="isEditing ? '编辑账号配置' : '添加账号配置'"
    width="480px"
    :close-on-click-modal="false"
    :show-close="!busy"
    :close-on-press-escape="!busy"
  >
    <form class="settings-form" @submit.prevent="save">
      <label class="field">
        <span>配置键</span>
        <input
          v-model="draft.key"
          maxlength="64"
          required
          :disabled="busy || isEditing"
          placeholder="输入配置键"
        />
      </label>
      <label class="field">
        <span>配置值</span>
        <textarea
          v-model="draft.value"
          maxlength="255"
          rows="3"
          required
          :disabled="busy"
          placeholder="输入配置值"
        />
      </label>
      <p
        v-if="
          !isEditing &&
          Object.prototype.hasOwnProperty.call(settings, draft.key.trim())
        "
        class="muted"
      >
        此配置键已存在，保存将更新现有值。
      </p>
      <p
        v-if="draft.key && draft.value && validationError"
        class="form-error"
        role="alert"
      >
        {{ validationError }}
      </p>
      <p v-if="actionError" class="form-error" role="alert">
        {{ actionError }}
      </p>
      <div class="dialog-actions">
        <button
          type="button"
          class="btn btn-secondary"
          :disabled="busy"
          @click="opened = false"
        >
          取消
        </button>
        <button type="submit" class="btn btn-primary" :disabled="!canSave">
          {{ busy ? '正在保存…' : '保存配置' }}
        </button>
      </div>
    </form>
  </el-dialog>
</template>

<style scoped>
.toolbar {
  justify-content: space-between;
}
.account-settings-panel {
  padding: 1.5rem;
}
.toolbar h2 {
  margin: 0 0 0.35rem;
}
.toolbar p {
  margin: 0;
}
.setting-list {
  margin-top: 1.25rem;
}
.setting-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  padding: 1rem 0;
  border-top: 1px solid var(--border, #e8e9ed);
}
.setting-row > div {
  min-width: 0;
}
.setting-row strong {
  font-size: 0.9rem;
  overflow-wrap: anywhere;
}
.setting-row p {
  margin: 0.35rem 0 0;
  font-size: 0.85rem;
  overflow-wrap: anywhere;
}
.settings-form {
  display: grid;
  gap: 1.2rem;
}
.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}
.form-error {
  color: var(--danger);
  font-size: 0.875rem;
}
@media (max-width: 640px) {
  .account-settings-panel {
    padding: 1rem;
  }
}
</style>
