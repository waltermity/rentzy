import { ref } from 'vue'
import { defineStore } from 'pinia'
import { rentalApi, reservationApi } from '@/api'
import type { Condition, Rental, Reservation } from '@/api/types'

/** Staff operations: open rentals (to take back) and open reservations (to hand over). */
export const useRentalStore = defineStore('rentals', () => {
  const openRentals = ref<Rental[]>([])
  const upcoming = ref<Reservation[]>([])
  const history = ref<Rental[]>([])
  const loading = ref(false)

  async function load() {
    loading.value = true
    try {
      const [open, res, closed] = await Promise.all([
        rentalApi.list({ status: 'ACTIVE,OVERDUE' }),
        reservationApi.list({ status: 'CONFIRMED,PENDING' }),
        rentalApi.list({ status: 'RETURNED,CANCELLED' }),
      ])
      openRentals.value = open.sort((a, b) => a.expected_end_date.localeCompare(b.expected_end_date))
      upcoming.value = res.sort((a, b) => a.start_date.localeCompare(b.start_date))
      history.value = closed
    } finally {
      loading.value = false
    }
  }

  async function handover(reservationId: number, condition?: Condition) {
    const r = await rentalApi.handover(reservationId, condition)
    await load()
    return r
  }

  async function giveBack(id: number, condition: Condition, operatingHours?: number | null) {
    const res = await rentalApi.return(id, { condition_at_return: condition, operating_hours: operatingHours })
    await load()
    return res
  }

  async function cancel(id: number, condition?: Condition) {
    const r = await rentalApi.cancel(id, condition)
    await load()
    return r
  }

  return { openRentals, upcoming, history, loading, load, handover, giveBack, cancel }
})
