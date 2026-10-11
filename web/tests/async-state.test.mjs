import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { test } from 'node:test'
import ts from 'typescript'
import * as Vue from 'vue'
import { compileScript, parse } from 'vue/compiler-sfc'

const filename = new URL('../src/components/common/AsyncState.vue', import.meta.url)
const { descriptor } = parse(await readFile(filename, 'utf8'))
const script = compileScript(descriptor, { id: 'async-state-test' })
const { outputText } = ts.transpileModule(script.content, {
  compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 },
})
const componentModule = { exports: {} }
new Function('require', 'module', 'exports', outputText)(
  (specifier) => {
    if (specifier === 'vue') return Vue
    if (specifier.endsWith('.vue')) return {}
    throw new Error(`未处理的组件导入：${specifier}`)
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

function checkState(props, check) {
  const container = {}
  const vnode = Vue.createVNode(component, props)
  renderer.render(vnode, container)
  try {
    check(vnode.component.setupState)
  } finally {
    renderer.render(null, container)
  }
}

test('首次加载隐藏内容，同一上下文刷新保留已有内容', () => {
  checkState({ loading: true }, (state) => assert.equal(state.showContent, false))
  checkState({ loading: true, hasContent: true }, (state) => assert.equal(state.showContent, true))
  checkState({ empty: true }, (state) => assert.equal(state.showContent, false))
  checkState({}, (state) => assert.equal(state.showContent, true))
})

test('普通刷新失败保留已有内容，没有已有内容时显示失败状态', () => {
  checkState({ error: '无法连接服务', hasContent: true }, (state) => assert.equal(state.showContent, true))
  checkState({ error: '无法连接服务' }, (state) => assert.equal(state.showContent, false))
})

test('权限不足时始终隐藏旧内容，包括保存后刷新添加的错误前缀', () => {
  for (const error of [
    '无权访问',
    '你没有执行此操作的权限',
    '操作已完成，但账单刷新失败：无权访问',
    '抄表已保存，但月度汇总刷新失败：你没有执行此操作的权限',
  ]) {
    checkState({ error, hasContent: true }, (state) => {
      assert.equal(state.isForbidden, true, error)
      assert.equal(state.showContent, false, error)
    })
  }
})
