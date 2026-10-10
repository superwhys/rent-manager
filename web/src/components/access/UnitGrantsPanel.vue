<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import AsyncState from '@/components/common/AsyncState.vue'
import type { Role, UnitGrant } from '@/types/access'
import type { Unit } from '@/types/units'
import { isAccountId } from '@/utils/accessValidation'

const props = defineProps<{
  units: Unit[]
  unitId?: number
  grants: UnitGrant[]
  roles: Role[]
  loading: boolean
  error: string
  busy: boolean
  actionError: string
  savedVersion: number
  offset: number
  hasNext: boolean
}>()
const emit = defineEmits<{
  selectUnit: [id: number | undefined]
  filter: [accountId: string, status: string]
  page: [offset: number]
  grant: [accountId: string, roleIds: number[]]
  revoke: [grant: UnitGrant]
  retry: []
  clearError: []
}>()
const showGrant = ref(false)
const accountId = ref('')
const roleIds = ref<number[]>([])
const filters = reactive({ accountId: '', status: 'active' })
const filterError = ref('')
const selectedUnit = computed(() =>
  props.units.find((unit) => unit.id === props.unitId),
)
const validAccount = computed(() => isAccountId(accountId.value.trim()))
const activeRoles = computed(() => props.roles.filter((role) => role.is_active))

function filter() {
  const id = filters.accountId.trim()
  if (id && !isAccountId(id)) {
    filterError.value = '请输入 64 位小写十六进制账号 ID。'
    return
  }
  filterError.value = ''
  emit('filter', id, filters.status)
}
function openGrant() {
  emit('clearError')
  accountId.value = ''
  roleIds.value = []
  showGrant.value = true
}
function grant() {
  if (
    !validAccount.value ||
    !roleIds.value.length ||
    props.busy ||
    !selectedUnit.value?.is_owner
  )
    return
  emit('grant', accountId.value.trim(), [...roleIds.value])
}
watch(
  () => props.savedVersion,
  () => {
    showGrant.value = false
  },
)
watch(
  () => props.unitId,
  () => {
    showGrant.value = false
    filters.accountId = ''
    filters.status = 'active'
    filterError.value = ''
  },
)
</script>

