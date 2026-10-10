<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { getRoomReadings, getUnitReadings, saveRoomReadings } from '@/api/readings'
import AsyncState from '@/components/common/AsyncState.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import RoomReadingCard from '@/components/readings/RoomReadingCard.vue'
import RoomReadingForm from '@/components/readings/RoomReadingForm.vue'
import { useUnits } from '@/stores/units'
import type { RoomMonthReadings, SaveReadingsInput, UnitMonthReadings } from '@/types/readings'
import { currentMonth, errorMessage, formatMoney } from '@/utils/format'

const { selectedUnitId } = useUnits()
const month = ref(currentMonth())
const result = ref<UnitMonthReadings>()
const loading = ref(false)
const error = ref('')
const search = ref('')
const onlyMissing = ref(false)
const page = ref(1)
const pageSize = 12
const editorOpen = ref(false)
const editingRoom = ref<RoomMonthReadings>()
const editorLoading = ref(false)
const editorError = ref('')
const submitting = ref(false)
const submitError = ref('')
const savedNotice = ref('')
let requestId = 0
let editorRequestId = 0

const recorded = computed(
  () => result.value?.summaries.reduce((total, item) => total + item.recorded, 0) ?? 0,
)
const filteredRooms = computed(() =>
  (result.value?.rooms ?? []).filter((room) => {
    return (
      (!search.value.trim() ||
        room.room_no.toLowerCase().includes(search.value.trim().toLowerCase())) &&
      (!onlyMissing.value || room.missing_count > 0)
    )
  }),
)
const visibleRooms = computed(() =>
  filteredRooms.value.slice((page.value - 1) * pageSize, page.value * pageSize),
)
const pageCount = computed(() => Math.max(1, Math.ceil(filteredRooms.value.length / pageSize)))

async function loadReadings(afterSave = false) {
  const unitId = selectedUnitId.value
  const currentRequest = ++requestId
  if (!unitId) {
    result.value = undefined
    loading.value = false
    error.value = ''
    return
  }
  if (!/^\d{4}-(0[1-9]|1[0-2])$/.test(month.value)) {
    loading.value = false
    error.value = '请选择有效的抄表月份'
    return
  }
  loading.value = true
  error.value = ''
  try {
    const next = await getUnitReadings(unitId, month.value)
    if (currentRequest !== requestId) return
    result.value = next
    page.value = Math.min(page.value, pageCount.value)
  } catch (cause) {
    if (currentRequest !== requestId) return
    error.value = afterSave
      ? `抄表已保存，但月度汇总刷新失败：${errorMessage(cause)}`
      : errorMessage(cause)
  } finally {
    if (currentRequest === requestId) loading.value = false
  }
}

async function openEditor(room: RoomMonthReadings) {
  if (submitting.value || !room.readings.length) return
  const currentRequest = ++editorRequestId
  editorOpen.value = true
  editingRoom.value = room
  editorLoading.value = true
  editorError.value = ''
  submitError.value = ''
  try {
    const latest = await getRoomReadings(room.room_id, room.period_month)
    if (currentRequest === editorRequestId) editingRoom.value = latest
  } catch (cause) {
    if (currentRequest === editorRequestId) editorError.value = errorMessage(cause)
  } finally {
    if (currentRequest === editorRequestId) editorLoading.value = false
  }
}

async function save(input: SaveReadingsInput) {
  const room = editingRoom.value
  const unitId = selectedUnitId.value
  if (!room || !unitId || submitting.value || editorLoading.value || editorError.value) return
  if (room.unit_id !== unitId || input.period_month !== month.value) return
  submitting.value = true
  submitError.value = ''
  try {
    await saveRoomReadings(room.room_id, input)
    if (unitId !== selectedUnitId.value || input.period_month !== month.value) return
    editorOpen.value = false
    savedNotice.value = `${room.room_no} · ${input.period_month} 的抄表记录已保存。`
    ElMessage.success('抄表已保存')
    await loadReadings(true)
  } catch (cause) {
    if (unitId === selectedUnitId.value && input.period_month === month.value)
      submitError.value = errorMessage(cause)
  } finally {
    submitting.value = false
  }
}

