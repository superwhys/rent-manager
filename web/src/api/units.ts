import { request } from '@/api/request'
import type { Unit, UnitInput, UnitScope } from '@/types/units'

export function listUnits(scope: UnitScope = 'all') {
  return request<Unit[]>('/units', { query: { scope } })
}

export function getUnit(id: number) {
  return request<Unit>(`/units/${id}`)
}

export function createUnit(body: UnitInput) {
  return request<Unit>('/units', { method: 'POST', body })
}

export function updateUnit(id: number, body: UnitInput) {
  return request<Unit>(`/units/${id}`, { method: 'PUT', body })
}

export function deleteUnit(id: number) {
  return request<null>(`/units/${id}`, { method: 'DELETE' })
}
