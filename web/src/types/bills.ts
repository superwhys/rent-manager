import type { MeterReadingLine } from '@/types/readings'

export type BillStatus = 'UNPAID' | 'PAID' | 'CANCELLED'

export interface RentBill {
  id: number
  room_id: number
  period_month: string
  base_rent: number
  meter_fee: number
  total_amount: number
  status: BillStatus
  paid_at?: string
  remark?: string
  created_at: string
  updated_at: string
}

export interface RentBillDetail extends RentBill {
  meter_readings: MeterReadingLine[] | null
}

export interface BillQuery {
  period_month?: string
  status?: BillStatus
  limit?: number
  offset?: number
}

export interface GenerateBillsInput {
  period_month: string
  remark?: string
}

export interface GeneratedBills {
  saved_bills: RentBill[]
  skipped_bills: RentBill[]
}
