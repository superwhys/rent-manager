import type { MeterItem } from '@/types/meterItems'

export type UnitMeterItem = Omit<MeterItem, 'account_id'>

export interface Unit {
  id: number
  account_id?: string
  name: string
  address?: string
  city?: string
  province?: string
  access_type: 'owned' | 'granted' | ''
  is_owner: boolean
  default_meter_items?: UnitMeterItem[]
  json_extend?: { meter_reading_day: number }
  room_count?: number
  rented_count?: number
  monthly_rent_amount?: string
  occupancy_status?: 'vacancy' | 'full' | ''
}

export interface UnitInput {
  name: string
  address: string
  city: string
  province: string
  default_meter_item_ids: number[]
  json_extend: { meter_reading_day: number }
}

export type UnitScope = 'all' | 'owned' | 'granted'
