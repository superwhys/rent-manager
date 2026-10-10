export interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  body?: unknown
  query?: Record<string, string | number | boolean | undefined>
  token?: string
  signal?: AbortSignal
}

export class ApiError extends Error {
  code: number
  status: number

  constructor(message: string, code: number, status: number) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.status = status
  }
}

let sessionToken = ''
let onSessionExpired: (() => void) | undefined

export function configureSession(token: string, onExpired?: () => void) {
  sessionToken = token
  if (onExpired) onSessionExpired = onExpired
}

function responseMessage(value: unknown, fallback: string): string {
  if (typeof value === 'string' && value.trim()) return value
  return fallback
}

export async function request<T>(
  path: string,
  options: RequestOptions = {},
): Promise<T> {
  const query = new URLSearchParams()
  for (const [key, value] of Object.entries(options.query ?? {})) {
    if (value !== undefined && value !== '') query.set(key, String(value))
  }
  const token = options.token ?? sessionToken
  const base = (import.meta.env?.VITE_API_BASE_URL || '/api').replace(/\/$/, '')
  const headers: Record<string, string> = { Accept: 'application/json' }
  if (token) headers.Authorization = `Bearer ${token}`
  if (options.body !== undefined) headers['Content-Type'] = 'application/json'

  let response: Response
  try {
    response = await fetch(`${base}${path}${query.size ? `?${query}` : ''}`, {
      method: options.method ?? 'GET',
      headers,
      body:
        options.body === undefined ? undefined : JSON.stringify(options.body),
      signal: options.signal,
      cache: 'no-store',
    })
  } catch (error) {
    if (error instanceof Error && error.name === 'AbortError') throw error
    throw new Error('无法连接服务，请检查网络后重试')
  }

  const payload: unknown = await response.json().catch(() => null)
  const envelope =
    typeof payload === 'object' && payload !== null ? payload : {}
  const code =
    'code' in envelope && typeof envelope.code === 'number' ? envelope.code : -1
  if (!response.ok || code !== 0) {
    const expired =
      response.status === 401 || code === 400002 || code === 400004
    if (expired && options.token === undefined && token === sessionToken)
      onSessionExpired?.()
    const fallback = expired
      ? '登录已失效，请重新登录'
      : response.status === 403
        ? '你没有执行此操作的权限'
        : response.status === 429
          ? '操作较频繁，请稍后重试'
          : '请求未完成，请稍后重试'
    throw new ApiError(
      responseMessage(
        'message' in envelope ? envelope.message : undefined,
        fallback,
      ),
      code,
      response.status,
    )
  }
  if (!('data' in envelope)) throw new Error('服务返回的数据不完整，请重试')
  return envelope.data as T
}
