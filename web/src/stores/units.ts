import { computed, ref } from 'vue'
import { listUnits } from '@/api/units'
import type { Unit } from '@/types/units'
import { errorMessage } from '@/utils/format'

const units = ref<Unit[]>([])
const selectedUnitId = ref<number>()
const loading = ref(false)
const error = ref('')
let sequence = 0

async function loadUnits() {
  const current = ++sequence
  loading.value = true
  error.value = ''
  try {
    const result = await listUnits('all')
    if (current !== sequence) return
    units.value = result
    if (!result.some((unit) => unit.id === selectedUnitId.value))
      selectedUnitId.value = result[0]?.id
  } catch (cause) {
    if (current === sequence) error.value = errorMessage(cause)
  } finally {
    if (current === sequence) loading.value = false
  }
}

export function resetUnits() {
  sequence++
  units.value = []
  selectedUnitId.value = undefined
  loading.value = false
  error.value = ''
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
