export interface Account {
  ID: string
  Username: string
  Email: string
}

export interface Session {
  account: Account
  expires_at: string
}

export interface LoginResult extends Session {
  access_token: string
  created: boolean
}

export interface PasswordLoginRequest {
  email: string
  password: string
}

export interface EmailLoginRequest {
  email: string
  code: string
}

export interface RegisterRequest extends PasswordLoginRequest {
  code: string
}

export interface EmailCodeResult {
  resend_after_seconds: number
  expires_in_seconds: number
}
