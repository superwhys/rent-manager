import { request } from '@/api/request'
import type {
  MeterItem,
  MeterItemInput,
  MeterItemQuery,
} from '@/types/meterItems'

export const listMeterItems = (query: MeterItemQuery = {}) =>
  request<MeterItem[]>('/meter-items', { query: { ...query } })
export const getMeterItem = (id: number) =>
  request<MeterItem>(`/meter-items/${id}`)
export const createMeterItem = (body: MeterItemInput) =>
  request<MeterItem>('/meter-items', { method: 'POST', body })
export const updateMeterItem = (id: number, body: MeterItemInput) =>
  request<MeterItem>(`/meter-items/${id}`, { method: 'PUT', body })