watch(
  [selectedUnitId, month],
  () => {
    result.value = undefined
    page.value = 1
    search.value = ''
    onlyMissing.value = false
    editorOpen.value = false
    editingRoom.value = undefined
    editorRequestId++
    savedNotice.value = ''
    void loadReadings()
  },
  { immediate: true },
)

watch([search, onlyMissing], () => {
  page.value = 1
})

onBeforeUnmount(() => {
  requestId++
  editorRequestId++
})
</script>

<template>
  <div class="readings-page">
    <PageHeader title="月度抄表" description="记录用量，让每月费用一目了然。">
      <template #actions>
        <el-button :loading="loading" :disabled="!selectedUnitId" @click="loadReadings()">
          刷新记录
        </el-button>
      </template>
    </PageHeader>

    <div class="panel toolbar readings-toolbar">
      <label class="field month-filter">
        <span>抄表月份</span>
        <input v-model="month" type="month" required :disabled="!selectedUnitId" />
      </label>
      <label class="field search-filter">
        <span>查找房间</span>
        <el-input v-model="search" clearable placeholder="输入房间号" :disabled="!selectedUnitId" />
      </label>
      <el-checkbox v-model="onlyMissing" class="missing-filter" :disabled="!selectedUnitId">
        只看待抄表
      </el-checkbox>
    </div>

    <el-alert
      v-if="savedNotice"
      :title="savedNotice"
      type="success"
      show-icon
      @close="savedNotice = ''"
    />

    <AsyncState
      :loading="loading"
      :error="error"
      :empty="!selectedUnitId || !result?.rooms.length"
      :empty-title="selectedUnitId ? '还没有房间可供抄表' : '先选择一个物业'"
      :empty-description="
        selectedUnitId
          ? '创建房间并绑定抄表项目，即可记录每月用量。'
          : '选择物业后，查看当月用量和抄表进度。'
      "
      @retry="loadReadings()"
    >
      <template v-if="result">
        <div class="reading-metrics">
          <div class="panel metric">
            <span>本月抄表费用</span>
            <strong>{{ formatMoney(result.total_fee) }}</strong>
          </div>
          <div class="panel metric">
            <span>已完成抄表</span>
            <strong>
              {{ recorded }}
              <small>项</small>
            </strong>
          </div>
          <div class="panel metric">
            <span>待录入</span>
            <strong>
              {{ result.missing_count }}
              <small>项</small>
            </strong>
          </div>
        </div>
        <section
          v-if="result.summaries.length"
          class="panel usage-summary"
          aria-label="抄表项目汇总"
        >
          <div v-for="summary in result.summaries" :key="summary.meter_item_id" class="usage-item">
            <div>
              <strong>{{ summary.code }}</strong>
              <span class="muted">
                {{ summary.recorded }} 项已录入 · {{ summary.missing }} 项待录入
              </span>
            </div>
            <div class="usage-totals">
              <span>
                {{ Number(summary.usage_amount.toFixed(4)) }}
                <small>{{ summary.unit }}</small>
              </span>
              <strong>{{ formatMoney(summary.fee) }}</strong>
            </div>
          </div>
        </section>
        <div class="room-section-heading">
          <h2>房间记录</h2>
          <span class="muted">共 {{ filteredRooms.length }} 间</span>
        </div>
        <AsyncState
          :empty="!filteredRooms.length"
          empty-title="没有匹配的房间"
          empty-description="试试其他房间号，或取消“只看待抄表”。"
        >
          <div class="reading-grid">
            <RoomReadingCard
              v-for="room in visibleRooms"
              :key="room.room_id"
              :room="room"
              :disabled="submitting"
              @edit="openEditor"
            />
          </div>
        </AsyncState>
        <div v-if="pageCount > 1" class="reading-pagination">
          <span class="muted">第 {{ page }} / {{ pageCount }} 页</span>
          <div>
            <el-button :disabled="page === 1" @click="page--">上一页</el-button>
            <el-button :disabled="page === pageCount" @click="page++">下一页</el-button>
          </div>
        </div>
      </template>
    </AsyncState>

    <el-dialog
      v-model="editorOpen"
      title="记录月度用量"
      width="min(600px, calc(100vw - 32px))"
      destroy-on-close
      :close-on-click-modal="false"
      :show-close="!submitting"
      :close-on-press-escape="!submitting"
    >
      <AsyncState
        :loading="editorLoading"
        :error="editorError"
        @retry="editingRoom && openEditor(editingRoom)"
      >
        <RoomReadingForm
          v-if="editingRoom"
          :key="`${editingRoom.room_id}-${editingRoom.period_month}`"
          :room="editingRoom"
          :busy="submitting"
          :error="submitError"
          @submit="save"
          @cancel="editorOpen = false"
        />
      </AsyncState>
    </el-dialog>
  </div>
