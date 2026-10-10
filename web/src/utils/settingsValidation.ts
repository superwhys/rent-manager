import type { SettingInput } from '@/types/settings'

export function settingError(value: SettingInput): string {
  if (!value.key.trim()) return '请输入配置键'
  if (!value.value.trim()) return '请输入配置值'
  const encoder = new TextEncoder()
  if (encoder.encode(value.key.trim()).length > 64)
    return '配置键不能超过 64 字节'
  if (encoder.encode(value.value.trim()).length > 255)
    return '配置值不能超过 255 字节'
  return ''
}

export function yuanToUnitPrice(value: number): number | undefined {
  if (
    !Number.isFinite(value) ||
    value < 0 ||
    value > Number.MAX_SAFE_INTEGER / 100
  )
    return undefined
  // Preserve fractional cents while removing ordinary binary multiplication noise.
  return Number((value * 100).toPrecision(15))
}
