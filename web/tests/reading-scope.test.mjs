import assert from 'node:assert/strict'
import { register } from 'node:module'
import { afterEach, test } from 'node:test'

register(`data:text/javascript,${encodeURIComponent(`
  export function resolve(specifier, context, nextResolve) {
    if (specifier.startsWith('@/'))
      specifier = new URL(specifier.slice(2) + '.ts', ${JSON.stringify(new URL('../src/', import.meta.url).href)}).href
    return nextResolve(specifier, context)
  }
`)}`, import.meta.url)

const { getUnitReadings } = await import('../src/api/readings.ts')
const originalFetch = globalThis.fetch
afterEach(() => { globalThis.fetch = originalFetch })

test('月度抄表默认只请求已出租房间，可显式包含所有空房', async () => {
  const urls = []
  const response = { rooms: [], summaries: [], missing_count: 0, total_fee: 0 }
  globalThis.fetch = async (url) => {
    urls.push(url)
    return Response.json({ code: 0, data: response })
  }
  assert.deepEqual(await getUnitReadings(3, '2026-10'), response)
  assert.deepEqual(await getUnitReadings(3, '2026-10', true), response)
  assert.deepEqual(urls, [
    '/api/units/3/meter-readings?period_month=2026-10&include_vacant_records=false',
    '/api/units/3/meter-readings?period_month=2026-10&include_vacant_records=true',
  ])
})
