<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import Button from 'primevue/button'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import Dialog from 'primevue/dialog'
import IconField from 'primevue/iconfield'
import InputIcon from 'primevue/inputicon'
import InputText from 'primevue/inputtext'
import StatusTag from '@/components/StatusTag.vue'
import { customerApi, rentalApi, reservationApi } from '@/api'
import type { Customer, Rental, Reservation } from '@/api/types'
import { useAuthStore } from '@/stores/auth'
import { date, dateTime, money } from '@/utils/format'
import { useNotify } from '@/utils/notify'

const auth = useAuthStore()
const notify = useNotify()

const customers = ref<Customer[]>([])
const loading = ref(false)
const search = ref('')

async function load() {
  loading.value = true
  try {
    customers.value = await customerApi.list()
  } catch (e) {
    notify.fail(e, 'Az ügyfelek betöltése nem sikerült')
  } finally {
    loading.value = false
  }
}
onMounted(load)

const visible = computed(() => {
  const q = search.value.trim().toLowerCase()
  if (!q) return customers.value
  return customers.value.filter((c) => [c.name, c.email, c.phone].some((s) => s.toLowerCase().includes(q)))
})

// ---------- create / edit (admin) ----------
const ed = reactive({ open: false, id: null as number | null, name: '', email: '', phone: '', address: '', busy: false })

function openEdit(c?: Customer) {
  Object.assign(ed, c ? { open: true, id: c.id, name: c.name, email: c.email, phone: c.phone, address: c.address }
    : { open: true, id: null, name: '', email: '', phone: '', address: '' })
}

async function save() {
  ed.busy = true
  const body = { name: ed.name, email: ed.email, phone: ed.phone, address: ed.address }
  try {
    if (ed.id) await customerApi.update(ed.id, body)
    else await customerApi.create(body)
    notify.ok(ed.id ? 'Ügyfél módosítva' : 'Ügyfél létrehozva')
    ed.open = false
    await load()
  } catch (e) {
    notify.fail(e, 'A mentés nem sikerült')
  } finally {
    ed.busy = false
  }
}

// ---------- details ----------
const detail = reactive({ customer: null as Customer | null, reservations: [] as Reservation[], rentals: [] as Rental[] })
async function openDetail(c: Customer) {
  Object.assign(detail, { customer: c, reservations: [], rentals: [] })
  try {
    const [res, ren] = await Promise.all([
      reservationApi.list({ customer_id: c.id }),
      rentalApi.list({ customer_id: c.id }),
    ])
    detail.reservations = res
    detail.rentals = ren
  } catch (e) {
    notify.fail(e, 'Az előzmények betöltése nem sikerült')
  }
}
</script>

<template>
  <div class="page">
    <div class="page-head">
      <h1>Ügyfelek</h1>
      <div class="row">
        <IconField>
          <InputIcon class="pi pi-search" />
          <InputText v-model="search" placeholder="Keresés: név, e-mail, telefon" aria-label="Ügyfél keresése" />
        </IconField>
        <Button v-if="auth.isAdmin" label="Új ügyfél" icon="pi pi-plus" @click="openEdit()" />
      </div>
    </div>

    <DataTable :value="visible" :loading="loading" data-key="id" striped-rows paginator :rows="20"
      selection-mode="single" @row-click="openDetail($event.data)" class="clickable">
      <template #empty><div class="empty">Nincs találat.</div></template>
      <Column field="id" header="#" />
      <Column field="name" header="Név" sortable />
      <Column field="email" header="E-mail" />
      <Column field="phone" header="Telefon" />
      <Column field="address" header="Cím" />
      <Column header="Regisztrált"><template #body="{ data }">{{ date(data.created_at) }}</template></Column>
      <Column v-if="auth.isAdmin" header="" class="num">
        <template #body="{ data }">
          <Button icon="pi pi-pencil" text size="small" aria-label="Szerkesztés" @click.stop="openEdit(data)" />
        </template>
      </Column>
    </DataTable>

    <Dialog v-model:visible="ed.open" modal :header="ed.id ? 'Ügyfél szerkesztése' : 'Új ügyfél'"
      :style="{ width: 'min(560px, 95vw)' }">
      <form id="cust-form" class="form-grid" @submit.prevent="save">
        <div class="field"><label for="c-name">Név</label><InputText id="c-name" v-model="ed.name" required /></div>
        <div class="field"><label for="c-email">E-mail cím</label><InputText id="c-email" v-model="ed.email" type="email" required /></div>
        <div class="field"><label for="c-phone">Telefonszám</label><InputText id="c-phone" v-model="ed.phone" /></div>
        <div class="field"><label for="c-addr">Cím</label><InputText id="c-addr" v-model="ed.address" /></div>
      </form>
      <p v-if="!ed.id" class="muted small">Ez bejelentkezés nélküli ügyfelet hoz létre. Fiókot a Felhasználók oldalon adhat neki.</p>
      <template #footer>
        <Button label="Bezárás" severity="secondary" text @click="ed.open = false" />
        <Button type="submit" form="cust-form" label="Mentés" icon="pi pi-check" :loading="ed.busy" />
      </template>
    </Dialog>

    <Dialog :visible="!!detail.customer" modal :header="detail.customer?.name" :style="{ width: 'min(860px, 95vw)' }"
      @update:visible="detail.customer = null">
      <h3 class="sub">Foglalások</h3>
      <DataTable :value="detail.reservations" data-key="id" size="small">
        <template #empty><div class="empty">Nincs.</div></template>
        <Column field="tool_type_name" header="Eszköz" />
        <Column header="Időszak"><template #body="{ data }">{{ dateTime(data.start_date) }} – {{ dateTime(data.end_date) }}</template></Column>
        <Column header="Állapot"><template #body="{ data }"><StatusTag :value="data.status" /></template></Column>
      </DataTable>
      <h3 class="sub">Kölcsönzések</h3>
      <DataTable :value="detail.rentals" data-key="id" size="small">
        <template #empty><div class="empty">Nincs.</div></template>
        <Column field="tool_type_name" header="Eszköz" />
        <Column header="Kezdete"><template #body="{ data }">{{ dateTime(data.start_date) }}</template></Column>
        <Column header="Visszahozva"><template #body="{ data }">{{ dateTime(data.actual_end_date) }}</template></Column>
        <Column header="Fizetendő" class="num"><template #body="{ data }">{{ money(data.total_fee) }}</template></Column>
        <Column header="Állapot"><template #body="{ data }"><StatusTag :value="data.status" /></template></Column>
      </DataTable>
    </Dialog>
  </div>
</template>

<style scoped>
.clickable :deep(tbody tr) {
  cursor: pointer;
}
.sub {
  margin: 0.5rem 0;
  font-size: 1rem;
}
.small {
  font-size: 0.85rem;
}
</style>
