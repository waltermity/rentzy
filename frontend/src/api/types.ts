// Mirrors the JSON of the Go API (backend/internal/domain and service).
// Timestamps are RFC 3339 strings; money is a JSON number (HUF).

export type Role = 'ADMIN' | 'STAFF' | 'CUSTOMER'
export type ToolStatus = 'AVAILABLE' | 'RENTED' | 'MAINTENANCE' | 'BROKEN' | 'RETIRED'
export type Condition = 'NEW' | 'GOOD' | 'FAIR' | 'POOR' | 'DAMAGED'
export type ReservationStatus = 'PENDING' | 'CONFIRMED' | 'CANCELLED' | 'CONVERTED' | 'EXPIRED'
export type RentalStatus = 'ACTIVE' | 'RETURNED' | 'OVERDUE' | 'CANCELLED'
export type MaintenanceType = 'SCHEDULED' | 'REPAIR' | 'INSPECTION'

export const CONDITIONS: Condition[] = ['NEW', 'GOOD', 'FAIR', 'POOR', 'DAMAGED']
export const MAINTENANCE_TYPES: MaintenanceType[] = ['SCHEDULED', 'REPAIR', 'INSPECTION']

export interface User {
  id: number
  username: string
  email: string
  role: Role
  customer_id: number | null
  created_at: string
}

export interface Customer {
  id: number
  name: string
  email: string
  phone: string
  address: string
  created_at: string
}

export interface Session {
  token: string
  expires_at: string
  user: User
  customer?: Customer
}

export interface Category {
  id: number
  name: string
}

export interface ToolType {
  id: number
  name: string
  category_id: number
  category_name?: string
  manufacturer: string
  model: string
  purchase_price: number
  daily_rate: number
  deposit: number
  maintenance_every_rentals: number | null
  maintenance_every_days: number | null
  maintenance_every_hours: number | null
  item_counts?: Partial<Record<ToolStatus, number>>
}

export interface ToolItem {
  id: number
  tool_type_id: number
  tool_type_name?: string
  status: ToolStatus
  current_condition: Condition
  acquired_at: string
}

export interface MaintenanceDue {
  tool_item_id: number
  tool_type_id: number
  tool_type_name: string
  status: ToolStatus
  last_maintenance_at: string | null
  since: string
  rentals_since: number
  operating_hours_since: number
  maintenance_every_rentals: number | null
  maintenance_every_days: number | null
  maintenance_every_hours: number | null
  days_since: number
  due: boolean
  reasons: ('RENTAL_COUNT' | 'ELAPSED_DAYS' | 'OPERATING_HOURS')[]
}

export interface ToolItemDetail extends ToolItem {
  maintenance: MaintenanceDue | null
}

export interface BusyInterval {
  tool_item_id: number
  start: string
  end: string
  source: 'reservation' | 'rental' | 'rental_overdue'
}

export interface Availability {
  tool_type_id: number
  from: string
  to: string
  bookable_item_ids: number[]
  available_item_ids: number[]
  busy: BusyInterval[]
}

export interface Reservation {
  id: number
  customer_id: number
  customer_name?: string
  tool_item_id: number
  tool_type_id: number
  tool_type_name?: string
  start_date: string
  end_date: string
  status: ReservationStatus
  created_at: string
}

export interface Rental {
  id: number
  customer_id: number
  customer_name?: string
  tool_item_id: number
  tool_type_id: number
  tool_type_name?: string
  reservation_id: number
  staff_id: number
  start_date: string
  expected_end_date: string
  actual_end_date: string | null
  deposit_paid: number
  condition_at_handover: Condition
  condition_at_return: Condition | null
  base_fee: number | null
  late_fee: number | null
  total_fee: number | null
  operating_hours: number | null
  status: RentalStatus
}

export interface Fees {
  base_fee: number
  late_fee: number
  total_fee: number
  late_days: number
}

export interface ReturnResult {
  rental: Rental
  fees: Fees
}

export interface Maintenance {
  id: number
  tool_item_id: number
  tool_type_name?: string
  type: MaintenanceType
  description: string
  started_at: string
  completed_at: string | null
  cost: number
  performed_by: string
}

export interface ReportWindow {
  from: string
  to: string
}

export interface TypeStats {
  tool_type_id: number
  tool_type_name: string
  category_name: string
  item_count: number
  rental_count: number
  rented_hours: number
  utilization: number
  revenue: number
  purchase_price: number
}

export interface ItemUtilization {
  tool_item_id: number
  tool_type_id: number
  tool_type_name: string
  category_name: string
  status: ToolStatus
  rental_count: number
  rented_hours: number
  capacity_hours: number
  revenue: number
  utilization: number
}

export interface PopularReport {
  window: ReportWindow
  types: TypeStats[]
}

export interface UtilizationReport {
  window: ReportWindow
  total_revenue: number
  types: TypeStats[]
  items: ItemUtilization[]
}

export interface ProcurementSuggestion extends TypeStats {
  suggested_additional: number
  reason: string
  payback_days: number | null
}

export interface ProcurementReport {
  window: ReportWindow
  threshold: number
  suggestions: ProcurementSuggestion[]
}
