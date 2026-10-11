<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import type { Room, RoomUpdateInput } from '@/types/rooms'
import type { RoomType } from '@/types/roomTypes'
import type { Unit, UnitMeterItem } from '@/types/units'

const props = defineProps<{
  initial?: Room
  unit: Unit
  units: Unit[]
  roomTypes: RoomType[]
  meterItems: UnitMeterItem[]
  saving: boolean
  error: string
}>()
const emit = defineEmits<{ submit: [value: RoomUpdateInput]; cancel: [] }>()
const form = ref<FormInstance>()
const validationError = ref('')
const draft = reactive({
  room_no: props.initial?.room_no ?? '',
  floor: props.initial?.floor ?? 1,
  rentYuan: props.initial ? props.initial.price / 100 : undefined,
  size: props.initial?.size ?? undefined,
  unit_id: props.initial?.unit_id ?? props.unit.id,
  room_type_id: props.initial?.room_type_id ?? undefined,
  is_active: props.initial?.is_active ?? true,
  meter_item_ids:
    props.initial?.meter_items.map((item) => item.id) ??
    props.unit.default_meter_items?.map((item) => item.id) ??
    [],
})
const availableUnits = computed(() =>
  props.units.filter((unit) =>
    unit.id === props.unit.id ||
    (props.unit.account_id && unit.account_id === props.unit.account_id),
  ),
)
const meterOptions = computed(() => {
  const items = new Map(props.meterItems.map((item) => [item.id, item]))
  for (const item of props.unit.default_meter_items ?? []) items.set(item.id, item)
  for (const item of props.initial?.meter_items ?? []) items.set(item.id, item)
  return [...items.values()]
})
const roomTypeOptions = computed(() => {
  if (
    !props.initial ||
    props.roomTypes.some((type) => type.id === props.initial?.room_type_id)
  )
    return props.roomTypes
  return [
    ...props.roomTypes,
    {
      id: props.initial.room_type_id,
      name: `房型 #${props.initial.room_type_id}`,
    },
  ]
})
const rules: FormRules = {
  room_no: [
    {
      required: true,
      whitespace: true,
      message: '请输入房间号',
      trigger: 'blur',
    },
  ],
  floor: [
    {
      required: true,
      type: 'integer',
      min: 0,
      message: '楼层须为不小于 0 的整数',
      trigger: 'change',
    },
  ],
  rentYuan: [
    {
      required: true,
      type: 'number',
      min: props.initial ? 0 : 0.01,
      message: '请输入有效月租金额',
      trigger: 'change',
    },
  ],
  size: [
    {
      required: true,
      type: 'integer',
      min: 1,
      message: '面积须为大于 0 的整数',
      trigger: 'change',
    },
  ],
  room_type_id: [
    {
      required: true,
      type: 'integer',
      min: 1,
      message: '请选择或填写有效房型',
      trigger: 'change',
    },
  ],
}

async function submit() {
  if (props.saving || !(await form.value?.validate().catch(() => false))) return
  validationError.value = ''
  const meterIDs = draft.meter_item_ids
  if (meterIDs.some((id) => !Number.isSafeInteger(id) || id < 1)) {
    validationError.value = '收费项目 ID 须为正整数，多个 ID 用逗号分隔'
    return
  }
  if (
    draft.size === undefined ||
    draft.rentYuan === undefined ||
    draft.room_type_id === undefined
  )
    return
  if (
    !Number.isSafeInteger(draft.floor) ||
    !Number.isSafeInteger(draft.size) ||
    !Number.isSafeInteger(draft.room_type_id)
  ) {
    validationError.value = '楼层、面积和房型 ID 须为有效整数'
    return
  }
  const price = Math.round(draft.rentYuan * 100)
  if (!Number.isSafeInteger(price)) {
    validationError.value = '月租金额超出可用范围'
    return
  }
  emit('submit', {
    room_no: draft.room_no.trim(),
    floor: draft.floor,
    size: draft.size,
    price,
    unit_id: draft.unit_id,
    room_type_id: draft.room_type_id,
    is_active: draft.is_active,
    meter_item_ids: [...new Set(meterIDs)],
  })
}
</script>

<template>
  <el-form
    ref="form"
    :model="draft"
    :rules="rules"
    label-position="top"
    @submit.prevent="submit"
  >
    <el-form-item v-if="initial" label="所属楼栋">
      <el-select v-model="draft.unit_id">
        <el-option
          v-for="item in availableUnits"
          :key="item.id"
          :value="item.id"
          :label="item.name"
        />
      </el-select>
    </el-form-item>
    <div class="form-grid">
      <el-form-item label="房间号" prop="room_no">
        <el-input v-model="draft.room_no" placeholder="例如：301" />
      </el-form-item>
      <el-form-item label="楼层" prop="floor">
        <el-input-number
          v-model="draft.floor"
          :min="0"
          :precision="0"
          controls-position="right"
        />
      </el-form-item>
      <el-form-item label="月租金（元）" prop="rentYuan">
        <el-input-number
          v-model="draft.rentYuan"
          :min="initial ? 0 : 0.01"
          :precision="2"
          :step="0.01"
          controls-position="right"
        />
      </el-form-item>
      <el-form-item label="面积（㎡）" prop="size">
        <el-input-number
          v-model="draft.size"
          :min="1"
          :precision="0"
          controls-position="right"
        />
      </el-form-item>
    </div>
    <el-form-item label="房型" prop="room_type_id">
      <el-select
        v-model="draft.room_type_id"
        placeholder="选择房型"
      >
        <el-option
          v-for="type in roomTypeOptions"
          :key="type.id"
          :value="type.id"
          :label="type.name"
        />
      </el-select>
      <p v-if="unit.is_owner && roomTypes.length === 0" class="muted form-hint">
        请先在“设置 → 房型配置”中添加房型。
      </p>
      <p v-if="!unit.is_owner" class="muted form-hint">
        房型和收费项目由楼栋所有者维护；查看选项需要相应的查看权限。
      </p>
    </el-form-item>
    <el-form-item v-if="initial" label="出租状态">
      <el-radio-group v-model="draft.is_active">
        <el-radio-button :value="true">空置</el-radio-button>
        <el-radio-button :value="false">已出租</el-radio-button>
      </el-radio-group>
    </el-form-item>
    <el-form-item label="收费项目">
      <el-select
        v-model="draft.meter_item_ids"
        multiple
        clearable
        placeholder="选择房间收费项目"
      >
        <el-option
          v-for="item in meterOptions"
          :key="item.id"
          :value="item.id"
          :label="`${item.code} · ${item.unit}`"
        />
      </el-select>
      <p v-if="!initial" class="muted form-hint">
        新房间初始为空置，楼栋的默认收费项目会自动加入。
      </p>
    </el-form-item>
    <el-alert
      v-if="error || validationError"
      :title="error || validationError"
      type="error"
      :closable="false"
      show-icon
    />
    <div class="form-actions">
      <el-button :disabled="saving" @click="emit('cancel')">取消</el-button>
      <el-button type="primary" native-type="submit" :loading="saving"
        >保存房间</el-button
      >
    </div>
  </el-form>
</template>

<style scoped>
.el-input-number {
  width: 100%;
}
.form-hint {
  margin: 0.5rem 0 0;
  font-size: 0.78rem;
  line-height: 1.6;
}
.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.6rem;
  margin-top: 1.5rem;
}
</style>
