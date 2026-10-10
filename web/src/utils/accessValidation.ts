export function isAccountId(value: string): boolean {
  return /^[0-9a-f]{64}$/.test(value)
}
