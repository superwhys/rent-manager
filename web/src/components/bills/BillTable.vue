<script setup lang="ts">
import type { RentBill } from '@/types/bills'
import { formatMoney } from '@/utils/format'

defineProps<{
  bills: RentBill[]
  roomNames: Map<number, string>
  disabled: boolean
}>()

defineEmits<{
  paid: [bill: RentBill]
  cancel: [bill: RentBill]
  detail: [bill: RentBill]
}>()

const statusLabels = { UNPAID: '待收款', PAID: '已收款', CANCELLED: '已取消' }
</script>

<template>
  <div class="bill-list">
    <div class="bill-heading" aria-hidden="true">
      <span>房间 / 账期</span>
      <span>费用明细</span>
      <span>应收金额</span>
      <span>状态</span>
      <span class="align-end">操作</span>
    </div>
    <article v-for="bill in bills" :key="bill.id" class="bill-row">
      <div class="bill-room">
        <button class="bill-detail" type="button" @click="$emit('detail', bill)">
          {{ roomNames.get(bill.room_id) ?? `房间 #${bill.room_id}` }}
        </button>
        <span class="muted">{{ bill.period_month }} · 账单 #{{ bill.id }}</span>
      </div>
      <div class="bill-breakdown">
        <span>租金 {{ formatMoney(bill.base_rent) }}</span>
        <span class="muted">抄表 {{ formatMoney(bill.meter_fee) }}</span>
      </div>
      <strong class="bill-amount">{{ formatMoney(bill.total_amount) }}</strong>
      <span class="bill-status" :class="bill.status.toLowerCase()">
        {{ statusLabels[bill.status] }}
      </span>
      <div class="bill-actions">
        <template v-if="bill.status === 'UNPAID'">
          <el-button :disabled="disabled" size="small" @click="$emit('paid', bill)">
            登记收款
          </el-button>
          <el-button :disabled="disabled" size="small" text @click="$emit('cancel', bill)">
            取消
          </el-button>
        </template>
        <el-button size="small" text @click="$emit('detail', bill)">查看详情</el-button>
      </div>
    </article>
  </div>
</template>

<style scoped>
.bill-heading,
.bill-row {
  display: grid;
  grid-template-columns: 1.3fr 1.2fr 1fr 0.8fr 1.5fr;
  gap: 1rem;
  align-items: center;
}

.bill-heading {
  padding: 0.75rem 1.25rem;
  color: var(--text-secondary, #667085);
  font-size: 0.75rem;
}

.bill-row {
  min-height: 6rem;
  padding: 1.25rem;
  border-top: 1px solid var(--border, #edf0f3);
}

.bill-room,
.bill-breakdown {
  display: flex;
  min-width: 0;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.25rem;
  font-size: 0.8125rem;
}

.bill-detail {
  max-width: 100%;
  border: 0;
  padding: 0;
  background: transparent;
  color: inherit;
  font: inherit;
  font-size: 0.9375rem;
  font-weight: 650;
  text-align: left;
  overflow-wrap: anywhere;
  cursor: pointer;
}

.bill-amount {
  font-variant-numeric: tabular-nums;
  letter-spacing: -0.025em;
  overflow-wrap: anywhere;
}

.bill-status {
  width: fit-content;
  padding: 0.3rem 0.65rem;
  border-radius: 999px;
  font-size: 0.75rem;
  white-space: nowrap;
}

.unpaid {
  color: #936225;
  background: #fff3df;
}
.paid {
  color: #28705b;
  background: #e9f5ef;
}
.cancelled {
  color: #67717c;
  background: #f0f2f5;
}

.bill-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 0.25rem;
}

.bill-actions :deep(.el-button + .el-button) {
  margin-left: 0;
}
.align-end {
  text-align: right;
}

@media (max-width: 1100px) {
  .bill-heading {
    display: none;
  }
  .bill-row {
    grid-template-columns: 1fr auto;
    gap: 0.875rem;
  }
  .bill-breakdown {
    grid-column: 1;
    grid-row: 2;
  }
  .bill-amount {
    grid-column: 2;
    grid-row: 1;
    text-align: right;
  }
  .bill-status {
    grid-column: 2;
    grid-row: 2;
    justify-self: end;
  }
  .bill-actions {
    grid-column: 1 / -1;
  }
}
</style>
