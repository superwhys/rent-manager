import assert from 'node:assert/strict'
import { register } from 'node:module'
import { afterEach, beforeEach, test } from 'node:test'

register(`data:text/javascript,${encodeURIComponent(`
  export function resolve(specifier, context, nextResolve) {
    if (specifier.startsWith('@/'))
      specifier = new URL(specifier.slice(2) + '.ts', ${JSON.stringify(new URL('../src/', import.meta.url).href)}).href
    if (specifier === './request') specifier += '.ts'
    return nextResolve(specifier, context)
  }
`)}`, import.meta.url)

const { request } = await import('../src/api/request.ts')
const originalFetch = globalThis.fetch
const originalGlobals = new Map(
  ['localStorage', 'sessionStorage', 'window'].map((key) => [
    key, Object.getOwnPropertyDescriptor(globalThis, key),
  ]),
)
const storageKey = 'rent-manager.session'
const session = {
  account: { ID: 'a'.repeat(64), Username: '房东', Email: 'owner@example.com' },
  expires_at: '2099-01-01T00:00:00Z',
  is_super_admin: false,
}
let saved
let reloads
let instance = 0

function storage(values = new Map()) {
  return {
    getItem: (key) => values.get(key) ?? null,
    setItem: (key, value) => values.set(key, String(value)),
    removeItem: (key) => values.delete(key),
    clear: () => values.clear(),
  }
}

async function newTab() {
  Object.defineProperty(globalThis, 'sessionStorage', {
    configurable: true, value: storage(),
  })
  const { useSession } = await import(`../src/stores/session.ts?tab=${++instance}`)
  return useSession()
}

beforeEach(() => {
  saved = new Map()
  reloads = 0
  Object.defineProperty(globalThis, 'localStorage', {
    configurable: true, value: storage(saved),
  })
  Object.defineProperty(globalThis, 'window', {
    configurable: true, value: { location: { reload() { reloads++ } } },
  })
})

afterEach(() => {
  globalThis.fetch = originalFetch
  for (const [key, descriptor] of originalGlobals) {
    if (descriptor) Object.defineProperty(globalThis, key, descriptor)
    else delete globalThis[key]
  }
})

test('密码登录、验证码登录和注册都持久化到共享存储', async () => {
  for (const [method, path] of [
    ['loginWithPassword', '/auth/password-login'],
    ['loginWithCode', '/auth/email-login'],
    ['register', '/auth/register'],
  ]) {
    saved.clear()
    const auth = await newTab()
    globalThis.fetch = async (url) => {
      assert.equal(url, `/api${path}`)
      return Response.json({ code: 0, data: { ...session, access_token: 'shared-token' } })
    }
    await auth[method]({ email: session.account.Email, password: 'test-password', code: '123456' })
    assert.equal(saved.get(storageKey), 'shared-token')
    assert.equal(sessionStorage.getItem(storageKey), null)
    assert.equal(auth.isAuthenticated.value, true)
  }
})

test('登录后独立新标签页读取共享凭证，验证服务端会话后恢复登录', async () => {
  const first = await newTab()
  globalThis.fetch = async () => Response.json({
    code: 0, data: { ...session, access_token: 'shared-token' },
  })
  await first.loginWithPassword({ email: session.account.Email, password: 'test-password' })

  let validations = 0
  globalThis.fetch = async (url, options) => {
    validations++
    assert.equal(url, '/api/auth/session')
    assert.equal(options.headers.Authorization, 'Bearer shared-token')
    return Response.json({ code: 0, data: session })
  }
  const second = await newTab()
  assert.equal(second.isAuthenticated.value, false)
  await second.initialize()
  assert.equal(validations, 1)
  assert.deepEqual(second.session.value, session)
  assert.equal(second.isAuthenticated.value, true)
})

test('失效凭证清除共享存储，网络失败保留凭证以便重试', async () => {
  saved.set(storageKey, 'expired-token')
  const auth = await newTab()
  globalThis.fetch = async () => { throw new TypeError('Failed to fetch') }
  await assert.rejects(auth.initialize(), /无法连接服务/)
  assert.equal(saved.get(storageKey), 'expired-token')

  globalThis.fetch = async () => Response.json({ code: 400002, data: null }, { status: 401 })
  await auth.initialize()
  assert.equal(saved.has(storageKey), false)
  assert.equal(auth.isAuthenticated.value, false)
})

test('退出登录撤销服务端会话并清除共享凭证', async () => {
  saved.set(storageKey, 'shared-token')
  const auth = await newTab()
  globalThis.fetch = async () => Response.json({ code: 0, data: session })
  await auth.initialize()
  globalThis.fetch = async (url, options) => {
    assert.equal(url, '/api/auth/session')
    assert.equal(options.method, 'DELETE')
    assert.equal(options.headers.Authorization, 'Bearer shared-token')
    return Response.json({ code: 0, data: null })
  }
  await auth.logout()
  assert.equal(saved.has(storageKey), false)
  assert.equal(auth.isAuthenticated.value, false)
})

test('其他标签页登录、切换账号、退出或清空存储时刷新页面恢复正确会话', async () => {
  for (const [before, after, key] of [
    ['', 'new-token', storageKey],
    ['old-token', 'new-token', storageKey],
    ['old-token', '', storageKey],
    ['old-token', '', null],
  ]) {
    saved.clear()
    if (before) saved.set(storageKey, before)
    const auth = await newTab()
    if (after) saved.set(storageKey, after)
    else saved.delete(storageKey)
    const previousReloads = reloads
    auth.syncSession({ storageArea: localStorage, key, newValue: after || null })
    assert.equal(reloads, previousReloads + 1)
  }
})

test('无关存储事件、相同凭证和已过时事件不会刷新页面', async () => {
  saved.set(storageKey, 'current-token')
  const auth = await newTab()
  auth.syncSession({ storageArea: localStorage, key: storageKey, newValue: 'current-token' })
  auth.syncSession({ storageArea: localStorage, key: storageKey, newValue: 'old-token' })
  saved.set(storageKey, 'new-token')
  auth.syncSession({ storageArea: localStorage, key: 'rent-manager.selected-unit.a', newValue: '2' })
  auth.syncSession({ storageArea: sessionStorage, key: storageKey, newValue: 'new-token' })
  assert.equal(reloads, 0)
})

test('旧标签页迟到的认证失败不会删除其他标签页的新凭证', async () => {
  saved.set(storageKey, 'old-token')
  await newTab()
  let finish
  globalThis.fetch = () => new Promise((resolve) => { finish = resolve })
  const pending = request('/units')
  saved.set(storageKey, 'new-token')
  finish(Response.json({ code: 400002, data: null }, { status: 401 }))
  await assert.rejects(pending)
  assert.equal(saved.get(storageKey), 'new-token')
})

test('挂载时同步初始化期间在其他标签页变化的凭证', async () => {
  const auth = await newTab()
  saved.set(storageKey, 'new-token')
  auth.syncSession()
  assert.equal(reloads, 1)
})

test('旧标签页退出完成时保留其他标签页刚登录的新凭证', async () => {
  saved.set(storageKey, 'old-token')
  const auth = await newTab()
  let finish
  globalThis.fetch = () => new Promise((resolve) => { finish = resolve })
  const pending = auth.logout()
  saved.set(storageKey, 'new-token')
  finish(Response.json({ code: 0, data: null }))
  await pending
  assert.equal(saved.get(storageKey), 'new-token')
})
