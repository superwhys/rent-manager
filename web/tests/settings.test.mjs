import assert from 'node:assert/strict'
import test from 'node:test'
import {
  settingError,
  yuanToUnitPrice,
} from '../src/utils/settingsValidation.ts'
import { isAccountId } from '../src/utils/accessValidation.ts'

test('账号配置按 UTF-8 字节验证，中文不按单字符计算', () => {
  assert.equal(
    settingError({ key: '中'.repeat(21) + 'a', value: '中'.repeat(85) }),
    '',
  )
  assert.match(settingError({ key: '中'.repeat(22), value: 'yes' }), /64 字节/)
  assert.match(settingError({ key: 'a', value: '中'.repeat(86) }), /255 字节/)
  assert.match(settingError({ key: ' ', value: 'value' }), /配置键/)
  assert.match(settingError({ key: 'key', value: ' ' }), /配置值/)
})

test('计费单价从元换算为分，保留零值与小数分', () => {
  assert.equal(yuanToUnitPrice(2), 200)
  assert.equal(yuanToUnitPrice(0), 0)
  assert.equal(yuanToUnitPrice(0.29), 29)
  assert.equal(yuanToUnitPrice(0.005), 0.5)
  for (const value of [-1, NaN, Infinity, -Infinity, Number.MAX_VALUE]) {
    assert.equal(yuanToUnitPrice(value), undefined)
  }
})

test('授权账号必须是完整的 64 位小写十六进制 ID', () => {
  assert.equal(isAccountId('a'.repeat(64)), true)
  for (const value of [
    '',
    'a'.repeat(63),
    'a'.repeat(65),
    'A'.repeat(64),
    'g'.repeat(64),
    ` ${'a'.repeat(64)}`,
  ]) {
    assert.equal(isAccountId(value), false)
  }
})
