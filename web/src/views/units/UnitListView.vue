<script setup lang="ts">
import { computed, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { createUnit, deleteUnit, updateUnit } from '@/api/units'
import { listMeterItems } from '@/api/meterItems'
import { useUnits } from '@/stores/units'
import { errorMessage } from '@/utils/format'
import type { Unit, UnitInput, UnitMeterItem, UnitScope } from '@/types/units'
import PageHeader from '@/components/common/PageHeader.vue'
import AsyncState from '@/components/common/AsyncState.vue'
import AppIcon from '@/components/common/AppIcon.vue'
import UnitForm from '@/components/units/UnitForm.vue'

const router = useRouter()
const { units, loading, error, loadUnits, setSelectedUnit } = useUnits()
const scope = ref<UnitScope>('all')
const search = ref('')
const visibleUnits = computed(() =>
  units.value.filter((unit) => {
    const matchesScope =
      scope.value === 'all' || unit.access_type === scope.value
    const text =
      `${unit.name} ${unit.province ?? ''} ${unit.city ?? ''} ${unit.address ?? ''}`.toLowerCase()
    return matchesScope && text.includes(search.value.trim().toLowerCase())
  }),
)
const totalRooms = computed(() =>
  units.value.some((unit) => unit.room_count === undefined)
    ? undefined
    : units.value.reduce((total, unit) => total + (unit.room_count ?? 0), 0),
)
const rentedRooms = computed(() =>
  units.value.some((unit) => unit.rented_count === undefined)
    ? undefined
    : units.value.reduce((total, unit) => total + (unit.rented_count ?? 0), 0),
)
const formVisible = ref(false)
const editing = ref<Unit>()
const formKey = ref(0)
const saving = ref(false)
const formError = ref('')
const meterItems = ref<UnitMeterItem[]>([])
const meterError = ref('')
const meterLoading = ref(false)
const deleting = ref<number>()
let meterVersion = 0
onUnmounted(() => {
  meterVersion += 1
})

async function openForm(unit?: Unit) {
  editing.value = unit
  formError.value = ''
  meterError.value = ''
  meterItems.value = []
  formKey.value += 1
  formVisible.value = true
  await loadMeterOptions()
}

async function loadMeterOptions() {
  const version = ++meterVersion
  meterLoading.value = true
  meterError.value = ''
  try {
    const items = await listMeterItems({ unit_id: editing.value?.id })
    if (version !== meterVersion) return
    meterItems.value = items
  } catch (error) {
    if (version === meterVersion)
      meterError.value = `收费项目加载失败：${errorMessage(error)}`
  } finally {
    if (version === meterVersion) meterLoading.value = false
  }
}

async function saveUnit(input: UnitInput) {
  if (saving.value) return
  saving.value = true
  formError.value = ''
  try {
    if (editing.value) await updateUnit(editing.value.id, input)
    else await createUnit(input)
    formVisible.value = false
    ElMessage.success('楼栋已保存')
    await loadUnits()
    if (error.value) ElMessage.warning('楼栋已保存，列表刷新失败，请重试')
  } catch (error) {
    formError.value = errorMessage(error)
  } finally {
    saving.value = false
  }
}

async function removeUnit(unit: Unit) {
  if (!unit.is_owner || deleting.value !== undefined) return
  deleting.value = unit.id
  try {
    await ElMessageBox.confirm(
      `删除“${unit.name}”后无法恢复。请先确认楼栋内的房间已妥善处理。`,
      '删除楼栋',
      {
        confirmButtonText: '删除楼栋',
        cancelButtonText: '保留楼栋',
        type: 'warning',
      },
    )
  } catch {
    deleting.value = undefined
    return
  }
  try {
    await deleteUnit(unit.id)
    ElMessage.success('楼栋已删除')
    await loadUnits()
    if (error.value) ElMessage.warning('楼栋已删除，列表刷新失败，请重试')
  } catch (error) {
    ElMessage.error(errorMessage(error))
  } finally {
    deleting.value = undefined
  }
}

function openRooms(unit: Unit) {
  setSelectedUnit(unit.id)
  router.push('/rooms')
}
</script>

<template>
  <div class="unit-page">
    <PageHeader title="楼栋管理" description="每一栋房子，都井井有条。">
      <template #actions>
        <el-button type="primary" @click="openForm()"
          ><AppIcon name="plus" />新增楼栋</el-button
        >
      </template>
    </PageHeader>

    <div class="portfolio-strip panel" aria-label="楼栋概览">
      <div>
        <span class="muted">管理楼栋</span
        ><strong>{{ units.length }}<small>栋</small></strong>
      </div>
      <div>
        <span class="muted">房间总数</span
        ><strong>{{ totalRooms ?? '—' }}<small>间</small></strong>
      </div>
      <div>
        <span class="muted">已出租</span
        ><strong>{{ rentedRooms ?? '—' }}<small>间</small></strong>
      </div>
      <div>
        <span class="muted">空置房间</span
        ><strong>{{ totalRooms !== undefined && rentedRooms !== undefined ? totalRooms - rentedRooms : '—' }}<small>间</small></strong>
      </div>
    </div>

    <div class="toolbar unit-toolbar">
      <div class="scope-switch" role="group" aria-label="楼栋范围">
        <button
          v-for="item in [
            { value: 'all', label: '全部楼栋' },
            { value: 'owned', label: '我拥有的' },
            { value: 'granted', label: '授权给我的' },
          ]"
          :key="item.value"
          type="button"
          :aria-pressed="scope === item.value"
          @click="scope = item.value as UnitScope"
        >
          {{ item.label }}
        </button>
      </div>
      <el-input
        v-model="search"
        class="unit-search"
        clearable
        placeholder="搜索楼栋或地址"
        aria-label="搜索楼栋或地址"
      >
        <template #prefix><AppIcon name="search" /></template>
      </el-input>
      <el-button aria-label="刷新楼栋" :loading="loading" @click="loadUnits"
        ><AppIcon name="refresh"
      /></el-button>
    </div>

    <AsyncState
      :loading="loading"
      :error="error"
      :empty="visibleUnits.length === 0"
      :empty-title="units.length ? '没有找到匹配的楼栋' : '从第一栋房子开始'"
      :empty-description="
        units.length
          ? '试试其他名称或切换楼栋范围。'
          : '添加楼栋，再为它创建房间和收费项目。'
      "
      @retry="loadUnits"
    >
      <div class="unit-grid">
        <article
          v-for="unit in visibleUnits"
          :key="unit.id"
          class="unit-card panel"
        >
          <div class="unit-card-top">
            <div class="building-mark"><AppIcon name="building" /></div>
            <span
              class="badge"
              :class="unit.is_owner ? 'owned-badge' : 'granted-badge'"
              >{{ unit.is_owner ? '我拥有的' : '协作管理' }}</span
            >
          </div>
          <h2>{{ unit.name }}</h2>
          <p v-if="unit.address !== undefined" class="unit-address">
            {{ unit.province }} · {{ unit.city }}<br />{{ unit.address }}
          </p>
          <p v-else class="unit-address">详细资料未授权</p>
          <div class="unit-counts">
            <div>
              <strong>{{ unit.room_count ?? '—' }}</strong
              ><span>全部房间</span>
            </div>
            <div>
              <strong>{{ unit.rented_count ?? '—' }}</strong
              ><span>已出租</span>
            </div>
            <div>
              <strong>{{ unit.room_count !== undefined && unit.rented_count !== undefined ? unit.room_count - unit.rented_count : '—' }}</strong
              ><span>空置</span>
            </div>
          </div>
          <div
            v-if="unit.room_count !== undefined && unit.rented_count !== undefined"
            class="occupancy-track"
            :aria-label="`已出租 ${unit.rented_count} 间，共 ${unit.room_count} 间`"
          >
            <span
              :style="{
                width: `${unit.room_count ? (unit.rented_count / unit.room_count) * 100 : 0}%`,
              }"
            />
          </div>
          <div v-if="unit.json_extend && unit.default_meter_items" class="unit-meta">
            <span>每月 {{ unit.json_extend.meter_reading_day }} 日抄表</span
            ><span>{{ unit.default_meter_items.length }} 个收费项目</span>
          </div>
          <div class="unit-actions">
            <el-button type="primary" plain @click="openRooms(unit)"
              >查看房间<AppIcon name="arrow-right"
            /></el-button>
            <div>
              <el-button
                v-if="unit.address !== undefined"
                text
                :aria-label="`编辑${unit.name}`"
                @click="openForm(unit)"
                ><AppIcon name="edit"
              /></el-button>
              <el-button
                v-if="unit.is_owner"
                text
                type="danger"
                :aria-label="`删除${unit.name}`"
                :loading="deleting === unit.id"
                :disabled="deleting !== undefined"
                @click="removeUnit(unit)"
                ><AppIcon name="trash"
              /></el-button>
            </div>
          </div>
        </article>
      </div>
    </AsyncState>

    <el-dialog
      v-model="formVisible"
      :title="editing ? '编辑楼栋' : '新增楼栋'"
      width="560px"
      class="unit-editor"
      destroy-on-close
      :close-on-click-modal="false"
      :close-on-press-escape="!saving"
      :show-close="!saving"
    >
      <div v-if="meterError" class="catalog-error">
        <el-alert
          :title="meterError"
          type="warning"
          :closable="false"
          show-icon
        />
        <el-button
          text
          type="primary"
          :loading="meterLoading"
          @click="loadMeterOptions"
          >重新加载收费项目</el-button
        >
      </div>
      <UnitForm
        v-if="formVisible"
        :key="formKey"
        :initial="editing"
        :meter-items="meterItems"
        :saving="saving"
        :error="formError"
        @submit="saveUnit"
        @cancel="formVisible = false"
      />
    </el-dialog>
  </div>
