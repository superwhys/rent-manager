import { computed, ref, watch } from 'vue'
import { listUnits } from '@/api/units'
import type { Unit } from '@/types/units'
import { errorMessage } from '@/utils/format'

const units = ref<Unit[]>([])
const selectedUnitId = ref<number>()
const loading = ref(false)
const error = ref('')
const selectionStoragePrefix = 'rent-manager.selected-unit.'
let accountId: string | undefined
let sequence = 0

function cachedUnitId(): number | undefined {
  if (!accountId) return
  try {
    const id = Number(sessionStorage.getItem(`${selectionStoragePrefix}${accountId}`))
    return Number.isSafeInteger(id) && id > 0 ? id : undefined
  } catch {
    return undefined
  }
}

function saveUnitSelection() {
  if (!accountId) return
  try {
    const key = `${selectionStoragePrefix}${accountId}`
    if (selectedUnitId.value === undefined) sessionStorage.removeItem(key)
    else sessionStorage.setItem(key, String(selectedUnitId.value))
  } catch {
    // 浏览器禁用存储时，仍保留当前页面内的选择。
  }
}

watch(selectedUnitId, saveUnitSelection, { flush: 'sync' })

async function loadUnits() {
  const current = ++sequence
  loading.value = true
  error.value = ''
  try {
    const result = await listUnits('all')
    if (current !== sequence) return
    units.value = result
    const preferredId = selectedUnitId.value ?? cachedUnitId()
    selectedUnitId.value = result.some((unit) => unit.id === preferredId)
      ? preferredId
      : result[0]?.id
    if (selectedUnitId.value === undefined) saveUnitSelection()
  } catch (cause) {
    if (current === sequence) error.value = errorMessage(cause)
  } finally {
    if (current === sequence) loading.value = false
  }
}

export function resetUnits(nextAccountId?: string) {
  sequence++
  accountId = undefined
  units.value = []
  selectedUnitId.value = undefined
  loading.value = false
  error.value = ''
  accountId = nextAccountId
}

export function useUnits() {
  return {
    units,
    selectedUnitId,
    loading,
    error,
    loadUnits,
    selectedUnit: computed(() =>
      units.value.find((unit) => unit.id === selectedUnitId.value),
    ),
    setSelectedUnit(id: number) {
      selectedUnitId.value = id
    },
  }
}
