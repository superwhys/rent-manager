export interface RoomType {
  id: number
  account_id: string
  name: string
}

export interface RoomTypeQuery {
  unit_id?: number
  name?: string
  limit?: number
  offset?: number
}
