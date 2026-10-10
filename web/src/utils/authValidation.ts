export function emailError(value: string): string {
  const email = value.trim()
  if (!email) return '请输入邮箱地址'
  if (email.length > 254 || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email))
    return '请输入有效的邮箱地址'
  return ''
}

export function verificationCodeError(value: string): string {
  return /^\d{6}$/.test(value) ? '' : '请输入邮件中的 6 位验证码'
}

export function registrationPasswordError(
  password: string,
  confirmation: string,
): string {
  const length = Array.from(password).length
  if (length < 8 || length > 128) return '密码长度应为 8–128 个字符'
  if (password !== confirmation) return '两次输入的密码不一致'
  return ''
}

export function loginPasswordError(password: string): string {
  return password ? '' : '请输入密码'
}

export function safeAuthRedirect(value: unknown): string {
  if (typeof value !== 'string') return '/dashboard'
  try {
    const decoded = decodeURIComponent(value)
    const path = decoded.split(/[?#]/)[0]
    if (
      !decoded.startsWith('/') ||
      decoded.startsWith('//') ||
      /[\\\u0000-\u001f]/.test(decoded) ||
      /^\/(login|register)(\/|$)/.test(path)
    )
      return '/dashboard'
    return value
  } catch {
    return '/dashboard'
  }
}

export function accountNameError(value: string): string {
  if (/[\u0000-\u001f\u007f-\u009f]/u.test(value))
    return '账号名称不能包含控制字符'
  const name = value.trim()
  if (!name) return '请输入账号名称'
  if ([...name].length > 64) return '账号名称不能超过 64 个字符'
  return ''
}
