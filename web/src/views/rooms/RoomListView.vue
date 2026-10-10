<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { createRoom, deleteRoom, listRooms, updateRoom } from '@/api/rooms'
import { listRoomTypes } from '@/api/roomTypes'
import { listMeterItems } from '@/api/meterItems'
import { useUnits } from '@/stores/units'
import { errorMessage, formatMoney } from '@/utils/format'
import type { Room, RoomUpdateInput } from '@/types/rooms'
import type { RoomType } from '@/types/roomTypes'
import type { Unit, UnitMeterItem } from '@/types/units'
import PageHeader from '@/components/common/PageHeader.vue'
import AsyncState from '@/components/common/AsyncState.vue'
import AppIcon from '@/components/common/AppIcon.vue'
import RoomForm from '@/components/rooms/RoomForm.vue'

const {
  units,
  selectedUnitId,
  selectedUnit,
  loading: unitsLoading,
  error: unitsError,
  loadUnits,
} = useUnits()
const rooms = ref<Room[]>([])
const roomTypes = ref<RoomType[]>([])
const meterItems = ref<UnitMeterItem[]>([])
const loading = ref(false)
const error = ref('')
const dictionaryError = ref('')
const statusFilter = ref<'all' | 'vacant' | 'rented'>('all')
const typeFilter = ref<number>()
const page = ref(0)
const pageSize = 20
const hasNext = ref(false)
let requestVersion = 0
let dictionaryVersion = 0
let disposed = false
const typeNames = computed(
  () => new Map(roomTypes.value.map((type) => [type.id, type.name])),
)
const formVisible = ref(false)
const editing = ref<Room>()
const formUnit = ref<Unit>()
const formKey = ref(0)
const saving = ref(false)
const formError = ref('')
const deleting = ref<number>()

async function loadRooms() {
  const version = ++requestVersion
  const unitId = selectedUnitId.value
  if (!unitId) {
    rooms.value = []
    loading.value = false
    error.value = ''
    return
  }
  loading.value = true
  error.value = ''
  try {
    const result = await listRooms(unitId, {
      limit: pageSize + 1,
      offset: page.value * pageSize,
      room_type_id: typeFilter.value,
      is_active:
        statusFilter.value === 'all'
          ? undefined
          : statusFilter.value === 'vacant',
    })
    if (version !== requestVersion || disposed) return
    rooms.value = result.slice(0, pageSize)
    hasNext.value = result.length > pageSize
  } catch (err) {
    if (version !== requestVersion || disposed) return
    error.value = errorMessage(err)
  } finally {
    if (version === requestVersion && !disposed) loading.value = false
  }
}

async function loadDictionaries() {
  const version = ++dictionaryVersion
  dictionaryError.value = ''
  const results = await Promise.allSettled([listRoomTypes(), listMeterItems()])
  if (disposed || version !== dictionaryVersion) return
  const [types, meters] = results
  if (types.status === 'fulfilled') roomTypes.value = types.value
  if (meters.status === 'fulfilled') meterItems.value = meters.value
  const failures = results.filter((result) => result.status === 'rejected')
  if (failures.length)
    dictionaryError.value =
      '房型或收费项目加载失败，已有房间仍可查看。请重试加载。'
}

watch(
  [selectedUnitId, statusFilter, typeFilter],
  () => {
    page.value = 0
    void loadRooms()
  },
  { immediate: true },
)
watch(selectedUnitId, () => {
  typeFilter.value = undefined
  if (!saving.value) formVisible.value = false
})
onMounted(loadDictionaries)
onUnmounted(() => {
  disposed = true
  requestVersion += 1
})

function changePage(direction: number) {
  if (loading.value || page.value + direction < 0) return
  page.value += direction
  void loadRooms()
}

function openForm(room?: Room) {
  if (!selectedUnit.value) return
  editing.value = room
  formUnit.value = selectedUnit.value
  formError.value = ''
  formKey.value += 1
  formVisible.value = true
}

async function saveRoom(input: RoomUpdateInput) {
  if (saving.value || !formUnit.value) return
  saving.value = true
  formError.value = ''
  const roomId = editing.value?.id
  const unitId = formUnit.value.id
  try {
    if (roomId) await updateRoom(roomId, input)
    else {
      const { is_active: _status, unit_id: _unitId, ...createInput } = input
      await createRoom(unitId, createInput)
    }
  } catch (err) {
    formError.value = errorMessage(err)
    return
  } finally {
    saving.value = false
  }
  formVisible.value = false
  ElMessage.success('房间已保存')
  await Promise.all([loadRooms(), loadUnits()])
  if (error.value || unitsError.value)
    ElMessage.warning('房间已保存，部分数据刷新失败，请重试')
}

