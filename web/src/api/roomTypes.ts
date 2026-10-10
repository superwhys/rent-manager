import { request } from '@/api/request'
import type { RoomType, RoomTypeQuery } from '@/types/roomTypes'

export function listRoomTypes(query: RoomTypeQuery = {}) {
  return request<RoomType[]>('/room-types', { query: { ...query } })
}

export function createRoomType(name: string) {
  return request<RoomType>('/room-types', { method: 'POST', body: { name } })
}

export function updateRoomType(id: number, name: string) {
  return request<RoomType>(`/room-types/${id}`, {
    method: 'PUT',
    body: { name },
  })
}

export function deleteRoomType(id: number) {
  return request<null>(`/room-types/${id}`, { method: 'DELETE' })
}
