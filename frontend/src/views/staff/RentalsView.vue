<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import Button from 'primevue/button'
import Checkbox from 'primevue/checkbox'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import DatePicker from 'primevue/datepicker'
import Dialog from 'primevue/dialog'
import InputNumber from 'primevue/inputnumber'
import Message from 'primevue/message'
import Select from 'primevue/select'
import Tabs from 'primevue/tabs'
import TabList from 'primevue/tablist'
import Tab from 'primevue/tab'
import TabPanels from 'primevue/tabpanels'
import TabPanel from 'primevue/tabpanel'
import { useConfirm } from 'primevue/useconfirm'
import StatusTag from '@/components/StatusTag.vue'
import { customerApi, reservationApi, toolApi } from '@/api'
import { CONDITIONS, type Condition, type Customer, type Rental, type Reservation, type ReturnResult, type ToolItemDetail, type ToolType } from '@/api/types'
import { useRentalStore } from '@/stores/rentals'
import { useToolStore } from '@/stores/tools'
import { addDays, dateTime, label, money, startedDays } from '@/utils/format'
import { useNotify } from '@/utils/notify'

const store = useRentalStore()
const tools = useToolStore()
const confirm = useConfirm()
const notify = useNotify()

const conditionOptions = CONDITIONS.map((c) => ({ label: label(c), value: c }))

async function refresh() {
  try {
    await Promise.all([store.load(), tools.loadCatalog()])
  } catch (e) {
    notify.fail(e, 'A kölcsönzések betöltése nem sikerült')
  }
}
onMounted(refresh)

const typeOf = (id: number): ToolType | undefined => tools.types.find((t) => t.id === id)
const isLate = (r: Rental) => r.status === 'OVERDUE' || new Date(r.expected_end_date) < new Date()

// ---------- handover ----------
const ho = reactive({
  open: false,
  busy: false,
  reservation: null as Reservation | null,
  item: null as ToolItemDetail | null,
  condition: 'GOOD' as Condition,
  depositReceived: false,
})

async function openHandover(r: Reservation) {
  Object.assign(ho, { open: true, reservation: r, item: null, depositReceived: false })
  try {
    ho.item = await toolApi.item(r.tool_item_id)
    ho.condition = ho.item.current_condition
  } catch (e) {
    notify.fail(e, 'A példány betöltése nem sikerült')
  }
}

const handoverBlocked = computed(() => ho.item != null && ho.item.status !== 'AVAILABLE')

async function doHandover() {
  if (!ho.reservation) return
  ho.busy = true
  try {
    const r = await store.handover(ho.reservation.id, ho.condition)
    notify.ok('Eszköz kiadva', `#${r.id} kölcsönzés, határidő: ${dateTime(r.expected_end_date)}`)
    ho.open = false
  } catch (e) {
    notify.fail(e, 'A kiadás nem sikerült')
  } finally {
    ho.busy = false
  }
}

// ---------- return ----------
const rt = reactive({
  open: false,
  busy: false,
  rental: null as Rental | null,
  condition: 'GOOD' as Condition,
  hours: null as number | null,
})
const result = ref<ReturnResult | null>(null)

function openReturn(r: Rental) {
  Object.assign(rt, { open: true, rental: r, condition: r.condition_at_handover, hours: null })
}

/** Same rule as the backend, for a preview before submitting. */
const preview = computed(() => {
  const r = rt.rental
  const t = r && typeOf(r.tool_type_id)
  if (!r || !t) return null
  const start = new Date(r.start_date)
  const due = new Date(r.expected_end_date)
  const now = new Date()
  const planned = Math.max(1, startedDays(start, due))
  const late = startedDays(due, now)
  return { base: t.daily_rate * planned, late: t.daily_rate * late, lateDays: late }
})

async function doReturn() {
  if (!rt.rental) return
  rt.busy = true
  try {
    result.value = await store.giveBack(rt.rental.id, rt.condition, rt.hours)
    rt.open = false
  } catch (e) {
    notify.fail(e, 'A visszavétel nem sikerült')
  } finally {
    rt.busy = false
  }
}

function cancelRental(r: Rental) {
  confirm.require({
    header: 'Törli a kölcsönzést?',
    message: `A(z) #${r.id} kölcsönzés (${r.tool_type_name}) használat nélkül jött vissza. Díjmentesen lezárjuk.`,
    icon: 'pi pi-exclamation-triangle',
    rejectProps: { label: 'Vissza', severity: 'secondary', outlined: true },
    acceptProps: { label: 'Kölcsönzés törlése', severity: 'danger' },
    accept: async () => {
      try {
        await store.cancel(r.id)
        notify.ok('Kölcsönzés törölve')
      } catch (e) {
        notify.fail(e, 'A törlés nem sikerült')
      }
    },
  })
}

