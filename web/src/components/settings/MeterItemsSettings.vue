<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import AsyncState from '@/components/common/AsyncState.vue'
import { yuanToUnitPrice } from '@/utils/settingsValidation'
import type {
  CalculationType,
  MeterItem,
  MeterItemInput,
} from '@/types/meterItems'

const props = defineProps<{
  items: MeterItem[]
  loading: boolean
  error: string
  busy: boolean
  actionError: string
  savedVersion: number
  offset: number
  hasNext: boolean
}>()
const emit = defineEmits<{
  retry: []
  page: [offset: number]
  save: [id: number | null, value: MeterItemInput]
  clearError: []
}>()
const opened = ref(false)
const editingId = ref<number | null>(null)
const originalRemark = ref('')
const draft = reactive({
  code: '',
  unit: '',
  unit_price: 0,
  calc_type: 'DIFFERENCE' as CalculationType,
  remark: '',
})
const canSave = computed(
  () =>
    !props.busy &&
    !!draft.code.trim() &&
    !!draft.unit.trim() &&
    yuanToUnitPrice(draft.unit_price) !== undefined &&
    (!originalRemark.value || !!draft.remark.trim()) &&
    [...draft.code.trim()].length <= 255 &&
    [...draft.unit.trim()].length <= 255,
)
const priceFormatter = new Intl.NumberFormat('zh-CN', {
  maximumFractionDigits: 6,
})
function edit(item?: MeterItem) {
  emit('clearError')
  editingId.value = item?.id ?? null
  originalRemark.value = item?.remark ?? ''
  Object.assign(draft, {
    code: item?.code ?? '',
    unit: item?.unit ?? '',
    unit_price: item ? item.unit_price / 100 : 0,
    calc_type: item?.calc_type ?? 'DIFFERENCE',
    remark: item?.remark ?? '',
  })
  opened.value = true
}
function save() {
  const unitPrice = yuanToUnitPrice(draft.unit_price)
  if (!canSave.value || unitPrice === undefined) return
  emit('save', editingId.value, {
    code: draft.code.trim(),
    unit: draft.unit.trim(),
    unit_price: unitPrice,
    calc_type: draft.calc_type,
    remark: draft.remark.trim() || null,
  })
}
watch(
  () => props.savedVersion,
  () => {
    opened.value = false
  },
)
</script>

