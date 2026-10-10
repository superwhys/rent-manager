import { request } from './request'
import type { HomeSummary } from '@/types/home'

export const getHomeSummary = () => request<HomeSummary>('/home')
