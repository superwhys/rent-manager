export interface RoomType {
  id: number
  account_id: string
  name: string
}

export interface RoomTypeQuery {
  name?: string
  limit?: number
  offset?: number
}
