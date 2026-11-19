<script setup lang="ts">
import { onMounted, ref } from 'vue'
import Button from 'primevue/button'
import Card from 'primevue/card'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import Password from 'primevue/password'
import Tabs from 'primevue/tabs'
import TabList from 'primevue/tablist'
import Tab from 'primevue/tab'
import TabPanels from 'primevue/tabpanels'
import TabPanel from 'primevue/tabpanel'
import { useConfirm } from 'primevue/useconfirm'
import StatusTag from '@/components/StatusTag.vue'
import { authApi } from '@/api'
import type { Reservation } from '@/api/types'
import { useAuthStore } from '@/stores/auth'
import { useReservationStore } from '@/stores/reservations'
import { date, dateTime, money } from '@/utils/format'
import { useNotify } from '@/utils/notify'

const auth = useAuthStore()
const store = useReservationStore()
const confirm = useConfirm()
const notify = useNotify()

onMounted(async () => {
  if (auth.isCustomer) {
    try {
      await store.load()
    } catch (e) {
      notify.fail(e, 'A foglalások betöltése nem sikerült')
    }
  }
})

function cancel(r: Reservation) {
  confirm.require({
    header: 'Lemondja a foglalást?',
    message: `${r.tool_type_name}, ${dateTime(r.start_date)} – ${dateTime(r.end_date)}`,
    icon: 'pi pi-exclamation-triangle',
    rejectProps: { label: 'Megtartom', severity: 'secondary', outlined: true },
    acceptProps: { label: 'Foglalás lemondása', severity: 'danger' },
    accept: async () => {
      try {
        await store.cancel(r.id)
        notify.ok('Foglalás lemondva')
      } catch (e) {
        notify.fail(e, 'A lemondás nem sikerült')
      }
    },
  })
}

// ---------- password ----------
const oldPw = ref('')
const newPw = ref('')
const newPw2 = ref('')
const pwBusy = ref(false)

async function changePassword() {
  if (newPw.value !== newPw2.value) {
    notify.fail(new Error('A két új jelszó nem egyezik.'), 'A jelszó nem változott')
    return
  }
  pwBusy.value = true
  try {
    await authApi.changePassword(oldPw.value, newPw.value)
    oldPw.value = newPw.value = newPw2.value = ''
    notify.ok('Jelszó megváltoztatva')
  } catch (e) {
    notify.fail(e, 'A jelszó nem változott')
  } finally {
    pwBusy.value = false
  }
}
</script>

