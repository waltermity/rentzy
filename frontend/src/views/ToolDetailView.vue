<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import Button from 'primevue/button'
import Card from 'primevue/card'
import DatePicker from 'primevue/datepicker'
import Select from 'primevue/select'
import Message from 'primevue/message'
import Skeleton from 'primevue/skeleton'
import Tag from 'primevue/tag'
import { customerApi, reservationApi, toolApi } from '@/api'
import { errorMessage } from '@/api/client'
import type { Availability, Customer, Reservation, ToolType } from '@/api/types'
import { useAuthStore } from '@/stores/auth'
import { useReservationStore } from '@/stores/reservations'
import { useToolStore } from '@/stores/tools'
import { addDays, date, dateTime, money, startOfDay, startedDays } from '@/utils/format'
import { useNotify } from '@/utils/notify'

const HORIZON_DAYS = 90

const route = useRoute()
const auth = useAuthStore()
const tools = useToolStore()
const reservations = useReservationStore()
const notify = useNotify()

const id = computed(() => Number(route.params.id))
const tool = ref<ToolType | null>(null)
const calendar = ref<Availability | null>(null)
const loadError = ref('')

const today = startOfDay(new Date())
const horizon = addDays(today, HORIZON_DAYS)

async function load() {
  loadError.value = ''
  try {
    const [t, a] = await Promise.all([
      tools.getType(id.value, true),
      toolApi.availability(id.value, today.toISOString(), horizon.toISOString()),
    ])
    tool.value = t
    calendar.value = a
  } catch (e) {
    loadError.value = errorMessage(e)
  }
}
onMounted(load)
watch(id, load)

// ---------- calendar ----------

/** True if the item's busy intervals cover all of [from, to). */
function fullyBusy(item: number, from: Date, to: Date, busy: Availability['busy']): boolean {
  const spans = busy
    .filter((b) => b.tool_item_id === item)
    .map((b) => [new Date(b.start).getTime(), new Date(b.end).getTime()] as const)
    .filter(([s, e]) => s < to.getTime() && e > from.getTime())
    .sort((x, y) => x[0] - y[0])
  let covered = from.getTime()
  for (const [s, e] of spans) {
    if (s > covered) return false
    covered = Math.max(covered, e)
    if (covered >= to.getTime()) return true
  }
  return covered >= to.getTime()
}

/** Days on which every bookable copy is busy all day. Partly booked days stay selectable. */
const disabledDates = computed<Date[]>(() => {
  const a = calendar.value
  if (!a) return []
  const out: Date[] = []
  for (let d = today; d < horizon; d = addDays(d, 1)) {
    const end = addDays(d, 1)
    if (a.bookable_item_ids.every((item) => fullyBusy(item, d, end, a.busy))) out.push(d)
  }
  return out
})

const noCopies = computed(() => calendar.value != null && calendar.value.bookable_item_ids.length === 0)

const range = ref<(Date | null)[] | null>(null)
const hours = Array.from({ length: 12 }, (_, i) => {
  const h = i + 7
  return { label: `${String(h).padStart(2, '0')}:00`, value: h }
})
const hour = ref(8)

const period = computed(() => {
  const [s, e] = range.value ?? []
  if (!s || !e) return null
  const start = new Date(s)
  start.setHours(hour.value, 0, 0, 0)
  const end = new Date(e)
  end.setHours(hour.value, 0, 0, 0)
  return { start, end, days: startedDays(start, end) }
})

const periodProblem = computed(() => {
  const p = period.value
  if (!p) return ''
  if (p.days < 1) return 'Két különböző napot válasszon: az átvétel és a visszahozatal napját.'
  if (p.start.getTime() < Date.now()) return 'Az átvétel időpontja már elmúlt. Válasszon későbbi időpontot vagy napot.'
  return ''
})

const estimate = computed(() =>
  tool.value && period.value && period.value.days > 0 ? tool.value.daily_rate * period.value.days : null,
)

// Exact check for the chosen period: how many copies are free for all of it.
const freeCount = ref<number | null>(null)
const checking = ref(false)
watch(period, async (p) => {
  freeCount.value = null
  if (!p || periodProblem.value) return
  checking.value = true
  try {
    const a = await toolApi.availability(id.value, p.start.toISOString(), p.end.toISOString())
    freeCount.value = a.available_item_ids.length
  } catch {
    freeCount.value = null
  } finally {
    checking.value = false
  }
})

// ---------- staff: reserve for a customer ----------

const customers = ref<Customer[]>([])
const customerId = ref<number | null>(null)
onMounted(async () => {
  if (auth.isStaff) {
    try {
      customers.value = await customerApi.list()
    } catch (e) {
      notify.fail(e, 'Az ügyfelek betöltése nem sikerült')
    }
  }
})

// ---------- reserve ----------

const busy = ref(false)
const created = ref<Reservation | null>(null)

const canSubmit = computed(
  () =>
    !!period.value &&
    !periodProblem.value &&
    freeCount.value !== 0 &&
    (!auth.isStaff || customerId.value != null) &&
    (auth.isCustomer || auth.isStaff),
)

