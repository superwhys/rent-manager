<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import AsyncState from '@/components/common/AsyncState.vue'
import type {
  BindingAction,
  PermissionGroup,
  Role,
  RoleInput,
} from '@/types/access'

const props = defineProps<{
  roles: Role[]
  groups: PermissionGroup[]
  loading: boolean
  error: string
  busy: boolean
  actionError: string
  savedVersion: number
}>()
const emit = defineEmits<{
  retry: []
  save: [id: number | null, value: RoleInput]
  remove: [role: Role]
  bind: [id: number, ids: number[], action: BindingAction]
  clearError: []
}>()
const editing = ref(false)
const selectedId = ref<number | null>(null)
const draft = reactive({ name: '', is_default_owner: false, is_active: true })
const bindingRole = ref<Role | null>(null)
const bindingIds = ref<number[]>([])
const bindingAction = ref<BindingAction>('add')
const canSave = computed(() => draft.name.trim().length > 0 && !props.busy)

function edit(role?: Role) {
  emit('clearError')
  selectedId.value = role?.id ?? null
  Object.assign(draft, {
    name: role?.name ?? '',
    is_default_owner: role?.is_default_owner ?? false,
    is_active: role?.is_active ?? true,
  })
  editing.value = true
}
function save() {
  if (!canSave.value) return
  emit('save', selectedId.value, { ...draft, name: draft.name.trim() })
}
function openBinding(role: Role) {
  emit('clearError')
  bindingRole.value = role
  bindingIds.value = []
  bindingAction.value = 'add'
}
function bind() {
  if (!bindingRole.value || !bindingIds.value.length || props.busy) return
  emit('bind', bindingRole.value.id, [...bindingIds.value], bindingAction.value)
}
watch(
  () => props.savedVersion,
  () => {
    editing.value = false
    bindingRole.value = null
  },
)
</script>

