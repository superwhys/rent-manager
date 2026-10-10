import type { MeterItem } from '@/types/meterItems'

export interface MeterReadingLine {
  id?: number
  room_id: number
  period_month: string
  meter_item: MeterItem
  previous_reading?: number
  previous_period_month?: string
  current_reading?: number
  usage_amount?: number
  reading_time?: string
  reader_id?: string
  source?: string
  remark?: string
  fee: number
  is_recorded: boolean
}

export interface RoomMonthReadings {
  room_id: number
  room_no: string
  unit_id: number
  period_month: string
  readings: MeterReadingLine[]
  total_fee: number
  missing_count: number
}

export interface MeterItemSummary {
  meter_item_id: number
  code: string
  unit: string
  calc_type: string
  unit_price: number
  usage_amount: number
  fee: number
  recorded: number
  missing: number
}

export interface UnitMonthReadings {
  unit_id: number
  period_month: string
  rooms: RoomMonthReadings[]
  summaries: MeterItemSummary[]
  total_fee: number
  missing_count: number
}

export interface MeterReadingInput {
  meter_item_id: number
  previous_reading?: number
  current_reading: number
  reading_time?: string
  remark?: string
}

export interface SaveReadingsInput {
  period_month: string
  readings: MeterReadingInput[]
}