<template>
  <section class="panel meter-items-panel">
    <div class="toolbar">
      <div>
        <h2>抄表与计费项</h2>
        <p class="muted">设置计量单位与单价，在房间和单元中使用。</p>
      </div>
      <button class="btn btn-primary" :disabled="busy" @click="edit()">
        新增计费项
      </button>
    </div>
    <AsyncState
      :loading="loading"
      :error="error"
      :empty="!items.length"
      empty-title="暂无计费项"
      empty-description="添加水费、电费等计费项，为抄表和账单做好准备。"
      @retry="emit('retry')"
    >
      <div class="meter-list">
        <article v-for="item in items" :key="item.id" class="meter-row">
          <div class="meter-mark" aria-hidden="true">
            {{ item.calc_type === 'DIFFERENCE' ? '↗' : '＝' }}
          </div>
          <div class="meter-info">
            <h3>
              {{ item.code }}
              <span v-if="item.is_system" class="badge">系统项</span>
            </h3>
            <p class="muted">
              {{ item.calc_type === 'DIFFERENCE' ? '差值计费' : '直接计费'
              }}<span v-if="item.remark"> · {{ item.remark }}</span>
            </p>
          </div>
          <p class="meter-price">
            <strong>¥{{ priceFormatter.format(item.unit_price / 100) }}</strong
            ><span class="muted"> / {{ item.unit }}</span>
          </p>
          <button
            class="btn btn-secondary"
            :disabled="busy"
            @click="edit(item)"
          >
            编辑
          </button>
        </article>
      </div>
    </AsyncState>
    <div v-if="offset || hasNext" class="pagination">
      <button
        class="btn btn-secondary"
        :disabled="loading || busy || !offset"
        @click="emit('page', Math.max(0, offset - 20))"
      >
        上一页
      </button>
      <span class="muted">第 {{ Math.floor(offset / 20) + 1 }} 页</span>
      <button
        class="btn btn-secondary"
        :disabled="loading || busy || !hasNext"
        @click="emit('page', offset + 20)"
      >
        下一页
      </button>
    </div>
  </section>
  <el-dialog
    v-model="opened"
    :title="editingId ? '编辑计费项' : '新增计费项'"
    width="540px"
    :close-on-click-modal="false"
    :show-close="!busy"
    :close-on-press-escape="!busy"
  >
    <form class="meter-form" @submit.prevent="save">
      <label class="field">
        <span>名称 / 编码</span>
        <input
          v-model="draft.code"
          maxlength="255"
          required
          placeholder="例如：水费"
          :disabled="busy"
        />
      </label>
      <div class="form-grid">
        <label class="field">
          <span>计量单位</span>
          <input
            v-model="draft.unit"
            maxlength="255"
            required
            placeholder="例如：吨、度"
            :disabled="busy"
          />
        </label>
        <label class="field">
          <span>单价（元）</span>
          <input
            v-model.number="draft.unit_price"
            type="number"
            min="0"
            step="any"
            required
            :disabled="busy"
          />
        </label>
      </div>
      <label class="field">
        <span>计费方式</span>
        <select v-model="draft.calc_type" :disabled="busy">
          <option value="DIFFERENCE">差值计费</option>
          <option value="DIRECT">直接计费</option>
        </select>
        <small class="muted">{{
          draft.calc_type === 'DIFFERENCE'
            ? '本次读数减去上次读数，计算实际用量。'
            : '以本次录入的数值作为用量。'
        }}</small>
      </label>
      <label class="field">
        <span>备注（选填）</span>
        <textarea v-model="draft.remark" rows="2" :disabled="busy" />
      </label>
      <p
        v-if="originalRemark && !draft.remark.trim()"
        class="form-error"
        role="alert"
      >
        已填写的备注暂不支持清空，可以修改备注内容。
      </p>
      <p
        v-if="yuanToUnitPrice(draft.unit_price) === undefined"
        class="form-error"
        role="alert"
      >
        请输入有效的非负单价。
      </p>
      <p v-if="actionError" class="form-error" role="alert">
        {{ actionError }}
      </p>
      <div class="dialog-actions">
        <button
          type="button"
          class="btn btn-secondary"
          :disabled="busy"
          @click="opened = false"
        >
          取消
        </button>
        <button type="submit" class="btn btn-primary" :disabled="!canSave">
          {{ busy ? '正在保存…' : '保存计费项' }}
        </button>
      </div>
    </form>
  </el-dialog>
</template>

<style scoped>
.toolbar {
  justify-content: space-between;
}
.meter-items-panel {
  padding: 1.5rem;
}
.toolbar h2 {
  margin: 0 0 0.35rem;
}
.toolbar p {
  margin: 0;
}
.meter-list {
  margin-top: 1.25rem;
}
.meter-row {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1.1rem 0;
  border-top: 1px solid var(--border, #e8e9ed);
}
.meter-mark {
  display: grid;
  place-items: center;
  width: 2.5rem;
  height: 2.5rem;
  border-radius: 0.75rem;
  color: var(--accent);
  background: var(--accent-soft);
  font-size: 1.3rem;
  flex: 0 0 auto;
}
.meter-info {
  flex: 1;
  min-width: 0;
}
.meter-info h3 {
  margin: 0;
  font-size: 0.95rem;
  overflow-wrap: anywhere;
}
.meter-info p {
  margin: 0.35rem 0 0;
  font-size: 0.82rem;
  overflow-wrap: anywhere;
}
.meter-price {
  flex: 0 1 auto;
  font-size: 0.9rem;
  overflow-wrap: anywhere;
  text-align: right;
}
.meter-price strong {
  font-variant-numeric: tabular-nums;
  font-weight: 600;
}
.meter-form {
  display: grid;
  gap: 1.2rem;
}
.dialog-actions,
.pagination {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.75rem;
  flex-wrap: wrap;
}
.pagination {
  margin-top: 1.25rem;
}
.form-error {
  color: var(--danger);
  font-size: 0.875rem;
}
@media (max-width: 640px) {
  .meter-items-panel {
    padding: 1rem;
  }
  .meter-row {
    flex-wrap: wrap;
  }
  .meter-info {
    flex-basis: calc(100% - 3.5rem);
  }
  .meter-price {
    flex: 1;
    text-align: left;
    margin-left: 3.5rem;
  }
}
</style>