// ---------- counter: reserve for a customer and hand over now ----------
const customers = ref<Customer[]>([])
const ct = reactive({
  open: false,
  busy: false,
  customerId: null as number | null,
  typeId: null as number | null,
  end: null as Date | null,
  handoverNow: true,
})

async function openCounter() {
  Object.assign(ct, { open: true, customerId: null, typeId: null, end: addDays(new Date(), 1), handoverNow: true })
  if (!customers.value.length) {
    try {
      customers.value = await customerApi.list()
    } catch (e) {
      notify.fail(e, 'Az ügyfelek betöltése nem sikerült')
    }
  }
}

async function doCounter() {
  if (!ct.customerId || !ct.typeId || !ct.end) return
  ct.busy = true
  try {
    const res = await reservationApi.create({
      tool_type_id: ct.typeId,
      customer_id: ct.customerId,
      start_date: new Date().toISOString(),
      end_date: ct.end.toISOString(),
    })
    if (ct.handoverNow) {
      const rental = await store.handover(res.id)
      notify.ok('Eszköz kiadva', `#${rental.tool_item_id} példány, kaució: ${money(rental.deposit_paid)}`)
    } else {
      await store.load()
      notify.ok('Lefoglalva', `#${res.tool_item_id} példány`)
    }
    ct.open = false
  } catch (e) {
    notify.fail(e, 'A művelet nem sikerült')
  } finally {
    ct.busy = false
  }
}
</script>

