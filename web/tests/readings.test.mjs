import assert from 'node:assert/strict'
import test from 'node:test'
import { formatReadingPrice, readingError, readingUsage } from '../src/utils/readings.ts'

test('单价保留小数分，避免把精细费率显示为零', () => {
  assert.match(formatReadingPrice(0.5), /0\.005/)
  assert.match(formatReadingPrice(123.456), /1\.23456/)
})

test('差值抄表保留合法零值，并计算实际用量', () => {
  assert.equal(readingError('DIFFERENCE', 0, 0), '')
  assert.equal(readingUsage('DIFFERENCE', 0, 0), 0)
  assert.equal(readingUsage('DIFFERENCE', 25.5, 10.5), 15)
})

test('首次差值抄表必须填写表底，不能倒退或录入负值', () => {
  assert.match(readingError('DIFFERENCE', 10, undefined), /起始读数/)
  assert.match(readingError('DIFFERENCE', 10, null), /起始读数/)
  assert.match(readingError('DIFFERENCE', 9, 10), /不能小于/)
  assert.match(readingError('DIFFERENCE', 1, -1), /上次读数/)
  assert.match(readingError('DIFFERENCE', -1, 0), /读数必须/)
  assert.equal(readingUsage('DIFFERENCE', 9, 10), undefined)
})

test('直接计费不要求表底，空行可以稍后填写', () => {
  assert.equal(readingError('DIRECT', 0, undefined), '')
  assert.equal(readingUsage('DIRECT', 12.75, undefined), 12.75)
  assert.equal(readingError('DIFFERENCE', undefined, undefined), '')
  assert.equal(readingUsage('DIRECT', null, undefined), undefined)
})

test('拒绝不能提交为 JSON 数字的非有限数值', () => {
  for (const value of [NaN, Infinity, -Infinity]) {
    assert.match(readingError('DIRECT', value, undefined), /读数必须/)
    assert.match(readingError('DIFFERENCE', 10, value), /上次读数/)
  }
})
