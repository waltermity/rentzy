import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { rentalApi, reservationApi } from '@/api'
import type { Rental, Reservation } from '@/api/types'

/** The logged-in customer's reservations and rentals. */
export const useReservationStore = defineStore('reservations', () => {
  const reservations = ref<Reservation[]>([])
  const rentals = ref<Rental[]>([])
  const loading = ref(false)

  const open = computed(() => reservations.value.filter((r) => r.status === 'CONFIRMED' || r.status === 'PENDING'))
  const past = computed(() => reservations.value.filter((r) => !(r.status === 'CONFIRMED' || r.status === 'PENDING')))

  async function load() {
    loading.value = true
    try {
      const [res, ren] = await Promise.all([reservationApi.list(), rentalApi.list()])
      reservations.value = res
      rentals.value = ren
    } finally {
      loading.value = false
    }
  }

  async function create(toolTypeId: number, start: Date, end: Date) {
    const r = await reservationApi.create({
      tool_type_id: toolTypeId,
      start_date: start.toISOString(),
      end_date: end.toISOString(),
    })
    reservations.value.unshift(r)
    return r
  }

  async function cancel(id: number) {
    const r = await reservationApi.cancel(id)
    const i = reservations.value.findIndex((x) => x.id === id)
    if (i >= 0) reservations.value[i] = r
  }

  return { reservations, rentals, loading, open, past, load, create, cancel }
})
