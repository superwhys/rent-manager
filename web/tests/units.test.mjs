import assert from 'node:assert/strict'
import { register } from 'node:module'
import { afterEach, beforeEach, test } from 'node:test'

register(`data:text/javascript,${encodeURIComponent(`
  export function resolve(specifier, context, nextResolve) {
    if (specifier.startsWith('@/'))
      specifier = new URL(specifier.slice(2) + '.ts', ${JSON.stringify(new URL('../src/', import.meta.url).href)}).href
    return nextResolve(specifier, context)
  }
`)}`, import.meta.url)

const { resetUnits, useUnits } = await import('../src/stores/units.ts')
const originalFetch = globalThis.fetch
const originalStorage = Object.getOwnPropertyDescriptor(globalThis, 'sessionStorage')
const accountA = 'a'.repeat(64)
const accountB = 'b'.repeat(64)
const storageKey = (accountId) => `rent-manager.selected-unit.${accountId}`
const unit = (id) => ({ id, name: `楼栋 ${id}`, access_type: 'owned', is_owner: true })
const store = useUnits()
let saved

function returnUnits(ids) {
  globalThis.fetch = async (url) => {
    assert.equal(url, '/api/units?scope=all')
    return Response.json({ code: 0, data: ids.map(unit) })
  }
}

beforeEach(() => {
  saved = new Map()
  Object.defineProperty(globalThis, 'sessionStorage', {
    configurable: true,
    value: {
      getItem: (key) => saved.get(key) ?? null,
      setItem: (key, value) => saved.set(key, String(value)),
      removeItem: (key) => saved.delete(key),
    },
  })
  resetUnits(accountA)
})

afterEach(() => {
  resetUnits()
  globalThis.fetch = originalFetch
  if (originalStorage) Object.defineProperty(globalThis, 'sessionStorage', originalStorage)
  else delete globalThis.sessionStorage
})

test('刷新后恢复该账号上次选择的楼栋', async () => {
  returnUnits([1, 2])
  await store.loadUnits()
  store.selectedUnitId.value = 2

  resetUnits(accountA)
  assert.equal(store.selectedUnitId.value, undefined)
  await store.loadUnits()
  assert.equal(store.selectedUnitId.value, 2)
  assert.equal(store.selectedUnit.value.name, '楼栋 2')
})

test('普通列表刷新保留当前选择，不受楼栋排序变化影响', async () => {
  returnUnits([1, 2])
  await store.loadUnits()
  store.setSelectedUnit(2)
  returnUnits([3, 2, 1])
  await store.loadUnits()
  assert.equal(store.selectedUnitId.value, 2)
})

test('切换账号和退出会清空当前状态，各账号的选择分别恢复', async () => {
  returnUnits([1, 2])
  await store.loadUnits()
  store.setSelectedUnit(2)

  resetUnits(accountB)
  assert.deepEqual(store.units.value, [])
  assert.equal(store.selectedUnitId.value, undefined)
  returnUnits([3, 4])
  await store.loadUnits()
  assert.equal(store.selectedUnitId.value, 3)
  store.setSelectedUnit(4)

  resetUnits()
  assert.deepEqual(store.units.value, [])
  assert.equal(store.selectedUnitId.value, undefined)
  resetUnits(accountA)
  returnUnits([1, 2])
  await store.loadUnits()
  assert.equal(store.selectedUnitId.value, 2)
  resetUnits(accountB)
  returnUnits([3, 4])
  await store.loadUnits()
  assert.equal(store.selectedUnitId.value, 4)
})

test('缓存楼栋已删除或失去访问权限时回退到可访问列表第一项', async () => {
  saved.set(storageKey(accountA), '2')
  returnUnits([1, 3])
  await store.loadUnits()
  assert.equal(store.selectedUnitId.value, 1)

  store.setSelectedUnit(3)
  returnUnits([1])
  await store.loadUnits()
  assert.equal(store.selectedUnitId.value, 1)
  resetUnits(accountA)
  await store.loadUnits()
  assert.equal(store.selectedUnitId.value, 1)
})

test('无可访问楼栋时清空列表和选择', async () => {
  returnUnits([1, 2])
  await store.loadUnits()
  store.setSelectedUnit(2)
  returnUnits([])
  await store.loadUnits()
  assert.deepEqual(store.units.value, [])
  assert.equal(store.selectedUnitId.value, undefined)
  assert.equal(store.selectedUnit.value, undefined)
})

test('无缓存和非法缓存仍能选择第一项', async () => {
  returnUnits([1, 2])
  for (const value of [null, '', 'wrong', '2.5', '-2', '0', 'Infinity']) {
    resetUnits(accountA)
    if (value === null) saved.delete(storageKey(accountA))
    else saved.set(storageKey(accountA), value)
    await store.loadUnits()
    assert.equal(store.selectedUnitId.value, 1)
    assert.equal(store.error.value, '')
  }
})

test('浏览器存储不可用时仍能加载和切换楼栋', async () => {
  Object.defineProperty(globalThis, 'sessionStorage', {
    configurable: true,
    get() { throw new Error('Storage is unavailable') },
  })
  returnUnits([1, 2])
  await store.loadUnits()
  assert.equal(store.selectedUnitId.value, 1)
  assert.equal(store.error.value, '')
  store.setSelectedUnit(2)
  assert.equal(store.selectedUnitId.value, 2)
})

test('旧账号迟到的成功或失败不能覆盖新账号的楼栋、选择和加载状态', async () => {
  for (const oldResponse of [
    Response.json({ code: 0, data: [unit(1), unit(2)] }),
    Response.json({ code: 207022, data: null, message: '权限不足' }),
  ]) {
    resetUnits(accountA)
    let finishOld
    globalThis.fetch = () => new Promise((resolve) => { finishOld = resolve })
    const pending = store.loadUnits()

    resetUnits(accountB)
    saved.set(storageKey(accountB), '4')
    let finishNew
    globalThis.fetch = () => new Promise((resolve) => { finishNew = resolve })
    const nextPending = store.loadUnits()
    finishOld(oldResponse)
    await pending
    assert.deepEqual(store.units.value, [])
    assert.equal(store.selectedUnitId.value, undefined)
    assert.equal(store.loading.value, true)
    assert.equal(store.error.value, '')

    finishNew(Response.json({ code: 0, data: [unit(3), unit(4)] }))
    await nextPending

    assert.deepEqual(store.units.value.map((item) => item.id), [3, 4])
    assert.equal(store.selectedUnitId.value, 4)
    assert.equal(store.loading.value, false)
    assert.equal(store.error.value, '')
  }
})
