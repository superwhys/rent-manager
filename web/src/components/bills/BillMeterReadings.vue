<script setup lang="ts">
import type { MeterReadingLine } from '@/types/readings'
import { formatDate, formatMoney } from '@/utils/format'
import { formatReadingPrice } from '@/utils/readings'

defineProps<{ readings: MeterReadingLine[]; periodMonth: string }>()

function sourceLabel(source?: string): string {
  switch (source) {
    case 'MANUAL':
      return '手动录入'
    case 'IMPORT':
      return '导入'
    case 'AUTO':
      return '自动抄表'
    default:
      return '—'
  }
}
</script>

<template>
  <section class="bill-readings">
    <div class="section-heading">
      <h3>抄表明细</h3>
      <span class="muted">{{ readings.length }} 个抄表项</span>
    </div>
    <p class="muted readings-note">
      以下为 {{ periodMonth }} 账单生成时保存的抄表明细。
    </p>
    <p v-if="!readings.length" class="muted readings-empty">生成账单时，该房间未配置抄表项</p>
    <article
      v-for="line in readings"
      :key="line.meter_item.id"
      class="reading-item"
    >
      <div class="item-heading">
        <h4>{{ line.meter_item.code }}</h4>
        <span class="reading-status" :class="{ pending: !line.is_recorded }">
          {{ line.is_recorded ? '已录入' : '未录入' }}
        </span>
      </div>
      <div class="reading-meta">
        <span>
          {{ line.meter_item.calc_type === 'DIFFERENCE' ? '按差值计费' : '按本次用量计费' }}
        </span>
        <span>
          {{ formatReadingPrice(line.meter_item.unit_price) }} / {{ line.meter_item.unit }}
        </span>
      </div>
      <dl class="reading-values">
        <div v-if="line.meter_item.calc_type === 'DIFFERENCE'">
          <dt>上次 / 起始读数</dt>
          <dd>{{ line.previous_reading ?? '—' }}</dd>
        </div>
        <div>
          <dt>{{ line.meter_item.calc_type === 'DIFFERENCE' ? '本次读数' : '本次用量' }}</dt>
          <dd>{{ line.current_reading ?? '—' }}</dd>
        </div>
        <div>
          <dt>用量</dt>
          <dd>
            {{ line.usage_amount ?? '—' }}
            <span v-if="line.usage_amount != null">{{ line.meter_item.unit }}</span>
          </dd>
        </div>
        <div>
          <dt>费用</dt>
          <dd>{{ line.is_recorded ? formatMoney(line.fee) : '—' }}</dd>
        </div>
        <div>
          <dt>抄表时间</dt>
          <dd>{{ line.reading_time ? formatDate(line.reading_time) : '—' }}</dd>
        </div>
        <div>
          <dt>来源</dt>
          <dd>{{ sourceLabel(line.source) }}</dd>
        </div>
        <div v-if="line.remark" class="reading-remark">
          <dt>备注</dt>
          <dd>{{ line.remark }}</dd>
        </div>
      </dl>
    </article>
  </section>
</template>

<style scoped>
.bill-readings {
  display: grid;
  gap: 0.875rem;
}
.section-heading,
.item-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}
.section-heading h3,
.item-heading h4 {
  margin: 0;
  font-weight: 600;
  overflow-wrap: anywhere;
}
.section-heading h3 {
  font-size: 0.9375rem;
}
.section-heading > span {
  flex-shrink: 0;
  font-size: 0.75rem;
}
.readings-note,
.readings-empty {
  margin: 0;
  font-size: 0.8125rem;
  line-height: 1.7;
}
.readings-empty {
  border-radius: 0.75rem;
  padding: 1rem;
  background: var(--surface-secondary, #f3f6f6);
}
.reading-item {
  min-width: 0;
  border: 1px solid var(--border, #e6e9ee);
  border-radius: 1rem;
  padding: 1rem;
}
.item-heading h4 {
  font-size: 0.875rem;
}
.reading-status {
  flex-shrink: 0;
  border-radius: 999px;
  padding: 0.3rem 0.6rem;
  background: var(--accent-soft, #e9f4ef);
  color: var(--accent, #187b62);
  font-size: 0.6875rem;
}
.reading-status.pending {
  background: #fff3df;
  color: #936225;
}
.reading-meta {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  gap: 0.5rem;
  margin: 0.5rem 0 1rem;
  color: var(--text-secondary, #67717c);
  font-size: 0.75rem;
  overflow-wrap: anywhere;
}
.reading-values {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.875rem 1rem;
  margin: 0;
}
.reading-values > div {
  min-width: 0;
}
.reading-values dt {
  margin-bottom: 0.2rem;
  color: var(--text-secondary, #67717c);
  font-size: 0.75rem;
}
.reading-values dd {
  margin: 0;
  font-size: 0.8125rem;
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere;
}
.reading-remark {
  grid-column: 1 / -1;
}
.reading-remark dd {
  white-space: pre-wrap;
}
</style>
