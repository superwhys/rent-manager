export interface HomeSummary {
  building_room_summary: {
    building_count: number
    room_count: number
    rented_room_count: number
    vacant_room_count: number
  }
  income_amount: string
  income_decimal: string
  monthly_growth: string
  current_month: string
  stats: { value: string; label: string; tone: string }[]
  todos: { title: string; description: string; tag: string }[]
}
