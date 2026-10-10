import { request } from '@/api/request'
import type { Room, RoomInput, RoomQuery, RoomUpdateInput } from '@/types/rooms'

export function listRooms(unitId: number, query: RoomQuery = {}) {
  return request<Room[]>(`/units/${unitId}/rooms`, { query: { ...query } })
}

export function getRoom(id: number) {
  return request<Room>(`/rooms/${id}`)
}

export function createRoom(unitId: number, body: RoomInput) {
  return request<Room>(`/units/${unitId}/rooms`, { method: 'POST', body })
}

export function updateRoom(id: number, body: RoomUpdateInput) {
  return request<Room>(`/rooms/${id}`, { method: 'PUT', body })
}

export function deleteRoom(id: number) {
  return request<null>(`/rooms/${id}`, { method: 'DELETE' })
}
