<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  cancelBill,
  generateRoomBill,
  generateUnitBills,
  getBill,
  listRoomBills,
  listUnitBills,
  markBillPaid,
} from '@/api/bills'
import { listRooms } from '@/api/rooms'
import AsyncState from '@/components/common/AsyncState.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import BillTable from '@/components/bills/BillTable.vue'
import GenerateBillForm from '@/components/bills/GenerateBillForm.vue'
import BillMeterReadings from '@/components/bills/BillMeterReadings.vue'
import { useUnits } from '@/stores/units'
import type { BillStatus, GenerateBillsInput, RentBill, RentBillDetail } from '@/types/bills'
import type { Room } from '@/types/rooms'
import { currentMonth, errorMessage, formatMoney } from '@/utils/format'

const { selectedUnitId } = useUnits()
const month = ref(currentMonth())
const status = ref<BillStatus | ''>('')
const roomId = ref<number>()
const rooms = ref<Room[]>([])
const roomsError = ref('')
const bills = ref<RentBill[]>([])
const loading = ref(false)
const error = ref('')
const offset = ref(0)
const pageSize = 20
const hasNext = ref(false)
const savedNotice = ref('')
const generateOpen = ref(false)
const submitting = ref(false)
const submitError = ref('')
const mutation = ref<{ kind: 'paid' | 'cancel'; bill: RentBill }>()
const cancelRemark = ref('')
const paidAt = ref('')
const detail = ref<RentBillDetail>()
const detailOpen = ref(false)
const detailLoading = ref(false)
const detailError = ref('')
let requestId = 0
let roomRequestId = 0
let detailRequestId = 0

const roomNames = computed(() => new Map(rooms.value.map((room) => [room.id, room.room_no])))
const unpaidAmount = computed(() =>
  bills.value.reduce((sum, bill) => sum + (bill.status === 'UNPAID' ? bill.total_amount : 0), 0),
)
const paidAmount = computed(() =>
  bills.value.reduce((sum, bill) => sum + (bill.status === 'PAID' ? bill.total_amount : 0), 0),
)
const validMonth = computed(() => !month.value || /^\d{4}-(0[1-9]|1[0-2])$/.test(month.value))
const statusLabels = { UNPAID: '待收款', PAID: '已收款', CANCELLED: '已取消' }

async function loadBills(afterSave = false) {
  const unitId = selectedUnitId.value
  const currentRequest = ++requestId
  if (!unitId) {
    bills.value = []
    loading.value = false
    error.value = ''
    return
  }
  if (!validMonth.value) {
    loading.value = false
    error.value = '请选择有效的账单月份'
    return
  }
  loading.value = true
  error.value = ''
  try {
    const query = {
      period_month: month.value || undefined,
      status: status.value || undefined,
      limit: pageSize + 1,
      offset: offset.value,
    }
    const result = roomId.value
      ? await listRoomBills(roomId.value, query)
      : await listUnitBills(unitId, query)
    if (currentRequest !== requestId) return
    bills.value = result.slice(0, pageSize)
    hasNext.value = result.length > pageSize
  } catch (cause) {
    if (currentRequest !== requestId) return
    error.value =
      afterSave || savedNotice.value
        ? `操作已完成，但账单刷新失败：${errorMessage(cause)}`
        : errorMessage(cause)
  } finally {
    if (currentRequest === requestId) loading.value = false
  }
}

async function loadRoomOptions() {
  const unitId = selectedUnitId.value
  const currentRequest = ++roomRequestId
  rooms.value = []
  roomsError.value = ''
  if (!unitId) return
  try {
    const result = await listRooms(unitId)
    if (currentRequest === roomRequestId) rooms.value = result
  } catch (cause) {
    if (currentRequest === roomRequestId)
      roomsError.value = `房间名称加载失败：${errorMessage(cause)}`
  }
}

function openGenerate() {
  if (!selectedUnitId.value || submitting.value) return
  submitError.value = ''
  generateOpen.value = true
}

async function generate(input: GenerateBillsInput, targetRoomId?: number) {
  const unitId = selectedUnitId.value
  if (!unitId || submitting.value) return
  if (targetRoomId !== undefined && !rooms.value.some((room) => room.id === targetRoomId)) return
  submitting.value = true
  submitError.value = ''
  try {
    const result = targetRoomId
      ? await generateRoomBill(targetRoomId, input)
      : await generateUnitBills(unitId, input)
    if (unitId !== selectedUnitId.value) return
    generateOpen.value = false
    savedNotice.value = `${input.period_month} 已生成或更新 ${result.saved_bills.length} 笔账单，跳过 ${result.skipped_bills.length} 笔已收款或已取消账单。`
    ElMessage.success('账单生成完成')
    if (month.value !== input.period_month) {
      month.value = input.period_month
    } else {
      await loadBills(true)
    }
  } catch (cause) {
    if (unitId === selectedUnitId.value) submitError.value = errorMessage(cause)
  } finally {
    submitting.value = false
  }
}

