<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import AsyncState from '@/components/common/AsyncState.vue'
import type {
  BindingAction,
  Permission,
  PermissionGroup,
  PermissionGroupInput,
} from '@/types/access'

const props = defineProps<{
  groups: PermissionGroup[]
  permissions: Permission[]
  groupsLoading: boolean
  groupsError: string
  permissionsLoading: boolean
  permissionsError: string
  busy: boolean
  actionError: string
  savedVersion: number
}>()
const emit = defineEmits<{
  retryGroups: []
  retryPermissions: []
  createGroup: [value: PermissionGroupInput]
  bind: [id: number, codes: string[], action: BindingAction]
  clearError: []
}>()
const management = ref(false)
const search = ref('')
const editor = ref(false)
const groupDraft = reactive({
  code: '',
  name: '',
  description: '',
  sort_order: 0,
})
const bindingGroup = ref<PermissionGroup | null>(null)
const permissionCodes = ref<string[]>([])
const bindingAction = ref<BindingAction>('add')
const bindingOptions = computed(() =>
  props.permissions.filter((permission) =>
    bindingAction.value === 'remove'
      ? bindingGroup.value?.permission_codes.includes(permission.code)
      : !bindingGroup.value?.permission_codes.includes(permission.code),
  ),
)

function boundPermissionNames(group: PermissionGroup) {
  return group.permission_codes.join('、') || '尚未绑定权限'
}
const visibleGroups = computed(() =>
  props.groups.filter((group) =>
    `${group.name} ${group.code} ${group.description ?? ''}`
      .toLowerCase()
      .includes(search.value.toLowerCase()),
  ),
)
const visiblePermissions = computed(() =>
  props.permissions.filter((permission) =>
    `${permission.code} ${permission.description}`
      .toLowerCase()
      .includes(search.value.toLowerCase()),
  ),
)
const canSave = computed(
  () =>
    !props.busy &&
    !!groupDraft.code.trim() &&
    !!groupDraft.name.trim() &&
    Number.isInteger(groupDraft.sort_order),
)

function openEditor() {
  if (!management.value) return
  emit('clearError')
  Object.assign(groupDraft, {
    code: '',
    name: '',
    description: '',
    sort_order: 0,
  })
  editor.value = true
}
function save() {
  if (!management.value || !canSave.value) return
  emit('createGroup', {
    ...groupDraft,
    code: groupDraft.code.trim(),
    name: groupDraft.name.trim(),
    description: groupDraft.description.trim(),
  })
}
function openBinding(group: PermissionGroup) {
  if (!management.value) return
  emit('clearError')
  bindingGroup.value = group
  permissionCodes.value = []
  bindingAction.value = 'add'
}
function bind() {
  if (
    !management.value ||
    !bindingGroup.value ||
    !permissionCodes.value.length ||
    props.busy
  )
    return
  emit(
    'bind',
    bindingGroup.value.id,
    [...permissionCodes.value],
    bindingAction.value,
  )
}
watch(
  () => props.savedVersion,
  () => {
    editor.value = false
    bindingGroup.value = null
  },
)
watch(bindingAction, () => {
  permissionCodes.value = []
})
</script>