<template>
  <section class="panel grants-panel">
    <div class="toolbar grants-heading">
      <div>
        <h2>单元协作</h2>
        <p class="muted">为已注册的账号授予角色，按单元管理访问范围。</p>
      </div>
      <button
        class="btn btn-primary"
        :disabled="busy || !selectedUnit?.is_owner"
        @click="openGrant"
      >
        添加授权
      </button>
    </div>
    <form class="grant-filters" @submit.prevent="filter">
      <label class="field">
        <span>管理单元</span>
        <el-select
          :model-value="unitId"
          size="large"
          placeholder="选择一个单元"
          :disabled="busy"
          @update:model-value="(value) => emit('selectUnit', value)"
        >
          <el-option
            v-for="unit in units"
            :key="unit.id"
            :value="unit.id"
            :label="unit.name"
          />
        </el-select>
      </label>
      <label class="field account-filter">
        <span>账号 ID</span>
        <input
          v-model="filters.accountId"
          placeholder="全部协作者"
          :disabled="busy"
        />
      </label>
      <label class="field">
        <span>授权状态</span>
        <select v-model="filters.status" :disabled="busy">
          <option value="active">有效</option>
          <option value="disabled">已停用</option>
          <option value="">全部状态</option>
        </select>
      </label>
      <button
        class="btn btn-secondary"
        type="submit"
        :disabled="busy || loading || !selectedUnit?.is_owner"
      >
        筛选
      </button>
    </form>
    <p v-if="filterError" class="form-error" role="alert">{{ filterError }}</p>
    <div v-if="selectedUnit && !selectedUnit.is_owner" class="ownership-note">
      仅单元所有者可以查看和管理协作授权。
    </div>
    <AsyncState
      v-else
      :loading="loading"
      :error="error"
      :empty="!grants.length"
      :empty-title="unitId ? '暂无协作授权' : '先选择一个单元'"
      :empty-description="
        unitId
          ? '你拥有自己的单元；需要协作时，再为其他账号添加授权。'
          : '创建单元后，就可以在这里管理协作权限。'
      "
      @retry="emit('retry')"
    >
      <div class="grant-list">
        <article v-for="item in grants" :key="item.id" class="grant-row">
          <div class="grant-info">
            <h3>{{ item.role?.name ?? `角色 #${item.role_id}` }}</h3>
            <code>{{ item.account_id }}</code>
          </div>
          <span class="badge">{{
            item.status === 'active' ? '有效授权' : '已停用'
          }}</span>
          <button
            class="btn btn-danger"
            :disabled="busy || item.status !== 'active'"
            @click="emit('revoke', item)"
          >
            撤销授权
          </button>
        </article>
      </div>
    </AsyncState>
    <div v-if="offset || hasNext" class="pagination">
      <button
        class="btn btn-secondary"
        :disabled="loading || busy || !offset"
        @click="emit('page', Math.max(0, offset - 20))"
      >
        上一页
      </button>
      <span class="muted">第 {{ Math.floor(offset / 20) + 1 }} 页</span>
      <button
        class="btn btn-secondary"
        :disabled="loading || busy || !hasNext"
        @click="emit('page', offset + 20)"
      >
        下一页
      </button>
    </div>
    <p v-if="actionError && !showGrant" class="form-error" role="alert">
      {{ actionError }}
    </p>
  </section>

  <el-dialog
    v-model="showGrant"
    title="添加协作授权"
    width="540px"
    :close-on-click-modal="false"
    :show-close="!busy"
    :close-on-press-escape="!busy"
  >
    <form class="grant-form" @submit.prevent="grant">
      <p class="muted">授权单元：{{ selectedUnit?.name }}</p>
      <label class="field">
        <span>协作者账号 ID</span>
        <input
          v-model="accountId"
          required
          maxlength="64"
          placeholder="请输入完整账号 ID"
          :disabled="busy"
        />
        <small :class="accountId && !validAccount ? 'form-error' : 'muted'"
          >由协作者提供其 64 位小写十六进制账号 ID。</small
        >
      </label>
      <label class="field">
        <span>授予角色</span>
        <el-select
          v-model="roleIds"
          multiple
          placeholder="选择角色"
          :disabled="busy"
        >
          <el-option
            v-for="role in activeRoles"
            :key="role.id"
            :value="role.id"
            :label="role.name"
          />
        </el-select>
      </label>
      <p v-if="!activeRoles.length" class="muted">
        请先在“我的角色”中新建并启用角色。
      </p>
      <p v-if="actionError" class="form-error" role="alert">
        {{ actionError }}
      </p>
      <div class="dialog-actions">
        <button
          type="button"
          class="btn btn-secondary"
          :disabled="busy"
          @click="showGrant = false"
        >
          取消
        </button>
        <button
          type="submit"
          class="btn btn-primary"
          :disabled="busy || !validAccount || !roleIds.length"
        >
          {{ busy ? '正在授权…' : '确认授权' }}
        </button>
      </div>
    </form>
  </el-dialog>
</template>

<style scoped>
.toolbar {
  justify-content: space-between;
}
.grants-panel {
  padding: 1.5rem;
}
.grants-heading h2 {
  margin: 0 0 0.35rem;
}
.grants-heading p {
  margin: 0;
}
.grant-filters {
  display: flex;
  align-items: flex-end;
  gap: 1rem;
  flex-wrap: wrap;
  margin: 1.5rem 0;
}
.grant-filters .field {
  flex: 1 1 10rem;
}
.grant-filters .account-filter {
  flex: 2 1 16rem;
}
.grant-row {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1.1rem 0;
  border-top: 1px solid var(--border, #e8e9ed);
}
.grant-info {
  flex: 1;
  min-width: 0;
}
.grant-info h3 {
  margin: 0 0 0.4rem;
  font-size: 0.95rem;
}
.grant-info code {
  font-size: 0.75rem;
  overflow-wrap: anywhere;
  color: var(--text-secondary, #68717d);
}
.grant-form {
  display: grid;
  gap: 1.1rem;
}
.dialog-actions,
.pagination {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.75rem;
  flex-wrap: wrap;
}
.pagination {
  margin-top: 1.25rem;
}
.form-error {
  color: var(--danger);
  font-size: 0.875rem;
}
.ownership-note {
  padding: 3rem 1rem;
  color: var(--text-secondary, #68717d);
  text-align: center;
}
@media (max-width: 640px) {
  .grants-panel {
    padding: 1rem;
  }
  .grant-row {
    flex-wrap: wrap;
  }
  .grant-info {
    flex-basis: 100%;
  }
}
</style>
