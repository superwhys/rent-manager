import assert from 'node:assert/strict'
import { setTimeout as sleep } from 'node:timers/promises'

// 仅向本地测试服务追加演示数据，使用 README 中公开的演示账号。
const baseUrl = 'http://127.0.0.1:8080/api'
const unitName = '演示·账单明细'
const roomTypeName = '快照单间'
const definitions = [
  { code: '快照演示水费', unit: '吨', unit_price: 450.5, calc_type: 'DIFFERENCE' },
  { code: '快照演示电费', unit: '度', unit_price: 85.75, calc_type: 'DIFFERENCE' },
  { code: '快照演示物业费', unit: '月', unit_price: 5000.25, calc_type: 'DIRECT' },
]
const roomNumbers = ['101', '102', '201', '202']
const statuses = ['PAID', 'UNPAID', 'CANCELLED', 'PAID']
let token = ''
let lastRequestAt = 0
let unitId = null
let roomTypeId = null
const meterItemIds = []
const roomIds = []

async function request(path, method = 'GET', body) {
  for (let attempt = 0; ; attempt++) {
    await sleep(Math.max(0, 1100 - (Date.now() - lastRequestAt)))
    const headers = { Accept: 'application/json' }
    if (token) headers.Authorization = `Bearer ${token}`
    if (body !== undefined) headers['Content-Type'] = 'application/json'
    lastRequestAt = Date.now()
    const response = await fetch(`${baseUrl}${path}`, {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
      signal: AbortSignal.timeout(30000),
      redirect: 'error',
    })
    // 429 由限流中间件在业务执行前拒绝；其他失败不重试，避免重复写入。
    if (response.status === 429 && attempt < 3) {
      await response.body?.cancel()
      console.log('触发限流，等待 11 秒后重试被拒绝的请求。')
      await sleep(11000)
      continue
    }
    const payload = await response.json().catch(() => null)
    if (!response.ok || payload?.code !== 0 || !Object.hasOwn(payload, 'data')) {
      throw new Error(`请求失败：${method} ${path}，HTTP ${response.status}，业务码 ${payload?.code ?? '未知'}`)
    }
    return payload.data
  }
}

