<script setup lang="ts">
import { computed, ref } from 'vue'
import type {
  MeterReadingInput,
  MeterReadingLine,
  RoomMonthReadings,
  SaveReadingsInput,
} from '@/types/readings'
import { formatReadingPrice, readingError, readingUsage } from '@/utils/readings'

const props = defineProps<{
  room: RoomMonthReadings
  busy: boolean
  error: string
}>()

const emit = defineEmits<{
  submit: [input: SaveReadingsInput]
  cancel: []
}>()

interface ReadingDraft {
  line: MeterReadingLine
  current?: number
  previous?: number
  remark: string
  changed: boolean
}

const rows = ref<ReadingDraft[]>(
  props.room.readings.map((line) => ({
    line,
    current: line.current_reading,
    previous: line.previous_reading,
    remark: line.remark ?? '',
    changed: false,
  })),
)
const attempted = ref(false)
const filledCount = computed(
  () => rows.value.filter((row) => row.current !== undefined && row.current !== null).length,
)

function rowError(row: ReadingDraft) {
  return readingError(row.line.meter_item.calc_type, row.current, row.previous)
}

function usage(row: ReadingDraft) {
  return readingUsage(row.line.meter_item.calc_type, row.current, row.previous)
}

function submit() {
  if (props.busy) return
  attempted.value = true
  if (!filledCount.value || rows.value.some((row) => !!rowError(row))) return
  const readings: MeterReadingInput[] = []
  for (const row of rows.value) {
    if (row.current === undefined || row.current === null) continue
    readings.push({
      meter_item_id: row.line.meter_item.id,
      current_reading: row.current,
      previous_reading: row.line.meter_item.calc_type === 'DIFFERENCE' ? row.previous : undefined,
      reading_time: row.changed ? undefined : row.line.reading_time,
      remark: row.remark.trim() || undefined,
    })
  }
  emit('submit', { period_month: props.room.period_month, readings })
}
</script>

<template>
  <form class="reading-form" @submit.prevent="submit">
    <p class="form-context">
      <strong>{{ room.room_no }}</strong>
      <span>{{ room.period_month }}</span>
    </p>
    <p class="muted form-intro">
      填写本月读数，可先保存已完成的项目。历史表底自动沿用，首次抄表请填写起始读数。
    </p>
    <fieldset
      v-for="row in rows"
      :key="row.line.meter_item.id"
      class="reading-item"
      :disabled="busy"
    >
      <legend>
        {{ row.line.meter_item.code }}
        <span class="muted">{{ row.line.meter_item.unit }}</span>
      </legend>
      <div class="reading-meta">
        <span>
          {{ row.line.meter_item.calc_type === 'DIFFERENCE' ? '按差值计费' : '按本次用量计费' }}
        </span>
        <span>
          {{ formatReadingPrice(row.line.meter_item.unit_price) }} / {{ row.line.meter_item.unit }}
        </span>
      </div>
      <div class="reading-fields">
        <label v-if="row.line.meter_item.calc_type === 'DIFFERENCE'" class="field">
          <span>上次 / 起始读数</span>
          <el-input-number
            v-model="row.previous"
            :min="0"
            :step="1"
            :controls="false"
            :disabled="busy || row.line.previous_reading !== undefined"
            :aria-label="`${row.line.meter_item.code}上次读数`"
            @change="row.changed = true"
          />
          <small v-if="row.line.previous_reading !== undefined" class="muted">
            沿用已记录的表底
          </small>
        </label>
        <label class="field">
          <span>
            {{ row.line.meter_item.calc_type === 'DIFFERENCE' ? '本次读数' : '本次用量' }}
          </span>
          <el-input-number
            v-model="row.current"
            :min="0"
            :step="1"
            :controls="false"
            :disabled="busy"
            :aria-label="`${row.line.meter_item.code}本次读数`"
            placeholder="待填写"
            @change="row.changed = true"
          />
        </label>
      </div>
      <p v-if="rowError(row)" class="field-error" role="alert">{{ rowError(row) }}</p>
      <p v-else-if="usage(row) !== undefined" class="usage-preview">
        本月用量
        <strong>{{ Number(usage(row)?.toFixed(4)) }} {{ row.line.meter_item.unit }}</strong>
        <span>费用以保存后的账目为准</span>
      </p>
      <label class="field reading-remark">
        <span>
          备注
          <span class="muted">（选填）</span>
        </span>
        <el-input
          v-model="row.remark"
          :disabled="busy"
          maxlength="500"
          placeholder="例如：更换表具、读数说明"
        />
      </label>
    </fieldset>
    <p v-if="attempted && !filledCount" class="field-error" role="alert">请至少填写一项本次读数</p>
    <el-alert v-if="error" :title="error" type="error" :closable="false" show-icon />
    <div class="form-actions">
      <span class="muted">已填写 {{ filledCount }} / {{ rows.length }} 项</span>
      <el-button :disabled="busy" @click="emit('cancel')">返回</el-button>
      <el-button type="primary" native-type="submit" :loading="busy" :disabled="!rows.length">
        保存抄表
      </el-button>
    </div>
  </form>
</template>

<style scoped>
.reading-form {
  display: grid;
  gap: 1.25rem;
}
.form-context {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin: 0;
}
.form-context strong {
  font-size: 1.5rem;
  letter-spacing: -0.035em;
  overflow-wrap: anywhere;
}
.form-context > span {
  color: var(--text-secondary, #67717c);
}
.form-intro {
  margin: 0;
  font-size: 0.8125rem;
  line-height: 1.7;
}
.reading-item {
  min-width: 0;
  border: 1px solid var(--border, #e6e9ee);
  border-radius: 1rem;
  padding: 1rem;
}
.reading-item legend {
  padding: 0 0.4rem;
  font-weight: 600;
}
.reading-item legend span {
  margin-left: 0.375rem;
  font-size: 0.75rem;
  font-weight: 400;
}
.reading-meta {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  gap: 0.5rem;
  margin-bottom: 1rem;
  color: var(--text-secondary, #67717c);
  font-size: 0.75rem;
}
.reading-fields {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1rem;
}
.reading-fields :deep(.el-input-number) {
  width: 100%;
}
.reading-fields :deep(.el-input__inner) {
  text-align: left;
}
.field-error {
  margin: 0.5rem 0 0;
  color: var(--danger, #b64646);
  font-size: 0.8125rem;
}
.usage-preview {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin: 0.75rem 0 0;
  font-size: 0.8125rem;
}
.usage-preview > span {
  color: var(--text-secondary, #67717c);
  font-size: 0.75rem;
}
.reading-remark {
  margin-top: 1rem;
}
.form-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.5rem;
  flex-wrap: wrap;
}
.form-actions > span {
  margin-right: auto;
  font-size: 0.75rem;
}

@media (max-width: 450px) {
  .reading-fields {
    grid-template-columns: 1fr;
  }
  .form-actions > span {
    width: 100%;
  }
}
</style>
