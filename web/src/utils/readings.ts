import type { CalculationType } from '@/types/meterItems'

export function formatReadingPrice(cents: number): string {
  return new Intl.NumberFormat('zh-CN', {
    style: 'currency',
    currency: 'CNY',
    minimumFractionDigits: 2,
    maximumFractionDigits: 6,
  }).format(cents / 100)
}

export function readingError(
  calcType: CalculationType,
  current: number | null | undefined,
  previous: number | null | undefined,
): string {
  if (current === undefined || current === null) return ''
  if (!Number.isFinite(current) || current < 0) return '读数必须为大于或等于 0 的数字'
  if (calcType === 'DIFFERENCE') {
    if (previous === undefined || previous === null) return '首次抄表需要填写起始读数'
    if (!Number.isFinite(previous) || previous < 0) return '上次读数必须为大于或等于 0 的数字'
    if (current < previous) return '本次读数不能小于上次读数'
  }
  return ''
}

export function readingUsage(
  calcType: CalculationType,
  current: number | null | undefined,
  previous: number | null | undefined,
): number | undefined {
  if (current === undefined || current === null || readingError(calcType, current, previous))
    return undefined
  return calcType === 'DIRECT' ? current : current - (previous ?? 0)
}