function openMutation(kind: 'paid' | 'cancel', bill: RentBill) {
  if (submitting.value || bill.status !== 'UNPAID') return
  mutation.value = { kind, bill }
  cancelRemark.value = ''
  paidAt.value = ''
  submitError.value = ''
}

async function saveMutation() {
  const action = mutation.value
  const unitId = selectedUnitId.value
  if (!action || submitting.value || action.bill.status !== 'UNPAID' || !unitId) return
  if (action.kind === 'paid' && paidAt.value && Number.isNaN(new Date(paidAt.value).getTime())) {
    submitError.value = '请选择有效的收款时间'
    return
  }
  submitting.value = true
  submitError.value = ''
  try {
    if (action.kind === 'paid') {
      await markBillPaid(
        action.bill.id,
        paidAt.value ? new Date(paidAt.value).toISOString() : undefined,
      )
    } else {
      await cancelBill(action.bill.id, cancelRemark.value.trim() || undefined)
    }
    if (unitId !== selectedUnitId.value) return
    mutation.value = undefined
    savedNotice.value = action.kind === 'paid' ? '收款已登记。' : '账单已取消。'
    ElMessage.success(savedNotice.value)
    await loadBills(true)
  } catch (cause) {
    if (unitId === selectedUnitId.value) submitError.value = errorMessage(cause)
  } finally {
    submitting.value = false
  }
}

async function showDetail(bill: RentBill) {
  const currentRequest = ++detailRequestId
  detail.value = { ...bill, meter_readings: null }
  detailOpen.value = true
  detailLoading.value = true
  detailError.value = ''
  try {
    const result = await getBill(bill.id)
    if (currentRequest === detailRequestId) detail.value = result
  } catch (cause) {
    if (currentRequest === detailRequestId) detailError.value = errorMessage(cause)
  } finally {
    if (currentRequest === detailRequestId) detailLoading.value = false
  }
}

function changePage(nextOffset: number) {
  if (loading.value || nextOffset < 0) return
  offset.value = nextOffset
  void loadBills()
}

function formatTime(value?: string) {
  return value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '—'
}

watch(
  selectedUnitId,
  () => {
    roomId.value = undefined
    generateOpen.value = false
    mutation.value = undefined
    detailOpen.value = false
    detailRequestId++
    savedNotice.value = ''
    bills.value = []
    void loadRoomOptions()
  },
  { immediate: true },
)

watch(
  [selectedUnitId, month, status, roomId],
  () => {
    offset.value = 0
    void loadBills()
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  requestId++
  roomRequestId++
  detailRequestId++
})
</script>

