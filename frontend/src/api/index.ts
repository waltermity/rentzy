// One function per backend endpoint. See the Swagger UI at /docs.
import { get, patch, post, put } from './client'
import type * as T from './types'

type Q = Record<string, unknown>

export const authApi = {
  login: (username: string, password: string) => post<T.Session>('/auth/login', { username, password }),
  register: (body: {
    username: string
    email: string
    password: string
    name: string
    phone?: string
    address?: string
  }) => post<T.Session>('/auth/register', body),
  me: () => get<{ user: T.User; customer?: T.Customer }>('/auth/me'),
  changePassword: (old_password: string, new_password: string) =>
    post<{ status: string }>('/auth/password', { old_password, new_password }),
}

export type ToolTypeInput = Omit<T.ToolType, 'id' | 'category_name' | 'item_counts'>

export const toolApi = {
  categories: () => get<T.Category[]>('/categories'),
  createCategory: (name: string) => post<T.Category>('/categories', { name }),
  updateCategory: (id: number, name: string) => put<T.Category>(`/categories/${id}`, { name }),

  types: (q?: { category_id?: number; q?: string }) => get<T.ToolType[]>('/tool-types', q),
  type: (id: number) => get<T.ToolType>(`/tool-types/${id}`),
  createType: (body: ToolTypeInput) => post<T.ToolType>('/tool-types', body),
  updateType: (id: number, body: ToolTypeInput) => put<T.ToolType>(`/tool-types/${id}`, body),
  availability: (id: number, from: string, to: string) =>
    get<T.Availability>(`/tool-types/${id}/availability`, { from, to }),

  items: (q?: { tool_type_id?: number; status?: string }) => get<T.ToolItem[]>('/tool-items', q),
  item: (id: number) => get<T.ToolItemDetail>(`/tool-items/${id}`),
  addItems: (typeId: number, body: { count: number; condition?: T.Condition; acquired_at?: string }) =>
    post<T.ToolItem[]>(`/tool-types/${typeId}/items`, body),
  updateItem: (id: number, body: { status?: T.ToolStatus; condition?: T.Condition }) =>
    patch<T.ToolItem>(`/tool-items/${id}`, body),
}

export const reservationApi = {
  list: (q?: Q) => get<T.Reservation[]>('/reservations', q),
  get: (id: number) => get<T.Reservation>(`/reservations/${id}`),
  create: (body: {
    tool_type_id?: number
    tool_item_id?: number
    customer_id?: number
    start_date: string
    end_date: string
  }) => post<T.Reservation>('/reservations', body),
  cancel: (id: number) => post<T.Reservation>(`/reservations/${id}/cancel`),
}

export const rentalApi = {
  list: (q?: Q) => get<T.Rental[]>('/rentals', q),
  get: (id: number) => get<T.Rental>(`/rentals/${id}`),
  handover: (reservation_id: number, condition_at_handover?: T.Condition) =>
    post<T.Rental>('/rentals', { reservation_id, condition_at_handover }),
  return: (id: number, body: { condition_at_return: T.Condition; operating_hours?: number | null; returned_at?: string }) =>
    post<T.ReturnResult>(`/rentals/${id}/return`, body),
  cancel: (id: number, condition_at_return?: T.Condition) =>
    post<T.Rental>(`/rentals/${id}/cancel`, condition_at_return ? { condition_at_return } : undefined),
}

export const customerApi = {
  list: (q?: string) => get<T.Customer[]>('/customers', q ? { q } : undefined),
  get: (id: number) => get<T.Customer>(`/customers/${id}`),
  create: (body: Omit<T.Customer, 'id' | 'created_at'>) => post<T.Customer>('/customers', body),
  update: (id: number, body: Omit<T.Customer, 'id' | 'created_at'>) => put<T.Customer>(`/customers/${id}`, body),
}

export const userApi = {
  list: (role?: T.Role) => get<T.User[]>('/users', role ? { role } : undefined),
  create: (body: {
    username: string
    email: string
    password: string
    role: T.Role
    customer_id?: number
    customer?: Omit<T.Customer, 'id' | 'created_at'>
  }) => post<T.User>('/users', body),
  update: (id: number, body: { email?: string; role?: T.Role; password?: string }) =>
    patch<T.User>(`/users/${id}`, body),
}

export const maintenanceApi = {
  list: (q?: { tool_item_id?: number; open?: boolean }) => get<T.Maintenance[]>('/maintenance', q),
  history: (itemId: number) => get<T.Maintenance[]>(`/tool-items/${itemId}/maintenance`),
  due: (all = false) => get<T.MaintenanceDue[]>('/maintenance/due', all ? { all: true } : undefined),
  start: (itemId: number, body: { type: T.MaintenanceType; description?: string; performed_by?: string }) =>
    post<T.Maintenance>(`/tool-items/${itemId}/maintenance`, body),
  complete: (
    id: number,
    body: {
      cost: number
      description?: string
      performed_by?: string
      result_status?: T.ToolStatus
      condition?: T.Condition
    },
  ) => post<T.Maintenance>(`/maintenance/${id}/complete`, body),
}

export const reportApi = {
  popular: (q?: Q) => get<T.PopularReport>('/reports/popular', q),
  utilization: (q?: Q) => get<T.UtilizationReport>('/reports/utilization', q),
  procurement: (q?: Q) => get<T.ProcurementReport>('/reports/procurement', q),
}