<template>
  <div class="page">
    <div class="page-head">
      <h1>{{ auth.displayName }}</h1>
      <StatusTag v-if="auth.role" :value="auth.role" />
    </div>

    <div class="cards">
      <Card>
        <template #title>Fiókadatok</template>
        <template #content>
          <dl class="facts">
            <dt>Felhasználónév</dt>
            <dd>{{ auth.user?.username }}</dd>
            <dt>E-mail cím</dt>
            <dd>{{ auth.user?.email }}</dd>
            <template v-if="auth.customer">
              <dt>Név</dt>
              <dd>{{ auth.customer.name }}</dd>
              <dt>Telefonszám</dt>
              <dd>{{ auth.customer.phone || '—' }}</dd>
              <dt>Cím</dt>
              <dd>{{ auth.customer.address || '—' }}</dd>
            </template>
            <dt>Regisztráció dátuma</dt>
            <dd>{{ date(auth.user?.created_at) }}</dd>
          </dl>
        </template>
      </Card>

      <Card>
        <template #title>Jelszó módosítása</template>
        <template #content>
          <form class="stack" @submit.prevent="changePassword">
            <div class="field">
              <label for="old">Jelenlegi jelszó</label>
              <Password input-id="old" v-model="oldPw" :feedback="false" toggle-mask fluid
                autocomplete="current-password" required />
            </div>
            <div class="field">
              <label for="new">Új jelszó</label>
              <Password input-id="new" v-model="newPw" toggle-mask fluid autocomplete="new-password" required />
            </div>
            <div class="field">
              <label for="new2">Új jelszó még egyszer</label>
              <Password input-id="new2" v-model="newPw2" :feedback="false" toggle-mask fluid
                autocomplete="new-password" required />
            </div>
            <Button type="submit" label="Jelszó módosítása" icon="pi pi-key" :loading="pwBusy" />
          </form>
        </template>
      </Card>
    </div>

    <template v-if="auth.isCustomer">
      <h2 class="section-title">Foglalásaim és kölcsönzéseim</h2>
      <Tabs value="open">
        <TabList>
          <Tab value="open">Közelgő foglalások ({{ store.open.length }})</Tab>
          <Tab value="rentals">Kölcsönzések ({{ store.rentals.length }})</Tab>
          <Tab value="past">Korábbi foglalások ({{ store.past.length }})</Tab>
        </TabList>
        <TabPanels>
          <TabPanel value="open">
            <DataTable :value="store.open" :loading="store.loading" data-key="id" striped-rows>
              <template #empty>
                <div class="empty">
                  Nincs közelgő foglalása. <RouterLink to="/catalog">Böngésszen a katalógusban</RouterLink>
                </div>
              </template>
              <Column field="tool_type_name" header="Eszköz" />
              <Column header="Átvétel"><template #body="{ data }">{{ dateTime(data.start_date) }}</template></Column>
              <Column header="Visszahozatal"><template #body="{ data }">{{ dateTime(data.end_date) }}</template></Column>
              <Column header="Állapot"><template #body="{ data }"><StatusTag :value="data.status" /></template></Column>
              <Column header="" class="num">
                <template #body="{ data }">
                  <Button label="Lemondás" icon="pi pi-times" severity="danger" text size="small" @click="cancel(data)" />
                </template>
              </Column>
            </DataTable>
          </TabPanel>
          <TabPanel value="rentals">
            <DataTable :value="store.rentals" :loading="store.loading" data-key="id" striped-rows>
              <template #empty><div class="empty">Még nincs kölcsönzése.</div></template>
              <Column field="tool_type_name" header="Eszköz" />
              <Column header="Kezdete"><template #body="{ data }">{{ dateTime(data.start_date) }}</template></Column>
              <Column header="Határidő"><template #body="{ data }">{{ dateTime(data.expected_end_date) }}</template></Column>
              <Column header="Visszahozva"><template #body="{ data }">{{ dateTime(data.actual_end_date) }}</template></Column>
              <Column header="Kaució" class="num"><template #body="{ data }">{{ money(data.deposit_paid) }}</template></Column>
              <Column header="Késedelmi díj" class="num"><template #body="{ data }">{{ money(data.late_fee) }}</template></Column>
              <Column header="Fizetendő" class="num"><template #body="{ data }">{{ money(data.total_fee) }}</template></Column>
              <Column header="Állapot"><template #body="{ data }"><StatusTag :value="data.status" /></template></Column>
            </DataTable>
          </TabPanel>
          <TabPanel value="past">
            <DataTable :value="store.past" :loading="store.loading" data-key="id" striped-rows>
              <template #empty><div class="empty">Még nincs korábbi foglalása.</div></template>
              <Column field="tool_type_name" header="Eszköz" />
              <Column header="Átvétel"><template #body="{ data }">{{ dateTime(data.start_date) }}</template></Column>
              <Column header="Visszahozatal"><template #body="{ data }">{{ dateTime(data.end_date) }}</template></Column>
              <Column header="Állapot"><template #body="{ data }"><StatusTag :value="data.status" /></template></Column>
            </DataTable>
          </TabPanel>
        </TabPanels>
      </Tabs>
    </template>
  </div>
</template>

<style scoped>
.cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 1rem;
}
.facts {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 0.5rem 1rem;
  margin: 0;
}
.facts dt {
  color: var(--app-muted);
}
.facts dd {
  margin: 0;
  font-weight: 500;
  overflow-wrap: anywhere;
}
</style>
