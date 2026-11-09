import axios, { AxiosError } from 'axios'
import { translateError } from '@/i18n/hu'

// All requests go to /api on the same origin (Vite proxies it in development).
export const http = axios.create({ baseURL: '/api' })

const TOKEN_KEY = 'rentzy.token'

export function getToken(): string | null {
  try {
    return localStorage.getItem(TOKEN_KEY)
  } catch {
    return null
  }
}

export function setToken(token: string | null) {
  try {
    if (token) localStorage.setItem(TOKEN_KEY, token)
    else localStorage.removeItem(TOKEN_KEY)
  } catch {
    // storage blocked (private mode): the session lasts until reload
  }
}

http.interceptors.request.use((config) => {
  const token = getToken()
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

/** Error with the API's error code ("not_found", "unavailable", ...) and message. */
export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
  ) {
    super(message)
  }
}

// Called on 401 for an authenticated request; the auth store registers it.
let onUnauthorized: (() => void) | null = null
export function setUnauthorizedHandler(fn: () => void) {
  onUnauthorized = fn
}

http.interceptors.response.use(
  (res) => res,
  (err: AxiosError<{ error?: { code: string; message: string } }>) => {
    const status = err.response?.status ?? 0
    const body = err.response?.data?.error
    if (status === 401 && err.config?.headers?.Authorization && onUnauthorized) {
      onUnauthorized()
    }
    if (!err.response) {
      return Promise.reject(new ApiError(0, 'network', translateError('network', '')))
    }
    return Promise.reject(new ApiError(status, body?.code ?? 'error', body?.message ?? err.message))
  },
)

/** Unwraps the API's {"data": ...} envelope. */
export async function get<T>(url: string, params?: Record<string, unknown>): Promise<T> {
  return (await http.get<{ data: T }>(url, { params })).data.data
}
export async function post<T>(url: string, body?: unknown): Promise<T> {
  return (await http.post<{ data: T }>(url, body ?? {})).data.data
}
export async function put<T>(url: string, body: unknown): Promise<T> {
  return (await http.put<{ data: T }>(url, body)).data.data
}
export async function patch<T>(url: string, body: unknown): Promise<T> {
  return (await http.patch<{ data: T }>(url, body)).data.data
}

/** A readable Hungarian message for any error thrown by the API helpers. */
export function errorMessage(e: unknown): string {
  if (e instanceof ApiError) return translateError(e.code, e.message)
  return e instanceof Error ? e.message : String(e)
}