async function removeRoom(room: Room) {
  if (deleting.value !== undefined) return
  deleting.value = room.id
  try {
    await ElMessageBox.confirm(
      `确认删除 ${room.room_no} 房间？删除后无法恢复。`,
      '删除房间',
      {
        confirmButtonText: '删除房间',
        cancelButtonText: '保留房间',
        type: 'warning',
      },
    )
  } catch {
    deleting.value = undefined
    return
  }
  try {
    await deleteRoom(room.id)
  } catch (err) {
    ElMessage.error(errorMessage(err))
    return
  } finally {
    deleting.value = undefined
  }
  ElMessage.success('房间已删除')
  if (rooms.value.length === 1 && page.value > 0) page.value -= 1
  await Promise.all([loadRooms(), loadUnits()])
  if (error.value || unitsError.value)
    ElMessage.warning('房间已删除，部分数据刷新失败，请重试')
}
</script>

<template>
  <div>
    <PageHeader
      title="房间管理"
      :description="
        selectedUnit
          ? `${selectedUnit.name} · 安排好每一间房`
          : '选择楼栋，查看房间与出租状态。'
      "
    >
      <template #actions>
        <el-button type="primary" :disabled="!selectedUnit" @click="openForm()"
          ><AppIcon name="plus" />新增房间</el-button
        >
      </template>
    </PageHeader>

    <div v-if="selectedUnit" class="room-summary panel">
      <div class="summary-building">
        <AppIcon name="building" /><span>{{ selectedUnit.name }}</span>
      </div>
      <div>
        <strong>{{ selectedUnit.room_count }}</strong
        ><span>间房</span>
      </div>
      <div>
        <i class="status-dot rented" /><strong>{{
          selectedUnit.rented_count
        }}</strong
        ><span>已出租</span>
      </div>
      <div>
        <i class="status-dot vacant" /><strong>{{
          selectedUnit.room_count - selectedUnit.rented_count
        }}</strong
        ><span>空置</span>
      </div>
    </div>

    <div class="toolbar room-toolbar">
      <el-radio-group v-model="statusFilter" aria-label="出租状态筛选">
        <el-radio-button value="all">全部房间</el-radio-button>
        <el-radio-button value="rented">已出租</el-radio-button>
        <el-radio-button value="vacant">空置</el-radio-button>
      </el-radio-group>
      <el-select
        v-if="selectedUnit?.is_owner"
        v-model="typeFilter"
        class="type-filter"
        clearable
        placeholder="全部房型"
        aria-label="按房型筛选"
      >
        <el-option
          v-for="type in roomTypes"
          :key="type.id"
          :label="type.name"
          :value="type.id"
        />
      </el-select>
      <el-button :loading="loading" aria-label="刷新房间" @click="loadRooms"
        ><AppIcon name="refresh"
      /></el-button>
    </div>

    <div v-if="dictionaryError" class="dictionary-error">
      <el-alert
        :title="dictionaryError"
        type="warning"
        :closable="false"
        show-icon
      />
      <el-button text type="primary" @click="loadDictionaries"
        >重新加载房型与收费项目</el-button
      >
    </div>

    <AsyncState
      :loading="loading || unitsLoading"
      :error="unitsError || error"
      :empty="rooms.length === 0 || !selectedUnit"
      :empty-title="
        !selectedUnit ? '请先添加或选择楼栋' : '这里还没有匹配的房间'
      "
      :empty-description="
        !selectedUnit
          ? '在楼栋管理中添加第一栋房子，开始管理房间。'
          : '新增房间，或切换出租状态和房型筛选。'
      "
      @retry="unitsError ? loadUnits() : loadRooms()"
    >
      <div class="room-grid">
        <article v-for="room in rooms" :key="room.id" class="room-card panel">
          <div class="room-card-header">
            <div class="room-number">
              <AppIcon name="grid" />
              <h2>{{ room.room_no }}</h2>
            </div>
            <span
              class="room-status"
              :class="room.is_active ? 'vacant' : 'rented'"
              >{{ room.is_active ? '空置' : '已出租' }}</span
            >
          </div>
          <p class="room-spec">
            {{ room.floor }} 层<span>·</span>{{ room.size }} ㎡<span>·</span
            >{{
              typeNames.get(room.room_type_id) || `房型 #${room.room_type_id}`
            }}
          </p>
          <p class="room-rent">
            {{ formatMoney(room.price) }}<span>/ 月</span>
          </p>
          <div class="meter-tags">
            <span v-for="item in room.meter_items" :key="item.id">{{
              item.code
            }}</span>
            <span v-if="!room.meter_items.length" class="no-meter"
              >暂无收费项目</span
            >
          </div>
          <div class="room-card-actions">
            <el-button text type="primary" @click="openForm(room)"
              ><AppIcon name="edit" />编辑房间</el-button
            >
            <el-button
              text
              type="danger"
              :aria-label="`删除 ${room.room_no} 房间`"
              :loading="deleting === room.id"
              :disabled="deleting !== undefined"
              @click="removeRoom(room)"
              ><AppIcon name="trash"
            /></el-button>
          </div>
        </article>
      </div>
    </AsyncState>

    <div
      v-if="selectedUnit && !error && (rooms.length || page > 0)"
      class="room-pagination"
    >
      <span class="muted"
        >第 {{ page + 1 }} 页 · 本页 {{ rooms.length }} 间</span
      >
      <div>
        <el-button :disabled="page === 0 || loading" @click="changePage(-1)"
          >上一页</el-button
        >
        <el-button :disabled="!hasNext || loading" @click="changePage(1)"
          >下一页</el-button
        >
      </div>
    </div>

    <el-dialog
      v-model="formVisible"
      :title="editing ? `编辑 ${editing.room_no} 房间` : '新增房间'"
      width="560px"
      destroy-on-close
      :close-on-click-modal="false"
      :show-close="!saving"
      :close-on-press-escape="!saving"
    >
      <RoomForm
        v-if="formVisible && formUnit"
        :key="formKey"
        :initial="editing"
        :unit="formUnit"
        :units="units"
        :room-types="formUnit.is_owner ? roomTypes : []"
        :meter-items="formUnit.is_owner ? meterItems : []"
        :saving="saving"
        :error="formError"
        @submit="saveRoom"
        @cancel="formVisible = false"
      />
    </el-dialog>
  </div>
