import { request } from '@/api/request'
import type { AccountSettings, SettingInput } from '@/types/settings'

export const getSettings = () => request<AccountSettings>('/settings')
export const updateSetting = (body: SettingInput) =>
  request<AccountSettings>('/settings', { method: 'PUT', body })
