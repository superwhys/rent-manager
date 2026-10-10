export type CalculationType = 'DIFFERENCE' | 'DIRECT'

export interface MeterItem {
  id: number
  account_id: string
  code: string
  unit: string
  unit_price: number
  is_system: boolean
  calc_type: CalculationType
  remark?: string | null
}

export interface MeterItemInput {
  code: string
  unit: string
  unit_price: number
  calc_type: CalculationType
  remark: string | null
}

export interface MeterItemQuery {
  unit_id?: number
  limit?: number
  offset?: number
  code?: string
  is_system?: boolean
  calc_type?: CalculationType
}