<template>
  <section class="permission-catalog">
    <div class="catalog-toolbar">
      <label class="field search-field">
        <span class="sr-only">搜索权限目录</span>
        <input
          v-model="search"
          type="search"
          placeholder="搜索权限名称或编码"
        />
      </label>
      <button
        class="btn btn-secondary"
        :aria-pressed="management"
        :disabled="busy"
        @click="management = !management"
      >
        {{ management ? '退出管理操作' : '管理权限组' }}
      </button>
    </div>
    <div v-if="management" class="management-note">
      <strong>全局权限组管理</strong>
      <p>更改会影响使用这些权限组的角色。仅超级管理员可提交。</p>
      <div class="row-actions">
        <button
          class="btn btn-secondary"
          :disabled="busy"
          @click="openEditor"
        >
          新建权限组
        </button>
      </div>
    </div>
    <div class="catalog-grid">
      <section class="panel catalog-section">
        <div class="section-heading">
          <h2>权限组</h2>
          <span class="badge">{{ groups.length }}</span>
        </div>
        <AsyncState
          :loading="groupsLoading"
          :error="groupsError"
          :empty="!visibleGroups.length"
          @retry="emit('retryGroups')"
        >
          <article
            v-for="group in visibleGroups"
            :key="group.id"
            class="catalog-row"
          >
            <div>
              <h3>
                {{ group.name }}
                <span v-if="!group.is_active" class="badge">已停用</span>
              </h3>
              <code>{{ group.code }}</code>
              <p v-if="group.description" class="muted">
                {{ group.description }}
              </p>
              <p class="muted">已绑定：{{ boundPermissionNames(group) }}</p>
            </div>
            <button
              v-if="management"
              class="btn btn-secondary"
              :disabled="busy"
              @click="openBinding(group)"
            >
              调整权限
            </button>
          </article>
        </AsyncState>
      </section>
      <section class="panel catalog-section">
        <div class="section-heading">
          <h2>内置权限</h2>
          <span class="badge">{{ permissions.length }}</span>
        </div>
        <p class="catalog-note muted">系统内置权限，可在权限组中选择并组合授权。</p>
        <AsyncState
          :loading="permissionsLoading"
          :error="permissionsError"
          :empty="!visiblePermissions.length"
          @retry="emit('retryPermissions')"
        >
          <article
            v-for="permission in visiblePermissions"
            :key="permission.code"
            class="catalog-row"
          >
            <div>
              <h3>{{ permission.description }}</h3>
              <code>{{ permission.code }}</code>
            </div>
          </article>
        </AsyncState>
      </section>
    </div>
  </section>

  <el-dialog
    :model-value="editor"
    title="新建权限组"
    width="500px"
    :close-on-click-modal="false"
    :show-close="!busy"
    :close-on-press-escape="!busy"
    @update:model-value="
      (value) => {
        if (!value) editor = false
      }
    "
  >
    <form class="catalog-form" @submit.prevent="save">
      <label class="field"
        ><span>名称</span
        ><input v-model="groupDraft.name" required :disabled="busy"
      /></label>
      <label class="field"
        ><span>编码</span
        ><input
          v-model="groupDraft.code"
          required
          :disabled="busy"
          placeholder="例如：room-management"
      /></label>
      <label class="field"
        ><span>说明</span
        ><textarea
          v-model="groupDraft.description"
          rows="3"
          :disabled="busy"
        />
      </label>
      <label class="field"
        ><span>排序</span
        ><input
          v-model.number="groupDraft.sort_order"
          type="number"
          step="1"
          required
          :disabled="busy"
      /></label>
      <p v-if="actionError" class="form-error" role="alert">
        {{ actionError }}
      </p>
      <div class="dialog-actions">
        <button
          type="button"
          class="btn btn-secondary"
          :disabled="busy"
          @click="editor = false"
        >
          取消
        </button>
        <button type="submit" class="btn btn-primary" :disabled="!canSave">
          {{ busy ? '正在创建…' : '创建' }}
        </button>
      </div>
    </form>
  </el-dialog>

  <el-dialog
    :model-value="!!bindingGroup"
    :title="`${bindingGroup?.name ?? ''} · 调整权限`"
    width="520px"
    :close-on-click-modal="false"
    :show-close="!busy"
    :close-on-press-escape="!busy"
    @update:model-value="
      (value) => {
        if (!value) bindingGroup = null
      }
    "
  >
    <form class="catalog-form" @submit.prevent="bind">
      <p class="muted">
        当前绑定：{{ bindingGroup ? boundPermissionNames(bindingGroup) : '' }}
      </p>
      <label class="field">
        <span>本次操作</span>
        <select v-model="bindingAction" :disabled="busy">
          <option value="add">添加权限</option>
          <option value="remove">移除权限</option>
        </select>
      </label>
      <label class="field">
        <span>权限</span>
        <el-select
          v-model="permissionCodes"
          multiple
          filterable
          placeholder="选择权限"
          :disabled="busy"
        >
          <el-option
            v-for="permission in bindingOptions"
            :key="permission.code"
            :value="permission.code"
            :label="`${permission.description} · ${permission.code}`"
          />
        </el-select>
      </label>
      <p v-if="actionError" class="form-error" role="alert">
        {{ actionError }}
      </p>
      <div class="dialog-actions">
        <button
          type="button"
          class="btn btn-secondary"
          :disabled="busy"
          @click="bindingGroup = null"
        >
          取消
        </button>
        <button
          type="submit"
          class="btn btn-primary"
          :disabled="busy || !permissionCodes.length"
        >
          {{
            busy
              ? '正在提交…'
              : bindingAction === 'add'
                ? '添加所选权限'
                : '移除所选权限'
          }}
        </button>
      </div>
    </form>
  </el-dialog>
</template>

<style scoped>
.catalog-toolbar,
.section-heading,
.row-actions,
.dialog-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}
.catalog-toolbar {
  justify-content: space-between;
  margin-bottom: 1.5rem;
}
.search-field {
  width: min(100%, 22rem);
}
.catalog-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.5rem;
  align-items: start;
}
.catalog-section {
  padding: 1.5rem;
  min-width: 0;
}
.section-heading {
  margin-bottom: 1rem;
}
.section-heading h2 {
  margin: 0;
  font-size: 1.1rem;
}
.catalog-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.8rem;
  padding: 1rem 0;
  border-top: 1px solid var(--border, #e8e9ed);
}
.catalog-row > div {
  flex: 1;
  min-width: 0;
}
.catalog-row > button {
  flex-shrink: 0;
  white-space: nowrap;
}
.catalog-row h3 {
  font-size: 0.92rem;
  margin: 0 0 0.4rem;
  overflow-wrap: anywhere;
}
.catalog-row code {
  font-size: 0.76rem;
  color: var(--text-secondary, #68717d);
  overflow-wrap: anywhere;
}
.catalog-row p {
  font-size: 0.8rem;
  margin: 0.5rem 0 0;
  overflow-wrap: anywhere;
}
.catalog-note {
  font-size: 0.8rem;
  margin: 0 0 1rem;
}
.management-note {
  padding: 1.2rem 1.5rem;
  margin-bottom: 1.5rem;
  border-radius: 1rem;
  background: var(--accent-soft);
}
.management-note p {
  font-size: 0.86rem;
  margin: 0.45rem 0 1rem;
}
.catalog-form {
  display: grid;
  gap: 1.1rem;
}
.dialog-actions {
  justify-content: flex-end;
  margin-top: 0.5rem;
}
.form-error {
  color: var(--danger);
  font-size: 0.875rem;
}
@media (max-width: 850px) {
  .catalog-grid {
    grid-template-columns: 1fr;
  }
}
@media (max-width: 640px) {
  .catalog-section {
    padding: 1rem;
  }
}
</style>
