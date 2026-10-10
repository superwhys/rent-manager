<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import type { Unit, UnitInput, UnitMeterItem } from '@/types/units'

const props = defineProps<{
  initial?: Unit
  meterItems: UnitMeterItem[]
  saving: boolean
  error: string
}>()

const emit = defineEmits<{
  submit: [value: UnitInput]
  cancel: []
}>()

const form = ref<FormInstance>()
const validationError = ref('')
const draft = reactive<UnitInput>({
  name: props.initial?.name ?? '',
  province: props.initial?.province ?? '',
  city: props.initial?.city ?? '',
  address: props.initial?.address ?? '',
  default_meter_item_ids:
    props.initial?.default_meter_items?.map((item) => item.id) ?? [],
  json_extend: {
    meter_reading_day: props.initial?.json_extend?.meter_reading_day ?? 1,
  },
})
const meterOptions = computed(() => {
  const items = new Map(props.meterItems.map((item) => [item.id, item]))
  for (const item of props.initial?.default_meter_items ?? [])
    items.set(item.id, item)
  return [...items.values()]
})
const rules: FormRules = {
  name: [
    {
      required: true,
      whitespace: true,
      message: '请输入楼栋名称',
      trigger: 'blur',
    },
  ],
  province: [
    {
      required: true,
      whitespace: true,
      message: '请输入省份',
      trigger: 'blur',
    },
  ],
  city: [
    {
      required: true,
      whitespace: true,
      message: '请输入城市',
      trigger: 'blur',
    },
  ],
  address: [
    {
      required: true,
      whitespace: true,
      message: '请输入详细地址',
      trigger: 'blur',
    },
  ],
  'json_extend.meter_reading_day': [
    {
      required: true,
      type: 'number',
      min: 1,
      max: 31,
      message: '请选择每月 1–31 日',
      trigger: 'change',
    },
  ],
}

async function submit() {
  if (props.saving || !(await form.value?.validate().catch(() => false))) return
  validationError.value = ''
  const meterIDs = draft.default_meter_item_ids
  if (meterIDs.some((id) => !Number.isSafeInteger(id) || id < 1)) {
    validationError.value = '收费项目 ID 须为正整数，多个 ID 用逗号分隔'
    return
  }
  emit('submit', {
    ...draft,
    name: draft.name.trim(),
    province: draft.province.trim(),
    city: draft.city.trim(),
    address: draft.address.trim(),
    default_meter_item_ids: [...new Set(meterIDs)],
    json_extend: { ...draft.json_extend },
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
    <el-form-item label="楼栋名称" prop="name">
      <el-input
        v-model="draft.name"
        maxlength="255"
        placeholder="例如：梧桐公寓"
      />
    </el-form-item>
    <div class="form-grid">
      <el-form-item label="省份" prop="province">
        <el-input
          v-model="draft.province"
          maxlength="255"
          placeholder="例如：广东省"
        />
      </el-form-item>
      <el-form-item label="城市" prop="city">
        <el-input
          v-model="draft.city"
          maxlength="255"
          placeholder="例如：深圳市"
        />
      </el-form-item>
    </div>
    <el-form-item label="详细地址" prop="address">
      <el-input
        v-model="draft.address"
        maxlength="255"
        placeholder="区、街道及门牌号"
      />
    </el-form-item>
    <el-form-item label="每月抄表日" prop="json_extend.meter_reading_day">
      <el-input-number
        v-model="draft.json_extend.meter_reading_day"
        :min="1"
        :max="31"
        :precision="0"
      />
      <span class="field-note">日</span>
    </el-form-item>
    <el-form-item label="默认收费项目">
      <el-select
        v-model="draft.default_meter_item_ids"
        multiple
        clearable
        placeholder="可不选择"
      >
        <el-option
          v-for="item in meterOptions"
          :key="item.id"
          :label="`${item.code} · ${item.unit}`"
          :value="item.id"
        />
      </el-select>
      <p v-if="initial && !initial.is_owner" class="muted form-hint">
        收费项目由楼栋所有者维护；查看选项需要收费项目查看权限。
      </p>
      <p class="muted form-hint">
        新建房间会自动使用这些收费项目。修改后不影响已有房间。
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
        >保存楼栋</el-button
      >
    </div>
  </el-form>
</template>

<style scoped>
.field-note {
  margin-left: 0.6rem;
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
