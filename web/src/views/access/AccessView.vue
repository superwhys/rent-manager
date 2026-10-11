<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import PageHeader from '@/components/common/PageHeader.vue'
import AppIcon from '@/components/common/AppIcon.vue'
import RolesPanel from '@/components/access/RolesPanel.vue'
import PermissionCatalog from '@/components/access/PermissionCatalog.vue'
import UnitGrantsPanel from '@/components/access/UnitGrantsPanel.vue'
import * as accessApi from '@/api/access'
import { useUnits } from '@/stores/units'
import { useSession } from '@/stores/session'
import { errorMessage } from '@/utils/format'
import { isAccountId } from '@/utils/accessValidation'
import type {
  BindingAction,
  Permission,
  PermissionGroup,
  PermissionGroupInput,
  Role,
  RoleInput,
  UnitGrant,
} from '@/types/access'

const {
  units,
  selectedUnitId,
  selectedUnit,
  error: unitsError,
  loadUnits,
} = useUnits()
const { isSuperAdmin } = useSession()
const activeTab = ref('roles')
const animateTabs = ref(false)
const roles = ref<Role[]>([])
const groups = ref<PermissionGroup[]>([])
const permissions = ref<Permission[]>([])
const grants = ref<UnitGrant[]>([])
const rolesLoading = ref(false)
const groupsLoading = ref(false)
const permissionsLoading = ref(false)
const grantsLoading = ref(false)
const rolesError = ref('')
const groupsError = ref('')
const permissionsError = ref('')
const grantsError = ref('')
const actionError = ref('')
const busy = ref(false)
const roleVersion = ref(0)
const catalogVersion = ref(0)
const grantVersion = ref(0)
const grantOffset = ref(0)
const grantHasNext = ref(false)
const grantAccountFilter = ref('')
const grantStatusFilter = ref('active')
const deletingRole = ref<Role | null>(null)
const revokingGrant = ref<UnitGrant | null>(null)
let rolesRequest = 0
let groupsRequest = 0
let permissionsRequest = 0
let grantsRequest = 0
let disposed = false