async function reserve() {
  const p = period.value
  if (!p || !tool.value) return
  busy.value = true
  try {
    if (auth.isStaff) {
      created.value = await reservationApi.create({
        tool_type_id: tool.value.id,
        customer_id: customerId.value!,
        start_date: p.start.toISOString(),
        end_date: p.end.toISOString(),
      })
    } else {
      created.value = await reservations.create(tool.value.id, p.start, p.end)
    }
    notify.ok('Foglalás visszaigazolva', `${tool.value.name}, #${created.value.tool_item_id} példány`)
    range.value = null
    await load()
  } catch (e) {
    notify.fail(e, 'A foglalás nem sikerült')
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div class="page">
    <RouterLink to="/catalog" class="back"><i class="pi pi-arrow-left" /> Katalógus</RouterLink>

    <Message v-if="loadError" severity="error" :closable="false">{{ loadError }}</Message>
    <Skeleton v-else-if="!tool" height="320px" border-radius="12px" />

    <div v-else class="layout">
      <Card class="info">
        <template #title>{{ tool.name }}</template>
        <template #subtitle>{{ [tool.manufacturer, tool.model].filter(Boolean).join(' · ') }}</template>
        <template #content>
          <Tag :value="tool.category_name" severity="secondary" />
          <dl class="facts">
            <dt>Napi díj</dt>
            <dd class="big">{{ money(tool.daily_rate) }}</dd>
            <dt>Kaució</dt>
            <dd>{{ money(tool.deposit) }}</dd>
            <dt>Most elérhető példány</dt>
            <dd>{{ tool.item_counts?.AVAILABLE ?? 0 }}</dd>
          </dl>
          <p class="muted small">
            Minden megkezdett nap díjköteles. Késedelmes visszahozatalnál minden megkezdett késett nap
            után a napi díjat számoljuk fel késedelmi díjként. A kauciót átvételkor kell megfizetni.
          </p>
        </template>
      </Card>

      <Card class="booking">
        <template #title>Foglalás</template>
        <template #content>
          <Message v-if="noCopies" severity="warn" :closable="false">
            Ebből a szerszámból jelenleg egy példány sem foglalható.
          </Message>
          <div v-else class="stack">
            <p class="muted small hint">
              Válassza ki az átvétel és a visszahozatal napját. A szürke napok teljesen foglaltak.
            </p>
            <DatePicker v-model="range" selection-mode="range" inline :min-date="today" :max-date="horizon"
              :disabled-dates="disabledDates" :manual-input="false" class="cal" />

            <div class="row">
              <div class="field">
                <label for="hour">Átvétel és visszahozatal ideje</label>
                <Select input-id="hour" v-model="hour" :options="hours" option-label="label" option-value="value" />
              </div>
              <div v-if="auth.isStaff" class="field grow">
                <label for="customer">Ügyfél</label>
                <Select input-id="customer" v-model="customerId" :options="customers" option-label="name"
                  option-value="id" filter placeholder="Válasszon ügyfelet" />
              </div>
            </div>

            <Message v-if="periodProblem" severity="warn" :closable="false">{{ periodProblem }}</Message>

            <div v-else-if="period" class="summary">
              <div>
                <strong>{{ dateTime(period.start.toISOString()) }}</strong> →
                <strong>{{ dateTime(period.end.toISOString()) }}</strong>
              </div>
              <div>{{ period.days }} nap · várható díj
                <strong>{{ money(estimate) }}</strong> + kaució {{ money(tool.deposit) }}</div>
              <div v-if="checking" class="muted"><i class="pi pi-spin pi-spinner" /> Elérhetőség ellenőrzése…</div>
              <div v-else-if="freeCount === 0" class="bad">
                <i class="pi pi-times-circle" /> Egy példány sem szabad a teljes időszakra. Próbáljon más napokat.
              </div>
              <div v-else-if="freeCount != null" class="good">
                <i class="pi pi-check-circle" /> {{ freeCount }} példány szabad erre az időszakra
              </div>
            </div>

            <Button v-if="!auth.isLoggedIn" as="router-link"
              :to="{ path: '/login', query: { redirect: route.fullPath } }" label="Jelentkezzen be a foglaláshoz"
              icon="pi pi-sign-in" />
            <Button v-else :label="auth.isStaff ? 'Foglalás az ügyfélnek' : 'Foglalás'" icon="pi pi-calendar-plus"
              :disabled="!canSubmit" :loading="busy" @click="reserve" />

            <Message v-if="created" severity="success" :closable="false">
              Lefoglalva: #{{ created.tool_item_id }} példány,
              {{ dateTime(created.start_date) }} – {{ dateTime(created.end_date) }}.
              <RouterLink v-if="auth.isCustomer" to="/profile">Foglalásaim megtekintése</RouterLink>
              <RouterLink v-else to="/staff/rentals">Ugrás a kölcsönzésekhez</RouterLink>
            </Message>
          </div>
        </template>
      </Card>
    </div>

    <p v-if="calendar" class="muted small foot">Az elérhetőség eddig a napig látható: {{ date(horizon.toISOString()) }}</p>
  </div>
</template>

<style scoped>
.back {
  display: inline-flex;
  gap: 0.4rem;
  align-items: center;
  margin-bottom: 1rem;
  text-decoration: none;
}
.layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1.3fr);
  gap: 1rem;
  align-items: start;
}
@media (max-width: 860px) {
  .layout {
    grid-template-columns: 1fr;
  }
}
.facts {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 0.5rem 1rem;
  margin: 1rem 0;
}
.facts dt {
  color: var(--app-muted);
}
.facts dd {
  margin: 0;
  font-weight: 600;
}
.facts .big {
  font-size: 1.3rem;
}
.small {
  font-size: 0.85rem;
}
.hint {
  margin: 0;
}
.cal {
  width: 100%;
}
.cal :deep(.p-datepicker-panel) {
  width: 100%;
}
.grow {
  flex: 1;
  min-width: 14rem;
}
.summary {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  padding: 0.75rem 1rem;
  border-radius: 10px;
  background: var(--p-content-hover-background);
}
.good {
  color: var(--p-green-500);
}
.bad {
  color: var(--p-red-400);
}
.foot {
  margin-top: 1rem;
}
</style>
