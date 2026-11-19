import { useToast } from 'primevue/usetoast'
import { errorMessage } from '@/api/client'

/** Toast helpers for success and API errors. */
export function useNotify() {
  const toast = useToast()
  return {
    ok: (summary: string, detail?: string) => toast.add({ severity: 'success', summary, detail, life: 3500 }),
    fail: (e: unknown, summary = 'Hiba történt') =>
      toast.add({ severity: 'error', summary, detail: errorMessage(e), life: 6000 }),
  }
}
