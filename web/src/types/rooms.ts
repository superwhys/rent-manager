import type { MeterItem } from '@/types/meterItems'

export interface Room {
  id: number
  room_no: string
  floor: number
  price: number
  size: number
  unit_id: number
  room_type_id: number
  is_active: boolean
  meter_items: MeterItem[]
}

export interface RoomInput {
  room_no: string
  floor: number
  price: number
  size: number
  room_type_id: number
  meter_item_ids: number[]
}

export interface RoomUpdateInput extends RoomInput {
  unit_id: number
  is_active: boolean
}

export interface RoomQuery {
  limit?: number
  offset?: number
  room_type_id?: number
  is_active?: boolean
}