</template>

<style scoped>
.readings-page {
  display: grid;
  gap: 1.5rem;
}
.readings-toolbar {
  padding: 1.25rem;
  align-items: flex-end;
  margin-bottom: 0;
}
.month-filter {
  width: 11rem;
}
.search-filter {
  width: 15rem;
}
.missing-filter {
  margin: 0 0 0.35rem auto;
}
.reading-metrics {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1rem;
}
.metric {
  display: grid;
  gap: 0.75rem;
  padding: 1.4rem 1.5rem;
}
.metric + .metric {
  margin-top: 0;
}
.metric > span {
  color: var(--text-secondary, #67717c);
  font-size: 0.8125rem;
}
.metric strong {
  font-size: clamp(1.25rem, 2.7vw, 2rem);
  font-weight: 600;
  letter-spacing: -0.045em;
  overflow-wrap: anywhere;
}
.metric small {
  color: var(--text-secondary, #67717c);
  font-size: 0.8125rem;
  font-weight: 400;
}
.usage-summary {
  padding: 0.25rem 1.25rem;
  margin-top: 1.5rem;
}
.usage-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 0.75rem;
  padding: 1rem 0;
}
.usage-item + .usage-item {
  border-top: 1px solid var(--border, #edf0f3);
}
.usage-item > div:first-child {
  display: grid;
  gap: 0.25rem;
}
.usage-item strong {
  font-size: 0.875rem;
  font-weight: 600;
}
.usage-item .muted,
.usage-totals small {
  font-size: 0.75rem;
}
.usage-totals {
  display: flex;
  align-items: baseline;
  gap: 2rem;
  font-variant-numeric: tabular-nums;
}
.usage-totals > span {
  font-size: 0.875rem;
}
.room-section-heading {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 1rem;
  margin: 1.75rem 0 1rem;
}
.room-section-heading h2 {
  margin: 0;
  font-size: 1.125rem;
  letter-spacing: -0.025em;
}
.room-section-heading > span {
  font-size: 0.8125rem;
}
.reading-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1rem;
}
.reading-grid > .reading-card {
  margin-top: 0;
}
.reading-pagination {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-top: 1.25rem;
  font-size: 0.8125rem;
}

@media (max-width: 1250px) {
  .reading-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 650px) {
  .readings-toolbar {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
  }
  .month-filter,
  .search-filter {
    width: 100%;
    min-width: 0;
  }
  .missing-filter {
    grid-column: 1 / -1;
    margin: 0;
  }
  .reading-grid {
    grid-template-columns: 1fr;
  }
  .reading-metrics {
    gap: 0.5rem;
  }
  .metric {
    padding: 1rem 0.75rem;
  }
}
</style>