try {
  assert.equal(process.argv.length, 2, '脚本不接受参数，只能连接本地测试服务。')
  assert.ok(Number(process.versions.node.split('.')[0]) >= 24, '请使用 Node.js 24 或更新版本。')
  const login = await request('/auth/password-login', 'POST', {
    email: 'owner@rent-demo.test',
    password: 'RentDemo2026!',
  })
  token = login.access_token
  assert.ok(typeof token === 'string' && token, '登录未返回可用会话。')

  const units = await request('/units?scope=owned')
  assert.ok(!units.some(unit => unit.name === unitName), `${unitName} 已存在；脚本只追加一次，不覆盖数据。`)
  const roomTypes = await request(`/room-types?name=${encodeURIComponent(roomTypeName)}`)
  assert.ok(!roomTypes.some(type => type.name === roomTypeName), '专属演示房型已存在，停止追加。')
  const meterItems = await request('/meter-items?code=快照演示')
  assert.ok(!meterItems.some(item => definitions.some(definition => definition.code === item.code)), '专属演示抄表项已存在，停止追加。')

  // 写入前确认运行中的服务已提供冻结明细契约，旧账单允许返回 null。
  let supportsSnapshots = false
  for (const unit of units) {
    const bills = await request(`/units/${unit.id}/rent-bills?limit=1`)
    if (bills.length === 0) continue
    const bill = await request(`/rent-bills/${bills[0].id}`)
    supportsSnapshots = Object.hasOwn(bill, 'meter_readings')
      && (bill.meter_readings === null || Array.isArray(bill.meter_readings))
    break
  }
  assert.ok(supportsSnapshots, '未找到支持 meter_readings 的既有账单详情，请先导入基础演示数据并启动新版服务。')

  const today = new Intl.DateTimeFormat('en-CA', {
    timeZone: 'Asia/Shanghai', year: 'numeric', month: '2-digit',
  }).formatToParts(new Date())
  const year = Number(today.find(part => part.type === 'year').value)
  const month = Number(today.find(part => part.type === 'month').value)
  const months = [-2, -1, 0].map(offset => {
    const date = new Date(Date.UTC(year, month - 1 + offset, 1))
    return date.toISOString().slice(0, 7)
  })

  console.log(`准备追加 ${unitName}，月份：${months.join('、')}。`)
  const roomType = await request('/room-types', 'POST', { name: roomTypeName })
  roomTypeId = roomType.id
  for (const definition of definitions) {
    const item = await request('/meter-items', 'POST', {
      ...definition, remark: '快照演示：单价单位为分，精细单价用于展示逐项四舍五入。',
    })
    meterItemIds.push(item.id)
  }
  const unit = await request('/units', 'POST', {
    name: unitName,
    address: '南山区快照演示路 10 号',
    city: '深圳市',
    province: '广东省',
    default_meter_item_ids: meterItemIds,
    json_extend: { meter_reading_day: 1 },
  })
  unitId = unit.id
  console.log(`已创建物业 ID ${unitId}，开始录入房间及账单。`)
  for (const [index, roomNo] of roomNumbers.entries()) {
    const room = await request(`/units/${unitId}/rooms`, 'POST', {
      room_no: roomNo,
      floor: Number(roomNo[0]),
      price: 200000 + index * 15000,
      size: 28 + index * 3,
      room_type_id: roomTypeId,
      meter_item_ids: meterItemIds,
    })
    roomIds.push(room.id)
    await request(`/rooms/${room.id}`, 'PUT', { is_active: false })
  }

  const expected = new Map()
  const previous = roomNumbers.map((_, index) => index === 3 ? [0, 0] : [100 + index * 10, 500 + index * 100])
  const results = []
  let frozenBill
  for (const [monthIndex, periodMonth] of months.entries()) {
    for (const [roomIndex, roomId] of roomIds.entries()) {
      const usage = roomIndex === 3 ? [0, 0, 0] : [8.25 + roomIndex + monthIndex * 1.25, 63.5 + roomIndex * 4 + monthIndex * 5.25, 1]
      const readings = meterItemIds.map((meterItemId, itemIndex) => ({
        meter_item_id: meterItemId,
        ...(itemIndex < 2 ? { previous_reading: previous[roomIndex][itemIndex] } : {}),
        current_reading: itemIndex < 2 ? previous[roomIndex][itemIndex] + usage[itemIndex] : usage[itemIndex],
        reading_time: `${periodMonth}-01T09:00:00+08:00`,
        remark: `快照演示：${roomNumbers[roomIndex]} ${definitions[itemIndex].code}${roomIndex === 3 ? '零用量' : '月度抄表'}。`,
      }))
      await request(`/rooms/${roomId}/meter-readings`, 'PUT', { period_month: periodMonth, readings })
      expected.set(`${roomId}/${periodMonth}`, { readings, usage })
      previous[roomIndex] = readings.slice(0, 2).map(reading => reading.current_reading)
    }

    const generated = await request(`/units/${unitId}/rent-bills/generate`, 'POST', {
      period_month: periodMonth, remark: '快照演示：水电差值计费、物业直接计费，逐项金额为开单时冻结值。',
    })
    assert.equal(generated.saved_bills.length, 4, '每月应生成 4 张新账单。')
    assert.equal(generated.skipped_bills.length, 0, '新演示物业不应存在跳过的账单。')
    for (const saved of generated.saved_bills) {
      const roomIndex = roomIds.indexOf(saved.room_id)
      assert.ok(roomIndex >= 0, '生成结果包含预期之外的房间。')
      const status = statuses[roomIndex]
      if (status === 'PAID') {
        await request(`/rent-bills/${saved.id}/paid`, 'POST', { paid_at: `${periodMonth}-03T12:00:00+08:00` })
      } else if (status === 'CANCELLED') {
        await request(`/rent-bills/${saved.id}/cancel`, 'POST', { remark: '快照演示：已取消账单仍保留原抄表明细。' })
      }
      const bill = await request(`/rent-bills/${saved.id}`)
      assert.equal(bill.status, status)
      assert.equal(bill.period_month, periodMonth)
      assert.equal(bill.meter_readings?.length, 3, '账单应有 3 条冻结明细。')
      const input = expected.get(`${saved.room_id}/${periodMonth}`)
      for (const [itemIndex, itemId] of meterItemIds.entries()) {
        const line = bill.meter_readings.find(reading => reading.meter_item.id === itemId)
        assert.ok(line, '冻结明细缺少抄表项。')
        assert.equal(line.room_id, saved.room_id)
        assert.equal(line.period_month, periodMonth)
        assert.equal(line.is_recorded, true)
        assert.equal(line.source, 'MANUAL')
        assert.equal(line.meter_item.code, definitions[itemIndex].code)
        assert.equal(line.meter_item.unit_price, definitions[itemIndex].unit_price)
        assert.equal(line.meter_item.calc_type, definitions[itemIndex].calc_type)
        assert.equal(line.current_reading, input.readings[itemIndex].current_reading)
        assert.equal(line.previous_reading ?? null, input.readings[itemIndex].previous_reading ?? null)
        assert.equal(line.usage_amount, input.usage[itemIndex])
        assert.equal(line.fee, Math.round(input.usage[itemIndex] * definitions[itemIndex].unit_price))
        assert.equal(line.remark, input.readings[itemIndex].remark)
        assert.equal(Date.parse(line.reading_time), Date.parse(input.readings[itemIndex].reading_time))
      }
      assert.equal(bill.meter_readings.reduce((sum, line) => sum + line.fee, 0), bill.meter_fee)
      assert.equal(bill.total_amount, bill.base_rent + bill.meter_fee)
      if (roomIndex === 3 && monthIndex === 2) frozenBill = bill
      results.push({ 月份: periodMonth, 房间: roomNumbers[roomIndex], 账单ID: bill.id, 状态: bill.status, 抄表费_元: (bill.meter_fee / 100).toFixed(2), 总额_元: (bill.total_amount / 100).toFixed(2) })
    }
    console.log(`${periodMonth}：4 张账单及 12 条冻结明细验证通过。`)
  }

  // 仅校正新建演示房间，直观展示已付账单保存的是开单时的读数。
  const latestMonth = months[2]
  const correctedReadings = expected.get(`${roomIds[3]}/${latestMonth}`).readings.map(reading => ({ ...reading }))
  correctedReadings[0].current_reading = 1
  correctedReadings[0].remark = '快照演示：开单后校正水表读数为 1；已付账单仍保留原零读数。'
  const corrected = await request(`/rooms/${roomIds[3]}/meter-readings`, 'PUT', { period_month: latestMonth, readings: correctedReadings })
  const correctedWater = corrected.readings.find(reading => reading.meter_item.id === meterItemIds[0])
  assert.equal(correctedWater.current_reading, 1)
  assert.equal(correctedWater.usage_amount, 1)
  const afterCorrection = await request(`/rent-bills/${frozenBill.id}`)
  assert.deepEqual(afterCorrection.meter_readings, frozenBill.meter_readings, '校正抄表后原账单快照发生变化。')
  assert.equal(afterCorrection.meter_fee, 0)
  console.log(`完成：物业 ID ${unitId}，4 间已租房，12 张账单，36 条冻结明细。`)
  console.log('状态分布：已付 6 张、未付 3 张、已取消 3 张。')
  console.table(results)
  console.log(`冻结对比：202 房 ${latestMonth} 当前水表已校正为 1，账单 ${frozenBill.id} 仍保存读数 0、用量 0、费用 0。`)
} catch (error) {
  console.error(error instanceof Error ? error.message : '演示数据追加失败。')
  console.error(`已创建：物业 ID ${unitId ?? '无'}；房型 ID ${roomTypeId ?? '无'}；抄表项 IDs ${meterItemIds.join(', ') || '无'}；房间 IDs ${roomIds.join(', ') || '无'}。`)
  console.error('已创建的数据保留，未自动删除；请定位失败后处理，重复执行会因同名数据停止。')
  process.exitCode = 1
} finally {
  if (token) {
    try {
      await request('/auth/session', 'DELETE')
    } catch {
      console.error('临时演示会话撤销失败，请检查本地服务。')
      process.exitCode = 1
    }
    token = ''
  }
}