<template>
  <div class="bills-page">
    <PageHeader title="租金账单" description="每一笔应收，都清晰有据。">
      <template #actions>
        <el-button type="primary" :disabled="!selectedUnitId || submitting" @click="openGenerate">
          生成月度账单
        </el-button>
      </template>
    </PageHeader>

    <div v-if="selectedUnitId" class="bill-metrics">
      <div class="panel metric">
        <span>本页待收</span>
        <strong>{{ loading || error ? '—' : formatMoney(unpaidAmount) }}</strong>
      </div>
      <div class="panel metric">
        <span>本页已收</span>
        <strong>{{ loading || error ? '—' : formatMoney(paidAmount) }}</strong>
      </div>
      <div class="panel metric">
        <span>本页账单</span>
        <strong>
          {{ loading || error ? '—' : bills.length }}
          <small>笔</small>
        </strong>
      </div>
    </div>

    <el-alert
      v-if="savedNotice"
      :title="savedNotice"
      type="success"
      show-icon
      @close="savedNotice = ''"
    />

    <section class="panel bill-panel" aria-label="账单列表">
      <div class="toolbar bill-toolbar">
        <label class="field filter-month">
          <span>账单月份</span>
          <el-date-picker
            v-model="month"
            type="month"
            format="YYYY-MM"
            value-format="YYYY-MM"
            value-on-clear=""
            popper-class="rent-date-picker"
            placeholder="全部月份"
            :disabled="!selectedUnitId"
            aria-label="账单月份"
          />
        </label>
        <label class="field filter-select">
          <span>收款状态</span>
          <el-select
            v-model="status"
            :disabled="!selectedUnitId"
            :empty-values="[null, undefined]"
            aria-label="收款状态"
          >
            <el-option value="" label="全部状态" />
            <el-option value="UNPAID" label="待收款" />
            <el-option value="PAID" label="已收款" />
            <el-option value="CANCELLED" label="已取消" />
          </el-select>
        </label>
        <label class="field filter-select">
          <span>房间</span>
          <el-select
            v-model="roomId"
            clearable
            filterable
            placeholder="全部房间"
            :disabled="!selectedUnitId || !!roomsError"
            aria-label="筛选房间"
          >
            <el-option
              v-for="room in rooms"
              :key="room.id"
              :label="room.room_no"
              :value="room.id"
            />
          </el-select>
        </label>
        <el-button
          class="refresh-button"
          :loading="loading"
          :disabled="!selectedUnitId"
          @click="loadBills()"
        >
          刷新
        </el-button>
      </div>
      <div v-if="roomsError" class="room-warning">
        <span>{{ roomsError }}</span>
        <el-button text size="small" @click="loadRoomOptions">重新加载房间</el-button>
      </div>
      <AsyncState
        :loading="loading"
        :error="error"
        :empty="!selectedUnitId || !bills.length"
        :empty-title="selectedUnitId ? '暂无符合条件的账单' : '先选择一个物业'"
        :empty-description="
          selectedUnitId
            ? '调整筛选条件，或完成抄表后生成月度账单。'
            : '选择物业后，查看和管理租金账单。'
        "
        @retry="loadBills()"
      >
        <BillTable
          :bills="bills"
          :room-names="roomNames"
          :disabled="submitting"
          @paid="openMutation('paid', $event)"
          @cancel="openMutation('cancel', $event)"
          @detail="showDetail"
        />
      </AsyncState>
      <div v-if="selectedUnitId && !error && (bills.length || offset)" class="bill-pagination">
        <span class="muted">第 {{ offset / pageSize + 1 }} 页 · 每页最多 {{ pageSize }} 笔</span>
        <div class="pagination-actions">
          <el-button :disabled="loading || offset === 0" @click="changePage(offset - pageSize)">
            上一页
          </el-button>
          <el-button :disabled="loading || !hasNext" @click="changePage(offset + pageSize)">
            下一页
          </el-button>
        </div>
      </div>
    </section>

    <el-dialog
      v-model="generateOpen"
      title="生成月度账单"
      width="min(560px, calc(100vw - 32px))"
      destroy-on-close
      :close-on-click-modal="false"
      :show-close="!submitting"
      :close-on-press-escape="!submitting"
    >
      <GenerateBillForm
        :month="month || currentMonth()"
        :rooms="rooms"
        :busy="submitting"
        :error="submitError"
        :rooms-error="roomsError"
        @submit="generate"
        @cancel="generateOpen = false"
      />
    </el-dialog>

    <el-dialog
      :model-value="!!mutation"
      :title="mutation?.kind === 'paid' ? '登记收款' : '取消账单'"
      width="min(500px, calc(100vw - 32px))"
      :close-on-click-modal="false"
      :show-close="!submitting"
      :close-on-press-escape="!submitting"
      @update:model-value="
        (value) => {
          if (!value) mutation = undefined
        }
      "
    >
      <form v-if="mutation" class="mutation-form" @submit.prevent="saveMutation">
        <div class="mutation-summary">
          <span>
            {{ roomNames.get(mutation.bill.room_id) ?? `房间 #${mutation.bill.room_id}` }} ·
            {{ mutation.bill.period_month }}
          </span>
          <strong>{{ formatMoney(mutation.bill.total_amount) }}</strong>
        </div>
        <template v-if="mutation.kind === 'paid'">
          <p class="muted">请确认已收到该笔款项。登记后账单金额将锁定。</p>
          <label class="field">
            <span>
              收款时间
              <span class="muted">（选填，默认现在）</span>
            </span>
            <el-date-picker
              v-model="paidAt"
              type="datetime"
              format="YYYY-MM-DD HH:mm"
              value-format="YYYY-MM-DD[T]HH:mm"
              value-on-clear=""
              popper-class="rent-date-picker"
              placeholder="选择收款时间"
              :disabled="submitting"
              aria-label="收款时间"
            />
          </label>
        </template>
        <template v-else>
          <p class="muted">取消后无法恢复，该房间同一月份不会重新生成账单。</p>
          <label class="field">
            <span>
              取消原因
              <span class="muted">（选填）</span>
            </span>
            <el-input
              v-model="cancelRemark"
              type="textarea"
              :rows="3"
              :disabled="submitting"
              maxlength="500"
            />
          </label>
        </template>
        <el-alert
          v-if="submitError"
          :title="submitError"
          type="error"
          :closable="false"
          show-icon
        />
        <div class="form-actions">
          <el-button :disabled="submitting" @click="mutation = undefined">返回</el-button>
          <el-button
            :type="mutation.kind === 'paid' ? 'primary' : 'danger'"
            native-type="submit"
            :loading="submitting"
          >
            {{ mutation.kind === 'paid' ? '确认已收款' : '确认取消账单' }}
          </el-button>
        </div>
      </form>
    </el-dialog>

    <el-dialog v-model="detailOpen" title="账单详情" width="min(680px, calc(100vw - 32px))">
      <AsyncState
        :loading="detailLoading"
        :error="detailError"
        @retry="detail && showDetail(detail)"
      >
        <dl v-if="detail" class="bill-details">
          <div>
            <dt>账单编号</dt>
            <dd>#{{ detail.id }}</dd>
          </div>
          <div>
            <dt>房间</dt>
            <dd>{{ roomNames.get(detail.room_id) ?? `房间 #${detail.room_id}` }}</dd>
          </div>
          <div>
            <dt>账单月份</dt>
            <dd>{{ detail.period_month }}</dd>
          </div>
          <div>
            <dt>租金</dt>
            <dd>{{ formatMoney(detail.base_rent) }}</dd>
          </div>
          <div>
            <dt>抄表费用</dt>
            <dd>{{ formatMoney(detail.meter_fee) }}</dd>
          </div>
          <div>
            <dt>总金额</dt>
            <dd>
              <strong>{{ formatMoney(detail.total_amount) }}</strong>
            </dd>
          </div>
          <div>
            <dt>状态</dt>
            <dd>{{ statusLabels[detail.status] }}</dd>
          </div>
          <div>
            <dt>收款时间</dt>
            <dd>{{ formatTime(detail.paid_at) }}</dd>
          </div>
          <div>
            <dt>创建时间</dt>
            <dd>{{ formatTime(detail.created_at) }}</dd>
          </div>
          <div>
            <dt>更新时间</dt>
            <dd>{{ formatTime(detail.updated_at) }}</dd>
          </div>
          <div>
            <dt>备注</dt>
            <dd>{{ detail.remark || '—' }}</dd>
          </div>
        </dl>
        <section v-if="detail" class="bill-readings-section" aria-label="账单抄表明细">
          <BillMeterReadings
            v-if="detail.meter_readings"
            :readings="detail.meter_readings"
            :period-month="detail.period_month"
          />
          <div v-else class="bill-readings-missing">
            <h3>抄表明细</h3>
            <p>此历史账单未保存生成时的抄表明细。</p>
            <p v-if="detail.status === 'UNPAID'">
              重新生成该月份账单后，可保存本次生成时的抄表明细。
            </p>
          </div>
        </section>
      </AsyncState>
    </el-dialog>
  </div>
