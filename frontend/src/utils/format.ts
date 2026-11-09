import { enumLabels } from '@/i18n/hu'

const dateFmt = new Intl.DateTimeFormat('hu-HU', { year: 'numeric', month: 'short', day: 'numeric' })
const dateTimeFmt = new Intl.DateTimeFormat('hu-HU', {
  year: 'numeric',
  month: 'short',
  day: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
})

const NBSP = '\u00a0'

/** 4500 -> "4 500 Ft": whole forints, space as thousands separator (also for 4 digits). */
export function money(v: number | null | undefined): string {
  if (v == null) return '—'
  const n = Math.round(v)
  const digits = Math.abs(n).toString().replace(/\B(?=(\d{3})+(?!\d))/g, NBSP)
  return `${n < 0 ? '−' : ''}${digits}${NBSP}Ft`
}
/** "2026. nov. 9." */
export const date = (v: string | null | undefined) => (v ? dateFmt.format(new Date(v)) : '—')
/** "2026. nov. 9. 08:00" */
export const dateTime = (v: string | null | undefined) => (v ? dateTimeFmt.format(new Date(v)) : '—')
export const percent = (v: number) => `${Math.round(v * 100)}%`

/** Hungarian name of an enum value: "AVAILABLE" -> "Elérhető". */
export const label = (v: string) => enumLabels[v] ?? v

/** Days between two instants, counting every started 24 hours (like the backend's fee rule). */
export function startedDays(from: Date, to: Date): number {
  const ms = to.getTime() - from.getTime()
  return ms <= 0 ? 0 : Math.ceil(ms / 86_400_000)
}

export function startOfDay(d: Date): Date {
  const x = new Date(d)
  x.setHours(0, 0, 0, 0)
  return x
}

export function addDays(d: Date, n: number): Date {
  const x = new Date(d)
  x.setDate(x.getDate() + n)
  return x
}

/** YYYY-MM-DD in local time, for query parameters. */
export function isoDay(d: Date): string {
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}`
}
