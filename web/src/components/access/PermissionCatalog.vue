<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import AsyncState from '@/components/common/AsyncState.vue'
import type {
  BindingAction,
  Permission,
  PermissionGroup,
  PermissionGroupInput,
  PermissionInput,
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
  createPermission: [value: PermissionInput]
  bind: [id: number, ids: number[], action: BindingAction]
  clearError: []
}>()
const management = ref(false)
const search = ref('')
const editor = ref<'group' | 'permission' | null>(null)
const groupDraft = reactive({
  code: '',
  name: '',
  description: '',
  sort_order: 0,
})
const permissionDraft = reactive({ code: '', description: '' })
const bindingGroup = ref<PermissionGroup | null>(null)
const permissionIds = ref<number[]>([])
const bindingAction = ref<BindingAction>('add')
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
    (editor.value === 'group'
      ? !!groupDraft.code.trim() &&
        !!groupDraft.name.trim() &&
        Number.isInteger(groupDraft.sort_order)
      : !!permissionDraft.code.trim() && !!permissionDraft.description.trim()),
)

function openEditor(type: 'group' | 'permission') {
  if (!management.value) return
  emit('clearError')
  Object.assign(groupDraft, {
    code: '',
    name: '',
    description: '',
    sort_order: 0,
  })
  Object.assign(permissionDraft, { code: '', description: '' })
  editor.value = type
}
function save() {
  if (!management.value || !canSave.value) return
  if (editor.value === 'group') {
    emit('createGroup', {
      ...groupDraft,
      code: groupDraft.code.trim(),
      name: groupDraft.name.trim(),
      description: groupDraft.description.trim(),
    })
  } else {
    emit('createPermission', {
      code: permissionDraft.code.trim(),
      description: permissionDraft.description.trim(),
    })
  }
}
function openBinding(group: PermissionGroup) {
  if (!management.value) return
  emit('clearError')
  bindingGroup.value = group
  permissionIds.value = []
  bindingAction.value = 'add'
}
function bind() {
  if (
    !management.value ||
    !bindingGroup.value ||
    !permissionIds.value.length ||
    props.busy
  )
    return
  emit(
    'bind',
    bindingGroup.value.id,
    [...permissionIds.value],
    bindingAction.value,
  )
}
watch(
  () => props.savedVersion,
  () => {
    editor.value = null
    bindingGroup.value = null
  },
)
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
        {{ management ? '退出管理操作' : '管理全局权限' }}
      </button>
    </div>
    <div v-if="management" class="management-note">
      <strong>全局权限管理</strong>
      <p>更改会影响全局权限目录。仅系统管理员可提交，操作时将核验账号权限。</p>
      <div class="row-actions">
        <button
          class="btn btn-secondary"
          :disabled="busy"
          @click="openEditor('group')"
        >
          新建权限组
        </button>
        <button
          class="btn btn-secondary"
          :disabled="busy"
          @click="openEditor('permission')"
        >
          新建权限
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
          <h2>权限</h2>
          <span class="badge">{{ permissions.length }}</span>
        </div>
        <AsyncState
          :loading="permissionsLoading"
          :error="permissionsError"
          :empty="!visiblePermissions.length"
          @retry="emit('retryPermissions')"
        >
          <article
            v-for="permission in visiblePermissions"
            :key="permission.id"
            class="catalog-row"
          >
            <div>
              <h3>
                {{ permission.description }}
                <span v-if="!permission.is_active" class="badge">已停用</span>
              </h3>
              <code>{{ permission.code }}</code>
            </div>
          </article>
        </AsyncState>
      </section>
    </div>
  </section>

  <el-dialog
    :model-value="!!editor"
    :title="editor === 'group' ? '新建权限组' : '新建权限'"
    width="500px"
    :close-on-click-modal="false"
    :show-close="!busy"
    :close-on-press-escape="!busy"
    @update:model-value="
      (value) => {
        if (!value) editor = null
      }
    "
  >
    <form class="catalog-form" @submit.prevent="save">
      <template v-if="editor === 'group'">
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
      </template>
      <template v-else>
        <label class="field"
          ><span>权限编码</span
          ><input
            v-model="permissionDraft.code"
            required
            :disabled="busy"
            placeholder="例如：room:read"
        /></label>
        <label class="field"
          ><span>权限说明</span
          ><textarea
            v-model="permissionDraft.description"
            required
            rows="3"
            :disabled="busy"
          />
        </label>
      </template>
      <p v-if="actionError" class="form-error" role="alert">
        {{ actionError }}
      </p>
      <div class="dialog-actions">
        <button
          type="button"
          class="btn btn-secondary"
          :disabled="busy"
          @click="editor = null"
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
      <p class="muted">只添加或移除本次选择的权限。当前绑定清单暂不可查看。</p>
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
          v-model="permissionIds"
          multiple
          filterable
          placeholder="选择权限"
          :disabled="busy"
        >
          <el-option
            v-for="permission in permissions"
            :key="permission.id"
            :value="permission.id"
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
          :disabled="busy || !permissionIds.length"
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
  min-width: 0;
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