<template>
  <div class="page">
    <div class="page-head">
      <h1>Kölcsönzések</h1>
      <div class="row">
        <Button label="Frissítés" icon="pi pi-refresh" severity="secondary" outlined :loading="store.loading"
          @click="refresh" />
        <Button label="Pultnál kiadás" icon="pi pi-plus" @click="openCounter" />
      </div>
    </div>

    <Tabs value="out">
      <TabList>
        <Tab value="out">Kint lévő eszközök ({{ store.openRentals.length }})</Tab>
        <Tab value="upcoming">Kiadásra vár ({{ store.upcoming.length }})</Tab>
        <Tab value="history">Lezárt kölcsönzések</Tab>
      </TabList>
      <TabPanels>
        <TabPanel value="out">
          <DataTable :value="store.openRentals" :loading="store.loading" data-key="id" striped-rows
            :row-class="(r: Rental) => (isLate(r) ? 'late-row' : '')">
            <template #empty><div class="empty">Jelenleg nincs kint eszköz.</div></template>
            <Column field="id" header="#" />
            <Column field="customer_name" header="Ügyfél" />
            <Column header="Eszköz">
              <template #body="{ data }">{{ data.tool_type_name }} <span class="muted">#{{ data.tool_item_id }}</span></template>
            </Column>
            <Column header="Kiadva"><template #body="{ data }">{{ dateTime(data.start_date) }}</template></Column>
            <Column header="Határidő"><template #body="{ data }">{{ dateTime(data.expected_end_date) }}</template></Column>
            <Column header="Kaució" class="num"><template #body="{ data }">{{ money(data.deposit_paid) }}</template></Column>
            <Column header="Állapot"><template #body="{ data }"><StatusTag :value="data.status" /></template></Column>
            <Column header="" class="num">
              <template #body="{ data }">
                <div class="actions">
                  <Button label="Visszavétel" icon="pi pi-download" size="small" @click="openReturn(data)" />
                  <Button v-tooltip.top="'Használat nélkül jött vissza: díjmentes lezárás'" icon="pi pi-times" size="small"
                    severity="secondary" text aria-label="Kölcsönzés törlése" @click="cancelRental(data)" />
                </div>
              </template>
            </Column>
          </DataTable>
        </TabPanel>

        <TabPanel value="upcoming">
          <DataTable :value="store.upcoming" :loading="store.loading" data-key="id" striped-rows>
            <template #empty><div class="empty">Nincs kiadásra váró foglalás.</div></template>
            <Column field="id" header="#" />
            <Column field="customer_name" header="Ügyfél" />
            <Column header="Eszköz">
              <template #body="{ data }">{{ data.tool_type_name }} <span class="muted">#{{ data.tool_item_id }}</span></template>
            </Column>
            <Column header="Átvétel"><template #body="{ data }">{{ dateTime(data.start_date) }}</template></Column>
            <Column header="Visszahozatal"><template #body="{ data }">{{ dateTime(data.end_date) }}</template></Column>
            <Column header="Állapot"><template #body="{ data }"><StatusTag :value="data.status" /></template></Column>
            <Column header="" class="num">
              <template #body="{ data }">
                <Button label="Kiadás" icon="pi pi-upload" size="small" @click="openHandover(data)" />
              </template>
            </Column>
          </DataTable>
        </TabPanel>

        <TabPanel value="history">
          <DataTable :value="store.history" :loading="store.loading" data-key="id" striped-rows paginator :rows="15">
            <template #empty><div class="empty">Még nincs lezárt kölcsönzés.</div></template>
            <Column field="id" header="#" />
            <Column field="customer_name" header="Ügyfél" />
            <Column header="Eszköz">
              <template #body="{ data }">{{ data.tool_type_name }} <span class="muted">#{{ data.tool_item_id }}</span></template>
            </Column>
            <Column header="Visszahozva"><template #body="{ data }">{{ dateTime(data.actual_end_date) }}</template></Column>
            <Column header="Állapot visszavételkor"><template #body="{ data }"><StatusTag v-if="data.condition_at_return" :value="data.condition_at_return" /></template></Column>
            <Column header="Késedelmi díj" class="num"><template #body="{ data }">{{ money(data.late_fee) }}</template></Column>
            <Column header="Fizetendő" class="num"><template #body="{ data }">{{ money(data.total_fee) }}</template></Column>
            <Column header="Állapot"><template #body="{ data }"><StatusTag :value="data.status" /></template></Column>
          </DataTable>
        </TabPanel>
      </TabPanels>
    </Tabs>

    <!-- handover -->
    <Dialog v-model:visible="ho.open" modal header="Eszköz kiadása" :style="{ width: 'min(480px, 95vw)' }">
      <div v-if="ho.reservation" class="stack">
        <dl class="facts">
          <dt>Ügyfél</dt><dd>{{ ho.reservation.customer_name }}</dd>
          <dt>Eszköz</dt><dd>{{ ho.reservation.tool_type_name }} #{{ ho.reservation.tool_item_id }}</dd>
          <dt>Visszahozatal</dt><dd>{{ dateTime(ho.reservation.end_date) }}</dd>
          <dt>Példány állapota</dt><dd><StatusTag v-if="ho.item" :value="ho.item.status" /></dd>
        </dl>
        <Message v-if="handoverBlocked" severity="error" :closable="false">
          Ez a példány nem adható ki (állapota: {{ label(ho.item!.status).toLowerCase() }}).
        </Message>
        <Message v-if="ho.item?.maintenance?.due" severity="warn" :closable="false">
          Ennek a példánynak esedékes a karbantartása.
        </Message>
        <div class="field">
          <label for="ho-cond">Állapot kiadáskor</label>
          <Select input-id="ho-cond" v-model="ho.condition" :options="conditionOptions.filter((c) => c.value !== 'DAMAGED')"
            option-label="label" option-value="value" />
        </div>
        <div class="row">
          <Checkbox v-model="ho.depositReceived" binary input-id="ho-dep" />
          <label for="ho-dep">
            Kaució átvéve: <strong>{{ money(typeOf(ho.reservation.tool_type_id)?.deposit) }}</strong>
          </label>
        </div>
      </div>
      <template #footer>
        <Button label="Bezárás" severity="secondary" text @click="ho.open = false" />
        <Button label="Kiadás" icon="pi pi-check" :loading="ho.busy"
          :disabled="!ho.depositReceived || !ho.item || handoverBlocked" @click="doHandover" />
      </template>
    </Dialog>

    <!-- return -->
    <Dialog v-model:visible="rt.open" modal header="Visszavétel" :style="{ width: 'min(480px, 95vw)' }">
      <div v-if="rt.rental" class="stack">
        <dl class="facts">
          <dt>Ügyfél</dt><dd>{{ rt.rental.customer_name }}</dd>
          <dt>Eszköz</dt><dd>{{ rt.rental.tool_type_name }} #{{ rt.rental.tool_item_id }}</dd>
          <dt>Határidő</dt><dd>{{ dateTime(rt.rental.expected_end_date) }}</dd>
          <dt>Állapot kiadáskor</dt><dd><StatusTag :value="rt.rental.condition_at_handover" /></dd>
        </dl>
        <div class="form-grid">
          <div class="field">
            <label for="rt-cond">Állapot visszavételkor</label>
            <Select input-id="rt-cond" v-model="rt.condition" :options="conditionOptions" option-label="label"
              option-value="value" />
          </div>
          <div class="field">
            <label for="rt-hours">Üzemóra (nem kötelező)</label>
            <InputNumber input-id="rt-hours" v-model="rt.hours" :min="0" :max-fraction-digits="2" suffix=" óra" />
          </div>
        </div>
        <Message v-if="rt.condition === 'DAMAGED'" severity="warn" :closable="false">
          A sérült eszköz „Meghibásodott” állapotba kerül, és javításig nem kölcsönözhető.
        </Message>
        <div v-if="preview" class="summary">
          <span>Várható díj: {{ money(preview.base) }}</span>
          <span v-if="preview.lateDays > 0" class="bad">
            + késedelmi díj {{ money(preview.late) }} ({{ preview.lateDays }} nap)
          </span>
        </div>
      </div>
      <template #footer>
        <Button label="Bezárás" severity="secondary" text @click="rt.open = false" />
        <Button label="Visszavétel" icon="pi pi-check" :loading="rt.busy" @click="doReturn" />
      </template>
    </Dialog>

    <!-- return result -->
    <Dialog :visible="!!result" modal header="Kölcsönzés lezárva" :style="{ width: 'min(420px, 95vw)' }"
      @update:visible="result = null">
      <dl v-if="result" class="facts fees">
        <dt>Alapdíj</dt><dd class="num">{{ money(result.fees.base_fee) }}</dd>
        <dt>Késedelmi díj <span v-if="result.fees.late_days" class="muted">({{ result.fees.late_days }} nap)</span></dt>
        <dd class="num">{{ money(result.fees.late_fee) }}</dd>
        <dt class="total">Fizetendő összesen</dt><dd class="num total">{{ money(result.fees.total_fee) }}</dd>
        <dt>Befizetett kaució</dt><dd class="num">{{ money(result.rental.deposit_paid) }}</dd>
      </dl>
      <template #footer><Button label="Kész" @click="result = null" /></template>
    </Dialog>

    <!-- counter rental -->
    <Dialog v-model:visible="ct.open" modal header="Kiadás a pultnál" :style="{ width: 'min(520px, 95vw)' }">
      <div class="stack">
        <p class="muted hint">Foglaljon szerszámot az ügyfélnek mostantól, és ha kéri, azonnal adja ki.</p>
        <div class="field">
          <label for="ct-cust">Ügyfél</label>
          <Select input-id="ct-cust" v-model="ct.customerId" :options="customers" option-label="name" option-value="id"
            filter placeholder="Válasszon ügyfelet" />
        </div>
        <div class="field">
          <label for="ct-type">Szerszám</label>
          <Select input-id="ct-type" v-model="ct.typeId" :options="tools.types" option-label="name" option-value="id"
            filter placeholder="Válasszon szerszámot" />
        </div>
        <div class="field">
          <label for="ct-end">Visszahozatal határideje</label>
          <DatePicker input-id="ct-end" v-model="ct.end" show-time hour-format="24" :min-date="new Date()" show-icon />
        </div>
        <div class="row">
          <Checkbox v-model="ct.handoverNow" binary input-id="ct-now" />
          <label for="ct-now">Azonnali kiadás (kaució átvéve: {{ money(ct.typeId ? typeOf(ct.typeId)?.deposit : null) }})</label>
        </div>
      </div>
      <template #footer>
        <Button label="Bezárás" severity="secondary" text @click="ct.open = false" />
        <Button :label="ct.handoverNow ? 'Foglalás és kiadás' : 'Foglalás'" icon="pi pi-check" :loading="ct.busy"
          :disabled="!ct.customerId || !ct.typeId || !ct.end" @click="doCounter" />
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.actions {
  display: inline-flex;
  gap: 0.25rem;
}
:deep(.p-button) {
  white-space: nowrap;
}
:deep(.late-row) {
  background: color-mix(in srgb, var(--p-red-500) 8%, transparent) !important;
}
.facts {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 0.4rem 1rem;
  margin: 0;
}
.facts dt {
  color: var(--app-muted);
}
.facts dd {
  margin: 0;
}
.fees .total {
  font-weight: 700;
  font-size: 1.1rem;
  color: var(--p-text-color);
}
.summary {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  padding: 0.6rem 0.8rem;
  border-radius: 8px;
  background: var(--p-content-hover-background);
}
.bad {
  color: var(--p-red-400);
}
.hint {
  margin: 0;
}
</style>
