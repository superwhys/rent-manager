import { request } from './request'
import type {
  Account,
  EmailCodeResult,
  EmailLoginRequest,
  LoginResult,
  PasswordLoginRequest,
  RegisterRequest,
  Session,
  UpdateAccountRequest,
} from '@/types/auth'

export const sendEmailCode = (email: string) =>
  request<EmailCodeResult>('/auth/email/code', {
    method: 'POST',
    body: { email },
    token: '',
  })
export const registerEmail = (body: RegisterRequest, signal?: AbortSignal) =>
  request<LoginResult>('/auth/register', {
    method: 'POST',
    body,
    signal,
    token: '',
  })
export const loginPassword = (
  body: PasswordLoginRequest,
  signal?: AbortSignal,
) =>
  request<LoginResult>('/auth/password-login', {
    method: 'POST',
    body,
    signal,
    token: '',
  })
export const loginEmail = (body: EmailLoginRequest, signal?: AbortSignal) =>
  request<LoginResult>('/auth/email-login', {
    method: 'POST',
    body,
    signal,
    token: '',
  })
export const getSession = (token?: string) =>
  request<Session>('/auth/session', { token })
export const updateAccount = (body: UpdateAccountRequest) =>
  request<Account>('/auth/account', { method: 'PATCH', body })
export const revokeSession = () =>
  request<null>('/auth/session', { method: 'DELETE' })
