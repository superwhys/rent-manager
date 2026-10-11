import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { afterEach, beforeEach, test } from 'node:test'
import ts from 'typescript'
import * as Vue from 'vue'
import { compileScript, parse } from 'vue/compiler-sfc'
import * as format from '../src/utils/format.ts'

const filename = new URL('../src/views/bills/BillListView.vue', import.meta.url)
const { descriptor } = parse(await readFile(filename, 'utf8'))
const script = compileScript(descriptor, { id: 'bill-export-test' })
const { outputText } = ts.transpileModule(script.content, {
  compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 },
})
const selectedUnitId = Vue.ref(1)
const pngUrl = 'data:image/png;base64,YmlsbA=='
let renderPng
let getBillDetail
let fontsReady
let pngCalls
let downloads
let links
let messages

function bill(id = 17) {
  return {
    id,
    room_id: 3,
    period_month: '2026-10',
    base_rent: 100000,
    meter_fee: 1000,
    total_amount: 101000,
    status: 'UNPAID',
    created_at: '2026-10-01T00:00:00Z',
    updated_at: '2026-10-01T00:00:00Z',
    meter_readings: [],
  }
}

const componentModule = { exports: {} }
const imports = {
  vue: Vue,
  'element-plus': {
    ElMessage: {
      success() {},
      error(message) { messages.push(message) },
    },
  },
  '@/api/bills': {
    listUnitBills: async () => [],
    listRoomBills: async () => [],
    getBill: (id) => getBillDetail(id),
  },
  '@/api/rooms': { listRooms: async () => [] },
  '@/stores/units': { useUnits: () => ({ selectedUnitId }) },
  '@/utils/format': format,
  'html-to-image': {
    toPng(node, options) {
      pngCalls.push({ node, options })
      return renderPng(node, options)
    },
  },
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
const originalDocument = Object.getOwnPropertyDescriptor(globalThis, 'document')
let container
let state

beforeEach(() => {
  selectedUnitId.value = 1
  pngCalls = []
  downloads = []
  links = []
  messages = []
  renderPng = async () => pngUrl
  getBillDetail = async (id) => bill(id)
  fontsReady = Promise.resolve()
  Object.defineProperty(globalThis, 'document', {
    configurable: true,
    writable: true,
    value: {
      fonts: { get ready() { return fontsReady } },
      body: {
        appendChild(link) { link.attached = true },
      },
      createElement(tag) {
        assert.equal(tag, 'a')
        const link = {
          attached: false,
          download: '',
          href: '',
          click() {
            assert.equal(this.attached, true)
            downloads.push({ filename: this.download, href: this.href })
          },
          remove() { this.attached = false },
        }
        links.push(link)
        return link
      },
    },
  })
  container = {}
  const vnode = Vue.createVNode(component)
  renderer.render(vnode, container)
  state = vnode.component.setupState
})

afterEach(() => {
  renderer.render(null, container)
  if (originalDocument) Object.defineProperty(globalThis, 'document', originalDocument)
  else delete globalThis.document
})

function deferred() {
  let resolve
  const promise = new Promise((done) => { resolve = done })
  return { promise, resolve }
}

async function readyDetail(id = 17) {
  await state.showDetail(bill(id))
  state.detailContent = Vue.markRaw({ scrollWidth: 620, scrollHeight: 3200 })
  assert.equal(state.canExportDetail, true)
}

async function flush() {
  await new Promise((resolve) => setImmediate(resolve))
  await Vue.nextTick()
}

test('详情未加载、正在加载或加载失败时禁止导出', async () => {
  state.detail = bill()
  state.detailOpen = true
  state.detailContent = Vue.markRaw({ scrollWidth: 620, scrollHeight: 3200 })
  assert.equal(state.canExportDetail, false)
  await state.exportDetailImage()

  const response = deferred()
  getBillDetail = () => response.promise
  const loading = state.showDetail(bill())
  assert.equal(state.detailLoading, true)
  assert.equal(state.canExportDetail, false)
  await state.exportDetailImage()
  response.resolve(bill())
  await loading

  getBillDetail = async () => { throw new Error('无法加载账单') }
  await state.showDetail(bill())
  assert.equal(state.detailLoaded, true)
  assert.equal(state.detailError, '无法加载账单')
  assert.equal(state.canExportDetail, false)
  await state.exportDetailImage()
  assert.equal(pngCalls.length, 0)
  assert.deepEqual(downloads, [])
})

test('正常导出 PNG，以完整正文尺寸生成并下载账单文件', async () => {
  await readyDetail()
  state.rooms = [{ id: 3, room_no: '302' }]
  assert.equal(state.detailRoomName, '302')
  const content = state.detailContent
  await state.exportDetailImage()
  assert.equal(pngCalls.length, 1)
  assert.equal(pngCalls[0].node, content)
  assert.deepEqual(pngCalls[0].options, {
    backgroundColor: '#fff',
    pixelRatio: 2,
    width: 620,
    height: 3200,
  })
  assert.deepEqual(downloads, [{ filename: '账单-2026-10-302-17.png', href: pngUrl }])
  assert.equal(links[0].attached, false)
  assert.equal(state.exporting, false)
  assert.deepEqual(messages, [])
})

test('导出文件名替换房间号中的非法字符并去除首尾空格', async () => {
  await readyDetail()
  const roomNo = ' 3<>:"/\\|?*\u0000\u001f\u007f02 '
  state.rooms = [{ id: 3, room_no: roomNo }]
  assert.equal(state.detailRoomName, roomNo)
  await state.exportDetailImage()
  assert.deepEqual(downloads, [{
    filename: '账单-2026-10-3____________02-17.png',
    href: pngUrl,
  }])
})

test('找不到房间名称时，标题和导出文件名使用房间 ID', async () => {
  await readyDetail()
  state.rooms = [{ id: 4, room_no: '402' }]
  assert.equal(state.detailRoomName, '房间 #3')
  await state.exportDetailImage()
  assert.deepEqual(downloads, [{ filename: '账单-2026-10-房间 #3-17.png', href: pngUrl }])
})

test('图片生成期间重复点击只触发一次生成和下载', async () => {
  await readyDetail()
  const image = deferred()
  renderPng = () => image.promise
  const firstExport = state.exportDetailImage()
  assert.equal(state.exporting, true)
  await state.exportDetailImage()
  await flush()
  assert.equal(pngCalls.length, 1)
  image.resolve(pngUrl)
  await firstExport
  assert.equal(downloads.length, 1)
  assert.equal(state.exporting, false)
})

test('字体准备期间切换账单，取消旧详情的图片生成', async () => {
  await readyDetail()
  const fonts = deferred()
  fontsReady = fonts.promise
  const exporting = state.exportDetailImage()
  await flush()
  await state.showDetail(bill(18))
  fonts.resolve()
  await exporting
  assert.equal(pngCalls.length, 0)
  assert.deepEqual(downloads, [])
  assert.equal(state.exporting, false)
})

test('图片生成期间关闭详情，取消下载', async () => {
  await readyDetail()
  const image = deferred()
  renderPng = () => image.promise
  const exporting = state.exportDetailImage()
  await flush()
  state.detailOpen = false
  image.resolve(pngUrl)
  await exporting
  assert.deepEqual(downloads, [])
  assert.deepEqual(messages, [])
  assert.equal(state.exporting, false)
})

test('图片生成期间重新加载同一账单，取消旧详情的下载', async () => {
  await readyDetail()
  const image = deferred()
  renderPng = () => image.promise
  const exporting = state.exportDetailImage()
  await flush()
  await state.showDetail(bill())
  image.resolve(pngUrl)
  await exporting
  assert.deepEqual(downloads, [])
  assert.equal(state.exporting, false)
})

test('图片生成期间切换物业，取消旧详情的下载', async () => {
  await readyDetail()
  const image = deferred()
  renderPng = () => image.promise
  const exporting = state.exportDetailImage()
  await flush()
  selectedUnitId.value = 2
  await Vue.nextTick()
  image.resolve(pngUrl)
  await exporting
  assert.equal(state.detailOpen, false)
  assert.deepEqual(downloads, [])
  assert.equal(state.exporting, false)
})

test('生成失败后显示错误并解除锁定，允许重试下载', async () => {
  await readyDetail()
  renderPng = async () => { throw new Error('无法生成图片') }
  await state.exportDetailImage()
  assert.equal(state.exporting, false)
  assert.deepEqual(downloads, [])
  assert.deepEqual(messages, ['图片导出失败，请重试'])

  renderPng = async () => pngUrl
  await state.exportDetailImage()
  assert.equal(pngCalls.length, 2)
  assert.equal(downloads.length, 1)
  assert.equal(state.exporting, false)
})

test('图片库返回非 PNG 数据时禁止下载并允许重试', async () => {
  await readyDetail()
  renderPng = async () => 'data:,'
  await state.exportDetailImage()
  assert.deepEqual(downloads, [])
  assert.deepEqual(messages, ['图片导出失败，请重试'])
  assert.equal(state.exporting, false)
})