</template>

<style scoped>
.portfolio-strip {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  padding: 1.3rem 1.7rem;
  gap: 1rem;
  margin-bottom: 1.7rem;
}
.portfolio-strip > div {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}
.portfolio-strip .muted {
  font-size: 0.78rem;
}
.portfolio-strip strong {
  font-size: 1.75rem;
  font-weight: 600;
  letter-spacing: -0.04em;
}
.portfolio-strip small {
  font-size: 0.75rem;
  font-weight: 400;
  color: var(--text-secondary, #727780);
  margin-left: 0.5rem;
}
.unit-toolbar {
  margin-bottom: 1.4rem;
}
.scope-switch {
  display: flex;
  padding: 0.25rem;
  border-radius: 0.7rem;
  background: var(--surface-muted, #eceef1);
}
.scope-switch button {
  border: 0;
  background: none;
  border-radius: 0.5rem;
  padding: 0.55rem 0.8rem;
  color: var(--text-secondary, #727780);
  cursor: pointer;
  font: inherit;
  font-size: 0.8rem;
  white-space: nowrap;
}
.scope-switch button[aria-pressed='true'] {
  background: white;
  box-shadow: 0 1px 5px #19283712;
  color: var(--text-primary, #17202a);
  font-weight: 600;
}
.unit-search {
  max-width: 17rem;
  margin-left: auto;
}
.unit-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1.1rem;
}
.unit-card {
  padding: 1.4rem;
}
.unit-card-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.4rem;
}
.building-mark {
  width: 3rem;
  height: 3rem;
  border-radius: 0.9rem;
  display: grid;
  place-items: center;
  color: #368571;
  background: #e9f3ef;
}
.building-mark :deep(svg) {
  width: 1.45rem;
  height: 1.45rem;
}
.owned-badge {
  background: #edf5f0;
  color: #3f7d64;
}
.granted-badge {
  background: #edf2fa;
  color: #59749e;
}
.unit-card h2 {
  margin: 0 0 0.55rem;
  font-size: 1.05rem;
  font-weight: 600;
  overflow-wrap: anywhere;
}
.unit-address {
  font-size: 0.79rem;
  color: var(--text-secondary, #727780);
  line-height: 1.75;
  margin: 0 0 1.45rem;
  overflow-wrap: anywhere;
}
.unit-counts {
  display: flex;
  gap: 2rem;
  margin-bottom: 1rem;
}
.unit-counts > div {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
}
.unit-counts strong {
  font-size: 1.3rem;
  font-weight: 600;
}
.unit-counts span,
.unit-meta {
  font-size: 0.7rem;
  color: var(--text-secondary, #727780);
}
.occupancy-track {
  height: 0.3rem;
  background: #edf0ed;
  border-radius: 0.3rem;
  overflow: hidden;
}
.occupancy-track span {
  display: block;
  height: 100%;
  background: #6aa791;
  border-radius: inherit;
}
.unit-meta {
  display: flex;
  justify-content: space-between;
  gap: 0.5rem;
  margin-top: 0.75rem;
}
.unit-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.5rem;
  margin-top: 1.6rem;
}
.unit-actions > div {
  display: flex;
}
.unit-actions :deep(.el-button + .el-button) {
  margin-left: 0;
}
.catalog-error {
  margin-bottom: 1rem;
}
@media (max-width: 1200px) {
  .unit-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
@media (max-width: 680px) {
  .portfolio-strip {
    grid-template-columns: repeat(2, 1fr);
    padding: 1.1rem;
  }
  .unit-grid {
    grid-template-columns: 1fr;
  }
  .unit-search {
    max-width: none;
    margin-left: 0;
    flex: 1;
    min-width: 12rem;
  }
  .scope-switch {
    width: 100%;
  }
  .scope-switch button {
    flex: 1;
    padding-inline: 0.5rem;
  }
}
</style>
