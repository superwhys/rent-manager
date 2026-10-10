import { request } from '@/api/request'
import type { BillQuery, GeneratedBills, GenerateBillsInput, RentBill } from '@/types/bills'

export function listUnitBills(unitId: number, query: BillQuery = {}) {
  return request<RentBill[]>(`/units/${unitId}/rent-bills`, { query: { ...query } })
}

export function listRoomBills(roomId: number, query: BillQuery = {}) {
  return request<RentBill[]>(`/rooms/${roomId}/rent-bills`, { query: { ...query } })
}

export function getBill(id: number) {
  return request<RentBill>(`/rent-bills/${id}`)
}

export function generateUnitBills(unitId: number, body: GenerateBillsInput) {
  return request<GeneratedBills>(`/units/${unitId}/rent-bills/generate`, { method: 'POST', body })
}

export function generateRoomBill(roomId: number, body: GenerateBillsInput) {
  return request<GeneratedBills>(`/rooms/${roomId}/rent-bills/generate`, { method: 'POST', body })
}

export function markBillPaid(id: number, paidAt?: string) {
  return request<RentBill>(`/rent-bills/${id}/paid`, {
    method: 'POST',
    body: paidAt ? { paid_at: paidAt } : {},
  })
}

export function cancelBill(id: number, remark?: string) {
  return request<RentBill>(`/rent-bills/${id}/cancel`, { method: 'POST', body: { remark } })
}
