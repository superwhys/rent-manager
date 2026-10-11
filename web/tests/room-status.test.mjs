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

const roomsApi = await import('../src/api/rooms.ts')
const roomTypesApi = await import('../src/api/roomTypes.ts')
const meterItemsApi = await import('../src/api/meterItems.ts')
const units = await import('../src/stores/units.ts')
const format = await import('../src/utils/format.ts')
const filename = new URL('../src/views/rooms/RoomListView.vue', import.meta.url)
const { descriptor } = parse(await readFile(filename, 'utf8'))
const script = compileScript(descriptor, { id: 'room-status-test' })
const { outputText } = ts.transpileModule(script.content, {
  compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 },
})
const messages = { success: [], warning: [], error: [] }
const componentModule = { exports: {} }
const imports = {
  vue: Vue,
  'element-plus': {
    ElMessage: Object.fromEntries(Object.keys(messages).map((kind) => [
      kind, (message) => messages[kind].push(message),
    ])),
    ElMessageBox: { confirm: async () => {} },
  },
  '@/api/rooms': roomsApi,
  '@/api/roomTypes': roomTypesApi,
  '@/api/meterItems': meterItemsApi,
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
const buildings = [
  { id: 1, name: '一号楼', is_owner: true, room_count: 2, rented_count: 1 },
  { id: 2, name: '二号楼', is_owner: true, room_count: 1, rented_count: 0 },
]
let container
let state
let requests

function room(id, isActive = true, unitId = 1) {
  return {
    id, unit_id: unitId, room_no: String(100 + id), floor: 1, price: 500,
    size: 10, room_type_id: 1, is_active: isActive,
    meter_items: [{ id: 3, code: '水费' }],
  }
}

async function flush() {
  await new Promise((resolve) => setImmediate(resolve))
  await Vue.nextTick()
}

function pending(method, path) {
  const request = requests.find((item) =>
    !item.finished && item.options.method === method && item.url.split('?')[0] === path,
  )
  assert.ok(request, `缺少请求：${method} ${path}`)
  return request
}

async function finish(request, data, message) {
  request.finished = true
  request.resolve(Response.json(message ? { code: 207022, message } : { code: 0, data }))
  await flush()
}

async function refresh(data, roomError, unitError) {
  const roomRequest = pending('GET', '/api/units/1/rooms')
  const unitRequest = pending('GET', '/api/units')
  await finish(roomRequest, data, roomError)
  await finish(unitRequest, buildings, unitError)
}

beforeEach(async () => {
  units.resetUnits()
  store.units.value = buildings.map((unit) => ({ ...unit }))
  store.setSelectedUnit(1)
  for (const values of Object.values(messages)) values.length = 0
  requests = []
  globalThis.fetch = (url, options) => new Promise((resolve) => {
    requests.push({ url, options, resolve, finished: false })
  })
  container = {}
  const vnode = Vue.createVNode(component)
  renderer.render(vnode, container)
  state = vnode.component.setupState
  await finish(pending('GET', '/api/units/1/rooms'), [room(1), room(2, false)])
  await finish(pending('GET', '/api/room-types'), [])
  await finish(pending('GET', '/api/meter-items'), [])
  requests = []
})

afterEach(() => {
  renderer.render(null, container)
  units.resetUnits()
  globalThis.fetch = originalFetch
})

for (const initialStatus of [true, false]) {
  test(`${initialStatus ? '空置改为已出租' : '已出租改为空置'}只提交出租状态并刷新房间与统计`, async () => {
    const current = state.rooms.find((item) => item.is_active === initialStatus)
    const action = state.toggleRoomStatus(current)
    const write = pending('PUT', `/api/rooms/${current.id}`)
    assert.deepEqual(JSON.parse(write.options.body), { is_active: !initialStatus })
    assert.equal(state.changingStatus, current.id)
    assert.equal(current.is_active, initialStatus)

    await finish(write, { ...current, is_active: !initialStatus })
    await refresh([room(1, false), room(2)])
    await action
    assert.equal(state.rooms.find((item) => item.id === current.id).is_active, !initialStatus)
    assert.equal(state.changingStatus, undefined)
    assert.equal(messages.success.length, 1)
    assert.deepEqual(messages.error, [])
    assert.deepEqual(messages.warning, [])
  })
}

test('提交及刷新期间重复点击其他房间也不会再次提交', async () => {
  const current = state.rooms[0]
  const action = state.toggleRoomStatus(current)
  await state.toggleRoomStatus(current)
  await state.toggleRoomStatus(state.rooms[1])
  state.openForm(current)
  assert.equal(state.formVisible, false)
  await state.removeRoom(current)
  state.changePage(1)
  assert.equal(state.page, 0)
  assert.equal(requests.length, 1)
  assert.equal(requests.filter((item) => item.options.method === 'PUT').length, 1)
  assert.equal(requests.filter((item) => item.options.method === 'DELETE').length, 0)

  await finish(pending('PUT', '/api/rooms/1'), room(1, false))
  await state.toggleRoomStatus(current)
  await state.toggleRoomStatus(state.rooms[1])
  assert.equal(requests.filter((item) => item.options.method === 'PUT').length, 1)
  assert.equal(state.changingStatus, 1)
  await refresh([room(1, false), room(2, false)])
  await action
  assert.equal(state.changingStatus, undefined)
})

test('保存、删除或加载期间禁止切换状态', async () => {
  const current = state.rooms[0]
  state.saving = true
  await state.toggleRoomStatus(current)
  state.saving = false
  state.deleting = current.id
  await state.toggleRoomStatus(current)
  state.deleting = undefined
  state.loading = true
  await state.toggleRoomStatus(current)
  state.loading = false
  store.loading.value = true
  await state.toggleRoomStatus(current)
  store.loading.value = false
  state.formVisible = true
  await state.toggleRoomStatus(current)
  state.formVisible = false
  await state.toggleRoomStatus(room(21, true, 2))
  assert.equal(requests.length, 0)
})

test('写入失败保留原状态并解除锁，重试可以成功', async () => {
  const current = state.rooms[0]
  const action = state.toggleRoomStatus(current)
  await finish(pending('PUT', '/api/rooms/1'), undefined, '没有管理房间的权限')
  await action
  assert.equal(state.rooms[0].is_active, true)
  assert.equal(state.changingStatus, undefined)
  assert.deepEqual(messages.error, ['没有管理房间的权限'])
  assert.equal(requests.length, 1)

  const retry = state.toggleRoomStatus(current)
  await finish(pending('PUT', '/api/rooms/1'), room(1, false))
  await refresh([room(1, false), room(2, false)])
  await retry
  assert.equal(state.rooms[0].is_active, false)
  assert.equal(state.changingStatus, undefined)
})

test('筛选页最后一间房切换后移出结果时回到上一页', async () => {
  state.statusFilter = 'vacant'
  await Vue.nextTick()
  await finish(pending('GET', '/api/units/1/rooms'), [room(1)])
  state.page = 1
  const action = state.toggleRoomStatus(state.rooms[0])
  state.changePage(-1)
  assert.equal(state.page, 1)
  state.changePage(1)
  assert.equal(state.page, 1)
  await finish(pending('PUT', '/api/rooms/1'), room(1, false))
  assert.equal(state.page, 0)
  const reload = pending('GET', '/api/units/1/rooms')
  assert.equal(new URL(reload.url, 'http://localhost').searchParams.get('offset'), '0')
  assert.equal(new URL(reload.url, 'http://localhost').searchParams.get('is_active'), 'true')
  await refresh([room(3)])
  await action
  assert.equal(state.rooms[0].id, 3)
})

test('全部房间筛选下切换状态保持当前页', async () => {
  state.rooms = [room(1)]
  state.page = 1
  const action = state.toggleRoomStatus(state.rooms[0])
  await finish(pending('PUT', '/api/rooms/1'), room(1, false))
  assert.equal(state.page, 1)
  assert.equal(new URL(pending('GET', '/api/units/1/rooms').url, 'http://localhost').searchParams.get('offset'), '20')
  await refresh([room(1, false)])
  await action
})

for (const failedRefresh of ['房间', '楼栋']) {
  test(`${failedRefresh}刷新失败提示状态已保存，允许再次操作`, async () => {
    const action = state.toggleRoomStatus(state.rooms[0])
    await finish(pending('PUT', '/api/rooms/1'), room(1, false))
    await refresh([room(1, false)], failedRefresh === '房间' ? '刷新失败' : undefined, failedRefresh === '楼栋' ? '刷新失败' : undefined)
    await action
    assert.equal(state.changingStatus, undefined)
    assert.equal(state.rooms[0].is_active, false)
    assert.deepEqual(messages.error, [])
    assert.equal(messages.warning.length, 1)
    assert.match(messages.warning[0], /出租状态已(?:保存|更新).*刷新失败/)
  })
}

test('写入期间切换楼栋，完成后只刷新楼栋统计而保留新楼栋房间', async () => {
  const action = state.toggleRoomStatus(state.rooms[0])
  const write = pending('PUT', '/api/rooms/1')
  store.setSelectedUnit(2)
  await Vue.nextTick()
  await finish(pending('GET', '/api/units/2/rooms'), [room(21, true, 2)])
  await finish(pending('GET', '/api/room-types'), [])
  await finish(pending('GET', '/api/meter-items'), [])
  const count = requests.length
  await finish(write, room(1, false))
  assert.equal(requests.length, count + 1)
  const updatedBuildings = buildings.map((unit) =>
    unit.id === 1 ? { ...unit, rented_count: 2 } : unit,
  )
  await finish(pending('GET', '/api/units'), updatedBuildings)
  await action
  assert.equal(requests.length, count + 1)
  assert.equal(store.units.value.find((unit) => unit.id === 1).rented_count, 2)
  assert.equal(state.selectedUnitId, 2)
  assert.equal(state.rooms[0].id, 21)
  assert.equal(state.rooms[0].is_active, true)
  assert.equal(state.changingStatus, undefined)
  assert.deepEqual(messages.success, [])
})

test('刷新期间切换楼栋，迟到的旧房间结果不能覆盖新楼栋', async () => {
  const action = state.toggleRoomStatus(state.rooms[0])
  await finish(pending('PUT', '/api/rooms/1'), room(1, false))
  const oldReload = pending('GET', '/api/units/1/rooms')
  store.setSelectedUnit(2)
  await Vue.nextTick()
  await finish(pending('GET', '/api/units/2/rooms'), [room(21, true, 2)])
  await finish(pending('GET', '/api/room-types'), [])
  await finish(pending('GET', '/api/meter-items'), [])
  await finish(oldReload, [room(1, false)])
  await finish(pending('GET', '/api/units'), buildings)
  await action
  assert.equal(state.rooms[0].id, 21)
  assert.equal(state.selectedUnitId, 2)
  assert.equal(state.changingStatus, undefined)
})

test('写入完成前离开页面，不会继续刷新或显示成功提示', async () => {
  const action = state.toggleRoomStatus(state.rooms[0])
  renderer.render(null, container)
  await finish(pending('PUT', '/api/rooms/1'), room(1, false))
  await action
  assert.equal(requests.length, 1)
  assert.deepEqual(messages.success, [])
  assert.deepEqual(messages.error, [])
  assert.deepEqual(messages.warning, [])
})
