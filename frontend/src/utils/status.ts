// PrimeVue Tag severities for the status enums.
type Severity = 'success' | 'info' | 'warn' | 'danger' | 'secondary' | 'contrast'

const severities: Record<string, Severity> = {
  // tool items
  AVAILABLE: 'success',
  RENTED: 'info',
  MAINTENANCE: 'warn',
  BROKEN: 'danger',
  RETIRED: 'secondary',
  // reservations
  PENDING: 'warn',
  CONFIRMED: 'success',
  CANCELLED: 'secondary',
  CONVERTED: 'info',
  EXPIRED: 'secondary',
  // rentals
  ACTIVE: 'info',
  RETURNED: 'success',
  OVERDUE: 'danger',
  // roles
  ADMIN: 'contrast',
  STAFF: 'info',
  CUSTOMER: 'secondary',
  // conditions
  NEW: 'success',
  GOOD: 'success',
  FAIR: 'info',
  POOR: 'warn',
  DAMAGED: 'danger',
}

export const severity = (status: string): Severity => severities[status] ?? 'secondary'
