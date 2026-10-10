<script setup lang="ts">
import type { RoomMonthReadings } from '@/types/readings'
import { formatMoney } from '@/utils/format'

defineProps<{ room: RoomMonthReadings; disabled: boolean }>()
defineEmits<{ edit: [room: RoomMonthReadings] }>()
</script>

<template>
  <article class="panel reading-card">
    <div class="card-heading">
      <div>
        <h2>{{ room.room_no }}</h2>
        <span class="muted">{{ room.readings.length }} 个抄表项</span>
      </div>
      <span class="reading-status" :class="{ pending: room.missing_count > 0 }">
        {{
          room.missing_count > 0
            ? `待录入 ${room.missing_count} 项`
            : room.readings.length
              ? '已完成'
              : '未配置'
        }}
      </span>
    </div>
    <div class="reading-lines">
      <div v-for="line in room.readings" :key="line.meter_item.id" class="reading-line">
        <div class="reading-name">
          <strong>{{ line.meter_item.code }}</strong>
          <span class="muted">
            {{
              line.is_recorded ? `${line.usage_amount ?? '—'} ${line.meter_item.unit}` : '尚未录入'
            }}
          </span>
        </div>
        <div class="reading-value">
          <strong>{{ line.is_recorded ? formatMoney(line.fee) : '—' }}</strong>
          <span v-if="line.is_recorded" class="muted">读数 {{ line.current_reading }}</span>
        </div>
      </div>
      <p v-if="!room.readings.length" class="muted unconfigured">
        在房间设置中绑定抄表项后，即可开始记录用量。
      </p>
    </div>
    <div class="card-footer">
      <div>
        <span class="muted">本月费用</span>
        <strong>{{ formatMoney(room.total_fee) }}</strong>
      </div>
      <el-button :disabled="disabled || !room.readings.length" @click="$emit('edit', room)">
        {{ room.missing_count ? '录入读数' : '查看与编辑' }}
      </el-button>
    </div>
  </article>
</template>

<style scoped>
.reading-card {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding: 0;
}
.card-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding: 1.25rem;
}
.card-heading h2 {
  margin: 0 0 0.25rem;
  font-size: 1.125rem;
  font-weight: 650;
  letter-spacing: -0.025em;
  overflow-wrap: anywhere;
}
.card-heading .muted {
  font-size: 0.75rem;
}
.reading-status {
  flex-shrink: 0;
  padding: 0.3rem 0.6rem;
  border-radius: 999px;
  background: #e9f5ef;
  color: #28705b;
  font-size: 0.6875rem;
}
.reading-status.pending {
  background: #fff3df;
  color: #936225;
}
.reading-lines {
  flex: 1;
  padding: 0 1.25rem;
}
.reading-line {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.8rem 0;
  border-top: 1px solid var(--border, #edf0f3);
}
.reading-name,
.reading-value {
  display: grid;
  gap: 0.2rem;
  font-size: 0.8125rem;
}
.reading-name strong {
  font-weight: 500;
  overflow-wrap: anywhere;
}
.reading-value {
  text-align: right;
  font-variant-numeric: tabular-nums;
}
.reading-line .muted {
  font-size: 0.75rem;
}
.card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 0.75rem;
  padding: 1rem 1.25rem;
  background: var(--surface-subtle, #f8fafb);
}
.card-footer > div {
  display: grid;
  gap: 0.25rem;
}
.card-footer .muted {
  font-size: 0.6875rem;
}
.card-footer strong {
  font-size: 1.125rem;
  letter-spacing: -0.035em;
}
.unconfigured {
  margin: 0 0 1.25rem;
  font-size: 0.8125rem;
  line-height: 1.7;
}
</style>
