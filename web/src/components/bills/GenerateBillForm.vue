<script setup lang="ts">
import { computed, ref } from 'vue'
import type { GenerateBillsInput } from '@/types/bills'
import type { Room } from '@/types/rooms'

const props = defineProps<{
  month: string
  rooms: Room[]
  busy: boolean
  error: string
  roomsError: string
}>()

const emit = defineEmits<{
  submit: [input: GenerateBillsInput, roomId?: number]
  cancel: []
}>()

const periodMonth = ref(props.month)
const scope = ref<'unit' | 'room'>('unit')
const roomId = ref<number>()
const remark = ref('')
const attempted = ref(false)
const monthValid = computed(() => /^\d{4}-(0[1-9]|1[0-2])$/.test(periodMonth.value))
const roomValid = computed(
  () => scope.value === 'unit' || props.rooms.some((room) => room.id === roomId.value),
)

function submit() {
  if (props.busy) return
  attempted.value = true
  if (!monthValid.value || !roomValid.value) return
  emit(
    'submit',
    {
      period_month: periodMonth.value,
      remark: remark.value.trim() || undefined,
    },
    scope.value === 'room' ? roomId.value : undefined,
  )
}
</script>

<template>
  <form class="generate-form" @submit.prevent="submit">
    <p class="muted form-intro">按房间租金和当月抄表费用生成账单，请先完成当月抄表。</p>
    <label class="field">
      <span>账单月份</span>
      <input v-model="periodMonth" type="month" required :disabled="busy" />
      <small v-if="attempted && !monthValid" class="field-error">请选择有效的账单月份</small>
    </label>
    <div class="field">
      <span>生成范围</span>
      <el-radio-group v-model="scope" :disabled="busy">
        <el-radio-button value="unit">当前物业</el-radio-button>
        <el-radio-button value="room">指定房间</el-radio-button>
      </el-radio-group>
    </div>
    <label v-if="scope === 'room'" class="field">
      <span>房间</span>
      <el-select
        v-model="roomId"
        placeholder="选择房间"
        filterable
        :disabled="busy"
        aria-label="生成账单的房间"
      >
        <el-option v-for="room in rooms" :key="room.id" :value="room.id" :label="room.room_no" />
      </el-select>
      <small v-if="roomsError" class="field-error">{{ roomsError }}</small>
      <small v-else-if="attempted && !roomValid" class="field-error">请选择房间</small>
    </label>
    <p class="generation-note">
      {{ scope === 'unit' ? '批量生成仅包含已出租房间。' : '为选定房间生成本月账单。' }}
      待收款账单将更新金额，已收款和已取消账单会自动跳过。
    </p>
    <label class="field">
      <span>
        备注
        <span class="muted">（选填）</span>
      </span>
      <el-input v-model="remark" type="textarea" :rows="2" :disabled="busy" maxlength="500" />
    </label>
    <el-alert v-if="error" :title="error" type="error" :closable="false" show-icon />
    <div class="form-actions">
      <el-button :disabled="busy" @click="emit('cancel')">返回</el-button>
      <el-button type="primary" native-type="submit" :loading="busy">生成账单</el-button>
    </div>
  </form>
</template>

<style scoped>
.generate-form {
  display: grid;
  gap: 1.25rem;
}
.form-intro {
  margin: 0;
}
.generation-note {
  margin: 0;
  border-radius: 0.75rem;
  padding: 0.875rem 1rem;
  color: var(--text-secondary, #637082);
  background: var(--surface-subtle, #f5f7fa);
  font-size: 0.8125rem;
  line-height: 1.65;
}
.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}
.field-error {
  color: var(--danger, #b64646);
}
</style>