async function loadRoles() {
  const sequence = ++rolesRequest
  rolesLoading.value = true
  rolesError.value = ''
  try {
    const result = await accessApi.listRoles()
    if (sequence === rolesRequest && !disposed) roles.value = result
  } catch (error) {
    if (sequence === rolesRequest && !disposed)
      rolesError.value = errorMessage(error)
  } finally {
    if (sequence === rolesRequest && !disposed) rolesLoading.value = false
  }
}
async function loadGroups() {
  const sequence = ++groupsRequest
  groupsLoading.value = true
  groupsError.value = ''
  try {
    const result = await accessApi.listPermissionGroups()
    if (sequence === groupsRequest && !disposed) groups.value = result
  } catch (error) {
    if (sequence === groupsRequest && !disposed)
      groupsError.value = errorMessage(error)
  } finally {
    if (sequence === groupsRequest && !disposed) groupsLoading.value = false
  }
}
async function loadPermissions() {
  const sequence = ++permissionsRequest
  permissionsLoading.value = true
  permissionsError.value = ''
  try {
    const result = await accessApi.listPermissions()
    if (sequence === permissionsRequest && !disposed) permissions.value = result
  } catch (error) {
    if (sequence === permissionsRequest && !disposed)
      permissionsError.value = errorMessage(error)
  } finally {
    if (sequence === permissionsRequest && !disposed)
      permissionsLoading.value = false
  }
}
async function loadGrants() {
  const sequence = ++grantsRequest
  const unitId = selectedUnitId.value
  grants.value = []
  grantsError.value = ''
  grantHasNext.value = false
  if (!unitId || !selectedUnit.value?.is_owner) {
    grantsLoading.value = false
    return
  }
  grantsLoading.value = true
  try {
    const result = await accessApi.listUnitGrants(unitId, {
      account_id: grantAccountFilter.value || undefined,
      status: grantStatusFilter.value || undefined,
      limit: 21,
      offset: grantOffset.value,
    })
    if (sequence === grantsRequest && !disposed) {
      grants.value = result.slice(0, 20)
      grantHasNext.value = result.length > 20
    }
  } catch (error) {
    if (sequence === grantsRequest && !disposed)
      grantsError.value = errorMessage(error)
  } finally {
    if (sequence === grantsRequest && !disposed) grantsLoading.value = false
  }
}
async function initializeUnits() {
  try {
    await loadUnits()
    if (!disposed) {
      if (unitsError.value) grantsError.value = unitsError.value
      else await loadGrants()
    }
  } catch (error) {
    if (!disposed) grantsError.value = errorMessage(error)
  }
}
async function saveRole(id: number | null, value: RoleInput) {
  if (busy.value || !value.name.trim()) return
  busy.value = true
  actionError.value = ''
  try {
    if (id) await accessApi.updateRole(id, value)
    else await accessApi.createRole(value)
    roleVersion.value++
    ElMessage.success('角色已保存')
    await loadRoles()
  } catch (error) {
    actionError.value = errorMessage(error)
  } finally {
    busy.value = false
  }
}
async function removeRole() {
  if (busy.value || !deletingRole.value) return
  busy.value = true
  actionError.value = ''
  try {
    await accessApi.deleteRole(deletingRole.value.id)
    deletingRole.value = null
    ElMessage.success('角色已删除')
    await loadRoles()
    await loadGrants()
  } catch (error) {
    actionError.value = errorMessage(error)
  } finally {
    busy.value = false
  }
}
async function bindRoleGroups(
  id: number,
  ids: number[],
  action: BindingAction,
) {
  if (busy.value || !ids.length) return
  busy.value = true
  actionError.value = ''
  try {
    await accessApi.changeRoleGroups(id, ids, action)
    roleVersion.value++
    ElMessage.success(
      action === 'add' ? '所选权限组已添加' : '所选权限组已移除',
    )
    await loadRoles()
    if (rolesError.value) ElMessage.warning('权限组已调整，角色列表刷新失败，请重试')
  } catch (error) {
    actionError.value = errorMessage(error)
  } finally {
    busy.value = false
  }
}
async function createGroup(value: PermissionGroupInput) {
  if (busy.value || !value.code.trim() || !value.name.trim()) return
  busy.value = true
  actionError.value = ''
  try {
    await accessApi.createPermissionGroup(value)
    catalogVersion.value++
    ElMessage.success('权限组已创建')
    await loadGroups()
  } catch (error) {
    actionError.value = errorMessage(error)
  } finally {
    busy.value = false
  }
}
async function bindGroupPermissions(
  id: number,
  codes: string[],
  action: BindingAction,
) {
  if (busy.value || !codes.length) return
  busy.value = true
  actionError.value = ''
  try {
    await accessApi.changeGroupPermissions(id, codes, action)
    catalogVersion.value++
    ElMessage.success(action === 'add' ? '所选权限已添加' : '所选权限已移除')
    await loadGroups()
    if (groupsError.value) ElMessage.warning('权限已调整，权限组列表刷新失败，请重试')
  } catch (error) {
    actionError.value = errorMessage(error)
  } finally {
    busy.value = false
  }
}
async function grant(accountId: string, roleIds: number[]) {
  const unitId = selectedUnitId.value
  if (
    busy.value ||
    !unitId ||
    !selectedUnit.value?.is_owner ||
    !isAccountId(accountId) ||
    !roleIds.length
  )
    return
  busy.value = true
  actionError.value = ''
  try {
    await accessApi.grantUnitRoles(unitId, accountId, roleIds)
    grantVersion.value++
    ElMessage.success('协作授权已添加')
    await loadGrants()
  } catch (error) {
    actionError.value = errorMessage(error)
  } finally {
    busy.value = false
  }
}
async function revoke() {
  const target = revokingGrant.value
  if (
    busy.value ||
    !target ||
    !selectedUnit.value?.is_owner ||
    selectedUnitId.value !== target.unit_id
  )
    return
  busy.value = true
  actionError.value = ''
  try {
    await accessApi.revokeUnitRoles(target.unit_id, target.account_id, [
      target.role_id,
    ])
    revokingGrant.value = null
    ElMessage.success('协作授权已撤销')
    await loadGrants()
  } catch (error) {
    actionError.value = errorMessage(error)
  } finally {
    busy.value = false
  }
}
function filterGrants(accountId: string, status: string) {
  grantAccountFilter.value = accountId
  grantStatusFilter.value = status
  grantOffset.value = 0
  void loadGrants()
}
function changePage(offset: number) {
  grantOffset.value = offset
  void loadGrants()
}
function closeConfirmation() {
  deletingRole.value = null
  revokingGrant.value = null
}
function selectTab(tab: string, event: MouseEvent) {
  animateTabs.value = event.detail > 0
  activeTab.value = tab
}
function navigateTabs(event: KeyboardEvent) {
  if (
    busy.value ||
    !['ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(event.key)
  )
    return
  const tabs = ['roles', 'grants', 'catalog']
  const index = tabs.indexOf(activeTab.value)
  const next =
    event.key === 'Home'
      ? 0
      : event.key === 'End'
        ? 2
        : (index + (event.key === 'ArrowRight' ? 1 : -1) + tabs.length) %
          tabs.length
  event.preventDefault()
  animateTabs.value = false
  activeTab.value = tabs[next] ?? 'roles'
  const container = event.currentTarget
  if (container instanceof HTMLElement) {
    container.querySelectorAll<HTMLButtonElement>('[role="tab"]')[next]?.focus()
  }
}
watch(selectedUnitId, () => {
  grantOffset.value = 0
  grantAccountFilter.value = ''
  grantStatusFilter.value = 'active'
  revokingGrant.value = null
  void loadGrants()
})
watch(activeTab, () => {
  actionError.value = ''
})
onMounted(() => {
  void loadRoles()
  void loadGroups()
  void loadPermissions()
  void initializeUnits()
})
onUnmounted(() => {
  disposed = true
})
</script>

<template>
  <div class="access-view">
    <PageHeader title="协作与权限" description="让合适的人，访问合适的空间。">
      <template v-if="isSuperAdmin" #actions>
        <span class="badge" title="当前账号可管理全局权限目录">
          <AppIcon name="shield" :size="15" />
          超级管理员
        </span>
      </template>
    </PageHeader>
    <div class="access-summary">
      <span class="summary-mark" aria-hidden="true"
        ><AppIcon name="shield" :size="24"
      /></span>
      <div>
        <strong>清晰的边界，安心的协作</strong>
        <p>
          自己的单元默认拥有完整访问权限。通过角色与权限组，为协作者分配工作范围。
        </p>
      </div>
    </div>
    <div
      class="segmented-control"
      role="tablist"
      aria-label="权限管理区域"
      @keydown="navigateTabs"
    >
      <button
        id="roles-tab"
        role="tab"
        :aria-selected="activeTab === 'roles'"
        :tabindex="activeTab === 'roles' ? 0 : -1"
        aria-controls="roles-panel"
        :class="{ active: activeTab === 'roles' }"
        :disabled="busy"
        @click="selectTab('roles', $event)"
      >
        我的角色
      </button>
      <button
        id="grants-tab"
        role="tab"
        :aria-selected="activeTab === 'grants'"
        :tabindex="activeTab === 'grants' ? 0 : -1"
        aria-controls="grants-panel"
        :class="{ active: activeTab === 'grants' }"
        :disabled="busy"
        @click="selectTab('grants', $event)"
      >
        单元协作
      </button>
      <button
        id="catalog-tab"
        role="tab"
        :aria-selected="activeTab === 'catalog'"
        :tabindex="activeTab === 'catalog' ? 0 : -1"
        aria-controls="catalog-panel"
        :class="{ active: activeTab === 'catalog' }"
        :disabled="busy"
        @click="selectTab('catalog', $event)"
      >
        权限目录
      </button>
    </div>
    <div
      v-show="activeTab === 'roles'"
      id="roles-panel"
      class="access-panel"
      :class="{ 'access-panel-animated': animateTabs }"
      :inert="activeTab !== 'roles'"
      :aria-hidden="activeTab !== 'roles'"
      role="tabpanel"
      aria-labelledby="roles-tab"
    >
      <RolesPanel
        :roles="roles"
        :groups="groups"
        :loading="rolesLoading"
        :error="rolesError"
        :busy="busy"
        :action-error="actionError"
        :saved-version="roleVersion"
        @retry="loadRoles"
        @save="saveRole"
        @bind="bindRoleGroups"
        @remove="
          (role) => {
            actionError = ''
            deletingRole = role
          }
        "
        @clear-error="actionError = ''"
      />
      <p v-if="groupsError" class="catalog-error" role="alert">
        权限组暂时无法加载：{{ groupsError }}
        <button class="btn btn-secondary" @click="loadGroups">重试</button>
      </p>
    </div>
    <div
      v-show="activeTab === 'grants'"
      id="grants-panel"
      class="access-panel"
      :class="{ 'access-panel-animated': animateTabs }"
      :inert="activeTab !== 'grants'"
      :aria-hidden="activeTab !== 'grants'"
      role="tabpanel"
      aria-labelledby="grants-tab"
    >
      <UnitGrantsPanel
        :units="units"
        :unit-id="selectedUnitId"
        :roles="roles"
        :grants="grants"
        :loading="grantsLoading"
        :error="grantsError"
        :busy="busy"
        :action-error="actionError"
        :saved-version="grantVersion"
        :offset="grantOffset"
        :has-next="grantHasNext"
        @select-unit="(id) => (selectedUnitId = id)"
        @retry="initializeUnits"
        @filter="filterGrants"
        @page="changePage"
        @grant="grant"
        @revoke="
          (item) => {
            actionError = ''
            revokingGrant = item
          }
        "
        @clear-error="actionError = ''"
      />
    </div>
    <div
      v-show="activeTab === 'catalog'"
      id="catalog-panel"
      class="access-panel"
      :class="{ 'access-panel-animated': animateTabs }"
      :inert="activeTab !== 'catalog'"
      :aria-hidden="activeTab !== 'catalog'"
      role="tabpanel"
      aria-labelledby="catalog-tab"
    >
      <PermissionCatalog
        :groups="groups"
        :permissions="permissions"
        :groups-loading="groupsLoading"
        :groups-error="groupsError"
        :permissions-loading="permissionsLoading"
        :permissions-error="permissionsError"
        :busy="busy"
        :action-error="actionError"
        :saved-version="catalogVersion"
        @retry-groups="loadGroups"
        @retry-permissions="loadPermissions"
        @create-group="createGroup"
        @bind="bindGroupPermissions"
        @clear-error="actionError = ''"
      />
    </div>

    <el-dialog
      :model-value="!!deletingRole || !!revokingGrant"
      :title="deletingRole ? '删除角色' : '撤销协作授权'"
      width="460px"
      :close-on-click-modal="false"
      :show-close="!busy"
      :close-on-press-escape="!busy"
      @update:model-value="
        (value) => {
          if (!value) {
            deletingRole = null
            revokingGrant = null
          }
        }
      "
    >
      <p v-if="deletingRole">
        确认删除“{{ deletingRole.name }}”？使用此角色的协作者将失去对应权限。
      </p>
      <p v-else>
        撤销后，该账号将失去“{{
          revokingGrant?.role?.name ?? `角色 #${revokingGrant?.role_id}`
        }}”在此单元的访问权限。
      </p>
      <p v-if="actionError" class="form-error" role="alert">
        {{ actionError }}
      </p>
      <template #footer>
        <button
          class="btn btn-secondary"
          :disabled="busy"
          @click="closeConfirmation"
        >
          取消
        </button>
        <button
          class="btn btn-danger"
          :disabled="busy"
          @click="deletingRole ? removeRole() : revoke()"
        >
          {{ busy ? '正在处理…' : '确认' }}
        </button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.access-view {
  display: grid;
  gap: 1.5rem;
}
.access-panel {
  min-width: 0;
}
.access-panel-animated {
  transition: opacity var(--motion-fast) var(--ease-out);
}
@starting-style {
  .access-panel-animated {
    opacity: 0;
  }
}
.access-summary {
  display: flex;
  align-items: center;
  gap: 1.1rem;
  padding: 1.5rem 1.75rem;
  background: linear-gradient(120deg, var(--accent-soft), #f8fafc);
  border-radius: 1.25rem;
}
.summary-mark {
  display: grid;
  place-items: center;
  width: 3rem;
  height: 3rem;
  border-radius: 1rem;
  background: white;
  color: var(--accent);
  font-size: 1.75rem;
  flex: 0 0 auto;
}
.access-summary strong {
  font-size: 1rem;
  font-weight: 600;
}
.access-summary p {
  margin: 0.4rem 0 0;
  color: var(--text-secondary, #68717d);
  font-size: 0.86rem;
  line-height: 1.65;
}
.segmented-control {
  display: flex;
  gap: 0.25rem;
  padding: 0.3rem;
  border-radius: 0.85rem;
  background: var(--surface-secondary, #eceef1);
  width: fit-content;
  max-width: 100%;
}
.segmented-control button {
  border: 0;
  background: none;
  border-radius: 0.6rem;
  padding: 0.65rem 1.25rem;
  color: var(--text-secondary, #68717d);
  font: inherit;
  font-size: 0.87rem;
  cursor: pointer;
}
.segmented-control button.active {
  background: var(--surface, white);
  color: var(--text-primary);
  box-shadow: 0 2px 6px #0000000c;
}
.segmented-control button:active {
  transform: scale(0.97);
}
.form-error,
.catalog-error {
  color: var(--danger);
  font-size: 0.875rem;
}
.catalog-error {
  display: flex;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
}
@media (max-width: 640px) {
  .access-summary {
    padding: 1.25rem;
    align-items: flex-start;
  }
  .summary-mark {
    display: none;
  }
  .segmented-control {
    width: 100%;
  }
  .segmented-control button {
    flex: 1;
    padding: 0.65rem 0.6rem;
  }
}
@media (prefers-reduced-motion: reduce) {
  .access-panel-animated {
    transition-duration: 100ms !important;
  }
}
</style>