</template>

<style scoped>
.room-summary {
  display: flex;
  align-items: center;
  gap: 2rem;
  padding: 1.3rem 1.5rem;
  margin-bottom: 1.6rem;
  flex-wrap: wrap;
}
.room-summary > div {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}
.room-summary strong {
  font-size: 1.3rem;
  font-weight: 600;
}
.room-summary span {
  color: var(--text-secondary, #727780);
  font-size: 0.8rem;
}
.room-summary .summary-building {
  margin-right: auto;
  gap: 0.8rem;
  min-width: 0;
}
.summary-building span {
  color: var(--text-primary, #17202a);
  font-weight: 600;
  overflow-wrap: anywhere;
}
.summary-building :deep(svg) {
  color: #528578;
  width: 1.3rem;
  height: 1.3rem;
  flex-shrink: 0;
}
.status-dot {
  width: 0.4rem;
  height: 0.4rem;
  border-radius: 50%;
}
.status-dot.rented {
  background: #408971;
}
.status-dot.vacant {
  background: #c49746;
}
.room-toolbar {
  margin-bottom: 1.3rem;
}
.type-filter {
  max-width: 10rem;
  margin-left: auto;
}
.room-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1.1rem;
}
.room-card {
  padding: 1.3rem;
}
.room-card-header {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: center;
}
.room-number {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  min-width: 0;
}
.room-number :deep(svg) {
  color: #7e918a;
  flex-shrink: 0;
}
.room-number h2 {
  font-size: 1.3rem;
  margin: 0;
  font-weight: 600;
  letter-spacing: -0.02em;
  overflow-wrap: anywhere;
}
.room-status {
  font-size: 0.68rem;
  padding: 0.3rem 0.6rem;
  border-radius: 0.45rem;
  white-space: nowrap;
}
.room-status.rented {
  color: #34735e;
  background: #edf5f0;
}
.room-status.vacant {
  color: #94692d;
  background: #fbf3e5;
}
.room-spec {
  color: var(--text-secondary, #727780);
  font-size: 0.78rem;
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  margin: 0.7rem 0 1.4rem;
}
.room-rent {
  font-size: 1.5rem;
  font-weight: 600;
  margin: 0;
  letter-spacing: -0.04em;
}
.room-rent span {
  font-size: 0.73rem;
  font-weight: 400;
  margin-left: 0.5rem;
  color: var(--text-secondary, #727780);
  letter-spacing: 0;
}
.meter-tags {
  display: flex;
  gap: 0.4rem;
  margin-top: 0.9rem;
  flex-wrap: wrap;
  min-height: 1.5rem;
}
.meter-tags span {
  padding: 0.15rem 0.45rem;
  font-size: 0.68rem;
  color: #75827c;
  background: #f4f6f4;
  border-radius: 0.35rem;
}
.meter-tags .no-meter {
  background: none;
  padding-left: 0;
  color: #9b9e9c;
}
.room-card-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 1.3rem;
  padding-top: 0.7rem;
  border-top: 1px solid var(--border, #eff1ee);
}
.room-card-actions :deep(.el-button) {
  margin-left: 0;
  padding-inline: 0.45rem;
}
.room-pagination {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-top: 1.5rem;
}
.room-pagination > span {
  font-size: 0.8rem;
}
.dictionary-error {
  margin-bottom: 1rem;
}
@media (max-width: 1200px) {
  .room-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
@media (max-width: 680px) {
  .room-grid {
    grid-template-columns: 1fr;
  }
  .room-summary {
    gap: 1rem;
    padding: 1rem;
  }
  .room-summary .summary-building {
    flex-basis: 100%;
  }
  .room-summary strong {
    font-size: 1.1rem;
  }
  .type-filter {
    margin-left: 0;
    flex: 1;
    max-width: none;
    min-width: 8rem;
  }
  .room-pagination {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