<template>
  <section class="panel roles-panel">
    <div class="toolbar">
      <div>
        <h2>我的角色</h2>
        <p class="muted">为协作者准备合适的访问权限。</p>
      </div>
      <button class="btn btn-primary" :disabled="busy" @click="edit()">
        新建角色
      </button>
    </div>
    <AsyncState
      :loading="loading"
      :error="error"
      :empty="!roles.length"
      @retry="emit('retry')"
    >
      <div class="role-list">
        <article v-for="role in roles" :key="role.id" class="role-row">
          <div class="role-symbol" aria-hidden="true">
            {{ role.name.slice(0, 1) }}
          </div>
          <div class="role-info">
            <h3>
              {{ role.name }}
              <span v-if="!role.is_active" class="badge">已停用</span>
            </h3>
            <p class="muted">
              {{ role.is_default_owner ? '默认所有者角色' : '协作角色' }} · #{{
                role.id
              }}
            </p>
          </div>
          <div class="row-actions">
            <button
              class="btn btn-secondary"
              :disabled="busy"
              @click="openBinding(role)"
            >
              权限组
            </button>
            <button
              class="btn btn-secondary"
              :disabled="busy"
              @click="edit(role)"
            >
              编辑
            </button>
            <button
              class="btn btn-danger"
              :disabled="busy"
              @click="emit('remove', role)"
            >
              删除
            </button>
          </div>
        </article>
      </div>
    </AsyncState>
    <p
      v-if="actionError && !editing && !bindingRole"
      class="form-error"
      role="alert"
    >
      {{ actionError }}
    </p>
  </section>

  <el-dialog
    v-model="editing"
    :title="selectedId ? '编辑角色' : '新建角色'"
    width="480px"
    :close-on-click-modal="false"
    :show-close="!busy"
    :close-on-press-escape="!busy"
  >
    <form class="role-form" @submit.prevent="save">
      <label class="field">
        <span>角色名称</span>
        <input
          v-model="draft.name"
          required
          placeholder="例如：物业管家"
          :disabled="busy"
        />
      </label>
      <label class="check-field">
        <input
          v-model="draft.is_default_owner"
          type="checkbox"
          :disabled="busy"
        />
        <span>设为默认所有者角色</span>
      </label>
      <label v-if="selectedId" class="check-field">
        <input v-model="draft.is_active" type="checkbox" :disabled="busy" />
        <span>启用角色</span>
      </label>
      <p v-if="actionError" class="form-error" role="alert">
        {{ actionError }}
      </p>
      <div class="dialog-actions">
        <button
          type="button"
          class="btn btn-secondary"
          :disabled="busy"
          @click="editing = false"
        >
          取消
        </button>
        <button type="submit" class="btn btn-primary" :disabled="!canSave">
          {{ busy ? '正在保存…' : '保存角色' }}
        </button>
      </div>
    </form>
  </el-dialog>

  <el-dialog
    :model-value="!!bindingRole"
    :title="`${bindingRole?.name ?? ''} · 调整权限组`"
    width="520px"
    :close-on-click-modal="false"
    :show-close="!busy"
    :close-on-press-escape="!busy"
    @update:model-value="
      (value) => {
        if (!value) bindingRole = null
      }
    "
  >
    <form class="role-form" @submit.prevent="bind">
      <p class="muted">
        选择本次要添加或移除的权限组。当前绑定清单暂不可查看。
      </p>
      <label class="field">
        <span>本次操作</span>
        <select v-model="bindingAction" :disabled="busy">
          <option value="add">添加权限组</option>
          <option value="remove">移除权限组</option>
        </select>
      </label>
      <label class="field">
        <span>权限组</span>
        <el-select
          v-model="bindingIds"
          multiple
          placeholder="选择权限组"
          :disabled="busy"
        >
          <el-option
            v-for="group in groups"
            :key="group.id"
            :value="group.id"
            :label="`${group.name} · ${group.code}`"
          />
        </el-select>
      </label>
      <p v-if="!groups.length" class="muted">
        暂无可选权限组，请先在权限目录中查看或创建。
      </p>
      <p v-if="actionError" class="form-error" role="alert">
        {{ actionError }}
      </p>
      <div class="dialog-actions">
        <button
          type="button"
          class="btn btn-secondary"
          :disabled="busy"
          @click="bindingRole = null"
        >
          取消
        </button>
        <button
          type="submit"
          class="btn btn-primary"
          :disabled="busy || !bindingIds.length"
        >
          {{
            busy
              ? '正在提交…'
              : bindingAction === 'add'
                ? '添加所选权限组'
                : '移除所选权限组'
          }}
        </button>
      </div>
    </form>
  </el-dialog>
</template>

<style scoped>
.toolbar {
  justify-content: space-between;
}
.roles-panel {
  padding: 1.5rem;
}
.toolbar h2 {
  margin: 0 0 0.35rem;
}
.toolbar p {
  margin: 0;
}
.role-list {
  margin-top: 1.25rem;
}
.role-row {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1.1rem 0;
  border-top: 1px solid var(--border, #e8e9ed);
}
.role-symbol {
  display: grid;
  place-items: center;
  width: 2.75rem;
  height: 2.75rem;
  flex: 0 0 auto;
  color: var(--accent);
  background: var(--accent-soft);
  border-radius: 0.85rem;
  font-weight: 600;
}
.role-info {
  flex: 1;
  min-width: 0;
}
.role-info h3 {
  margin: 0;
  font-size: 0.96rem;
  overflow-wrap: anywhere;
}
.role-info p {
  margin: 0.25rem 0 0;
  font-size: 0.8rem;
}
.row-actions,
.dialog-actions {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}
.dialog-actions {
  justify-content: flex-end;
  margin-top: 0.6rem;
}
.role-form {
  display: grid;
  gap: 1.25rem;
}
.check-field {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  min-height: 2.75rem;
}
.check-field input {
  width: 1.05rem;
  height: 1.05rem;
}
.form-error {
  color: var(--danger);
  font-size: 0.875rem;
}
@media (max-width: 640px) {
  .roles-panel {
    padding: 1rem;
  }
  .role-row {
    flex-wrap: wrap;
  }
  .row-actions {
    width: 100%;
    justify-content: flex-end;
  }
}
</style>
