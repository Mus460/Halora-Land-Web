const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
const API_PREFIX = '/api/v1'

type FetchOptions = RequestInit & {
  auth?: boolean
}

async function buildHeaders(custom: HeadersInit = {}): Promise<HeadersInit> {
  // Auth is cookie-based: the backend reads the `halora_session` HttpOnly
  // cookie set at login (credentials: 'include' below).
  return {
    'Content-Type': 'application/json',
    ...custom,
  }
}

function apiUrl(path: string): string {
  if (path.startsWith('http')) return path
  return `${API_URL}${API_PREFIX}${path.startsWith('/') ? path : `/${path}`}`
}

export async function api<T = any>(
  path: string,
  options: FetchOptions = {},
): Promise<T> {
  const { auth: _auth, headers, ...rest } = options
  const finalHeaders = await buildHeaders(headers)

  const res = await fetch(apiUrl(path), {
    ...rest,
    headers: finalHeaders,
    credentials: 'include',
  })

  if (res.status === 204) return undefined as T

  const text = await res.text()
  const data = text ? JSON.parse(text) : null

  if (!res.ok) {
    const message = data?.error || res.statusText
    const err = new Error(message) as Error & { status: number; data: any }
    err.status = res.status
    err.data = data
    throw err
  }

  return data as T
}

export const apiClient = {
  get: <T = any>(path: string, options?: FetchOptions) =>
    api<T>(path, { ...options, method: 'GET' }),
  post: <T = any>(path: string, body?: any, options?: FetchOptions) =>
    api<T>(path, { ...options, method: 'POST', body: body ? JSON.stringify(body) : undefined }),
  put: <T = any>(path: string, body?: any, options?: FetchOptions) =>
    api<T>(path, { ...options, method: 'PUT', body: body ? JSON.stringify(body) : undefined }),
  delete: <T = any>(path: string, options?: FetchOptions) =>
    api<T>(path, { ...options, method: 'DELETE' }),
}

export { API_URL, API_PREFIX }
