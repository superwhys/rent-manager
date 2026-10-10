import { request } from '@/api/request'
import type { RoomMonthReadings, SaveReadingsInput, UnitMonthReadings } from '@/types/readings'

export function getUnitReadings(unitId: number, periodMonth: string) {
  return request<UnitMonthReadings>(`/units/${unitId}/meter-readings`, {
    query: { period_month: periodMonth },
  })
}

export function getRoomReadings(roomId: number, periodMonth: string) {
  return request<RoomMonthReadings>(`/rooms/${roomId}/meter-readings`, {
    query: { period_month: periodMonth },
  })
}

export function saveRoomReadings(roomId: number, body: SaveReadingsInput) {
  return request<RoomMonthReadings>(`/rooms/${roomId}/meter-readings`, { method: 'PUT', body })
}
