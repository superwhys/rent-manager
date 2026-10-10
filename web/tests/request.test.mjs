import { afterEach, test } from 'node:test'
import assert from 'node:assert/strict'
import { ApiError, configureSession, request } from '../src/api/request.ts'

const originalFetch = globalThis.fetch
afterEach(() => { globalThis.fetch = originalFetch; configureSession('') })

test('发送会话和真实查询参数，保留 false、0 和 DELETE 请求体', async () => {
  configureSession('test-session')
  globalThis.fetch = async (url, options) => {
    assert.equal(url, '/api/roles/2/permission-groups?is_active=false&offset=0')
    assert.equal(options.headers.Authorization, 'Bearer test-session')
    assert.equal(options.method, 'DELETE')
    assert.deepEqual(JSON.parse(options.body), { permission_group_ids: [1] })
    return Response.json({ code: 0, data: null })
  }
  assert.equal(await request('/roles/2/permission-groups', {
    method: 'DELETE', body: { permission_group_ids: [1] }, query: { is_active: false, offset: 0, empty: undefined },
  }), null)
})

test('HTTP 200 业务错误会抛出，权限不足不退出会话', async () => {
  let expired = 0
  configureSession('test-session', () => expired++)
  globalThis.fetch = async () => Response.json({ code: 207022, data: null, message: '权限不足' })
  await assert.rejects(request('/roles'), error => error instanceof ApiError && error.message === '权限不足')
  assert.equal(expired, 0)
})

test('登录失效会通知上层；显式验证候选会话不清除已有会话', async () => {
  let expired = 0
  configureSession('test-session', () => expired++)
  globalThis.fetch = async () => Response.json({ code: 400002, data: null, message: '未登录' }, { status: 401 })
  await assert.rejects(request('/auth/session', { token: 'candidate' }))
  assert.equal(expired, 0)
  await assert.rejects(request('/auth/session'))
  assert.equal(expired, 1)
})

test('旧会话的迟到失败不会清除新登录会话', async () => {
  let expired = 0
  let finish
  configureSession('old', () => expired++)
  globalThis.fetch = () => new Promise(resolve => { finish = resolve })
  const pending = request('/units')
  configureSession('new')
  finish(Response.json({ code: 400002, data: null }, { status: 401 }))
  await assert.rejects(pending)
  assert.equal(expired, 0)
})

test('公开邮箱认证请求不发送旧会话，登录失败不清除已有会话', async () => {
  let expired = 0
  configureSession('existing-session', () => expired++)
  globalThis.fetch = async (_url, options) => {
    assert.equal(options.headers.Authorization, undefined)
    return Response.json({ code: 400002, data: null, message: '邮箱或密码错误' }, { status: 401 })
  }
  await assert.rejects(request('/auth/password-login', {
    method: 'POST', token: '', body: { email: 'owner@example.com', password: ' wrong ' },
  }), /邮箱或密码错误/)
  assert.equal(expired, 0)
})

test('网络和非 JSON 失败提供可操作提示', async () => {
  globalThis.fetch = async () => { throw new TypeError('Failed to fetch') }
  await assert.rejects(request('/home'), /无法连接服务/)
  globalThis.fetch = async () => new Response('<html>bad gateway</html>', { status: 502 })
  await assert.rejects(request('/home'), /请求未完成/)
})
