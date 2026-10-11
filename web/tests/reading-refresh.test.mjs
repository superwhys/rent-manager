import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { register } from 'node:module'
import { afterEach, beforeEach, test } from 'node:test'
import ts from 'typescript'
import * as Vue from 'vue'
import { compileScript, parse } from 'vue/compiler-sfc'

register(`data:text/javascript,${encodeURIComponent(`
  export function resolve(specifier, context, nextResolve) {
    if (specifier.startsWith('@/'))
      specifier = new URL(specifier.slice(2) + '.ts', ${JSON.stringify(new URL('../src/', import.meta.url).href)}).href
    return nextResolve(specifier, context)
  }
`)}`, import.meta.url)

const readings = await import('../src/api/readings.ts')
const units = await import('../src/stores/units.ts')
const format = await import('../src/utils/format.ts')
const filename = new URL('../src/views/readings/MeterReadingView.vue', import.meta.url)
const { descriptor } = parse(await readFile(filename, 'utf8'))
const script = compileScript(descriptor, { id: 'reading-refresh-test' })
const { outputText } = ts.transpileModule(script.content, {
  compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 },
})
const componentModule = { exports: {} }
const imports = {
  vue: Vue,
  'element-plus': { ElMessage: { success() {} } },
  '@/api/readings': readings,
  '@/stores/units': units,
  '@/utils/format': format,
}
new Function('require', 'module', 'exports', outputText)(
  (specifier) => {
    if (specifier.endsWith('.vue')) return {}
    assert.ok(specifier in imports, `未处理的组件导入：${specifier}`)
    return imports[specifier]
  },
  componentModule,
  componentModule.exports,
)
const component = { ...componentModule.exports.default, render: () => null }
const renderer = Vue.createRenderer({
  insert() {}, remove() {}, createElement: () => ({}), createText: () => ({}),
  createComment: () => ({}), setText() {}, setElementText() {}, parentNode: () => null,
  nextSibling: () => null, patchProp() {},
})
const originalFetch = globalThis.fetch
const store = units.useUnits()
let container
let state
let requests

beforeEach(() => {
  units.resetUnits()
  store.setSelectedUnit(1)
  requests = []
  globalThis.fetch = (url) => new Promise((resolve) => { requests.push({ url, resolve }) })
  container = {}
  const vnode = Vue.createVNode(component)
  renderer.render(vnode, container)
  state = vnode.component.setupState
})

afterEach(() => {
  renderer.render(null, container)
  units.resetUnits()
  globalThis.fetch = originalFetch
})

function result(totalFee) {
  return { rooms: [], summaries: [], missing_count: 0, total_fee: totalFee }
}

async function finish(index, data) {
  requests[index].resolve(Response.json({ code: 0, data }))
  await new Promise((resolve) => setImmediate(resolve))
  await Vue.nextTick()
}

test('初次加载无数据，切换房间范围保留已有结果直到新结果到达', async () => {
  assert.equal(state.loading, true)
  assert.equal(state.result, undefined)
  await finish(0, result(100))
  const previous = state.result
  assert.equal(state.loading, false)

  state.includeVacantRecords = true
  await Vue.nextTick()
  assert.equal(requests.length, 2)
  assert.match(requests[1].url, /include_vacant_records=true/)
  assert.equal(state.loading, true)
  assert.equal(state.result, previous)

  await finish(1, result(200))
  assert.equal(state.loading, false)
  assert.equal(state.result.total_fee, 200)
})

test('手动刷新保留已有结果，刷新期间禁止打开旧房间编辑器', async () => {
  await finish(0, result(100))
  const previous = state.result
  const refresh = state.loadReadings()
  assert.equal(state.loading, true)
  assert.equal(state.result, previous)
  void state.openEditor({ room_id: 1, period_month: state.month, readings: [{}] })
  assert.equal(state.editorOpen, false)
  assert.equal(requests.length, 2)
  await finish(1, result(200))
  await refresh
  assert.equal(state.loading, false)
  assert.equal(state.result.total_fee, 200)
})

test('切换楼栋或月份清空旧结果，避免展示另一账期的抄表数据', async () => {
  await finish(0, result(100))
  store.setSelectedUnit(2)
  await Vue.nextTick()
  assert.equal(state.loading, true)
  assert.equal(state.result, undefined)
  assert.match(requests[1].url, /^\/api\/units\/2\/meter-readings/)
  await finish(1, result(200))

  state.month = state.month === '2026-01' ? '2026-02' : '2026-01'
  await Vue.nextTick()
  assert.equal(state.loading, true)
  assert.equal(state.result, undefined)
  assert.match(requests[2].url, new RegExp(`period_month=${state.month}`))
  await finish(2, result(300))
  assert.equal(state.result.total_fee, 300)
})

test('快速切换范围，迟到的成功和失败都不能覆盖最新结果与加载状态', async () => {
  await finish(0, result(100))
  state.includeVacantRecords = true
  await Vue.nextTick()
  state.includeVacantRecords = false
  await Vue.nextTick()
  await finish(1, result(200))
  assert.equal(state.result.total_fee, 100)
  assert.equal(state.loading, true)
  await finish(2, result(300))
  assert.equal(state.result.total_fee, 300)

  state.includeVacantRecords = true
  await Vue.nextTick()
  state.includeVacantRecords = false
  await Vue.nextTick()
  await finish(4, result(400))
  requests[3].resolve(Response.json({ code: 207022, message: '权限不足' }))
  await new Promise((resolve) => setImmediate(resolve))
  assert.equal(state.result.total_fee, 400)
  assert.equal(state.loading, false)
  assert.equal(state.error, '')
})