</template>

<style scoped>
.bills-page {
  display: grid;
  gap: 1.5rem;
}
.bill-metrics {
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
.bill-panel {
  overflow: hidden;
  padding: 0;
}
.bill-toolbar {
  padding: 1.25rem;
  align-items: flex-end;
  margin-bottom: 0;
}
.filter-month {
  width: 11rem;
}
.filter-select {
  width: 10rem;
}
.filter-select :deep(.el-select__wrapper) {
  height: 40px;
}
.refresh-button {
  margin-left: auto;
}
.room-warning {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
  padding: 0 1.25rem 1rem;
  color: #936225;
  font-size: 0.8125rem;
}
.bill-pagination {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  flex-wrap: wrap;
  align-items: center;
  padding: 1.25rem;
  border-top: 1px solid var(--border, #edf0f3);
  font-size: 0.8125rem;
}
.mutation-form {
  display: grid;
  gap: 1rem;
}
.mutation-form p {
  margin: 0;
  line-height: 1.7;
}
.mutation-summary {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 1rem;
  border-radius: 0.75rem;
  background: var(--surface-subtle, #f5f7fa);
}
.mutation-summary strong {
  font-size: 1.5rem;
  letter-spacing: -0.04em;
}
.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  margin-top: 0.5rem;
}
.bill-details {
  display: grid;
  gap: 0.875rem;
  margin: 0;
}
.bill-details > div {
  display: grid;
  grid-template-columns: 6rem 1fr;
  gap: 1rem;
}
.bill-details dt {
  color: var(--text-secondary, #67717c);
}
.bill-details dd {
  margin: 0;
  text-align: right;
  overflow-wrap: anywhere;
}
.bill-readings-section {
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--border, #edf0f3);
}
.bill-readings-missing h3 {
  margin: 0 0 0.75rem;
  font-size: 0.9375rem;
  font-weight: 600;
}
.bill-readings-missing p {
  margin: 0.25rem 0 0;
  color: var(--text-secondary, #67717c);
  font-size: 0.8125rem;
  line-height: 1.7;
}

@media (max-width: 600px) {
  .bill-metrics {
    gap: 0.5rem;
  }
  .metric {
    padding: 1rem 0.75rem;
  }
  .bill-toolbar {
    display: grid;
    grid-template-columns: 1fr 1fr;
  }
  .filter-month,
  .filter-select {
    width: 100%;
    min-width: 0;
  }
  .refresh-button {
    width: 100%;
    margin: auto 0 0;
  }
}
</style>
