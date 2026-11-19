<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import Button from 'primevue/button'
import Card from 'primevue/card'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import DatePicker from 'primevue/datepicker'
import InputNumber from 'primevue/inputnumber'
import Message from 'primevue/message'
import ProgressBar from 'primevue/progressbar'
import SelectButton from 'primevue/selectbutton'
import Tabs from 'primevue/tabs'
import TabList from 'primevue/tablist'
import Tab from 'primevue/tab'
import TabPanels from 'primevue/tabpanels'
import TabPanel from 'primevue/tabpanel'
import StatusTag from '@/components/StatusTag.vue'
import { reportApi } from '@/api'
import type { PopularReport, ProcurementReport, UtilizationReport } from '@/api/types'
import { useAuthStore } from '@/stores/auth'
import { addDays, date, isoDay, money, percent, startOfDay } from '@/utils/format'
import { useNotify } from '@/utils/notify'

const auth = useAuthStore()
const notify = useNotify()

// ---------- window ----------
const presets = [
  { label: '30 nap', value: 30 },
  { label: '90 nap', value: 90 },
  { label: '1 év', value: 365 },
  { label: 'Egyéni', value: 0 },
]
const preset = ref(90)
const custom = ref<(Date | null)[] | null>(null)

const period = computed(() => {
  if (preset.value === 0) {
    const [f, t] = custom.value ?? []
    if (!f || !t) return null
    return { from: isoDay(f), to: isoDay(addDays(t, 1)) }
  }
  const to = addDays(startOfDay(new Date()), 1)
  return { from: isoDay(addDays(to, -preset.value)), to: isoDay(to) }
})

// ---------- data ----------
const tab = ref('popular')
const threshold = ref(70)
const popular = ref<PopularReport | null>(null)
const util = ref<UtilizationReport | null>(null)
const proc = ref<ProcurementReport | null>(null)
const loading = ref(false)

async function load() {
  const w = period.value
  if (!w) return
  loading.value = true
  try {
    const tasks: Promise<unknown>[] = [reportApi.popular({ ...w, limit: 50 }).then((r) => (popular.value = r))]
    if (auth.isAdmin) {
      tasks.push(reportApi.utilization(w).then((r) => (util.value = r)))
      tasks.push(reportApi.procurement({ ...w, threshold: threshold.value / 100 }).then((r) => (proc.value = r)))
    }
    await Promise.all(tasks)
  } catch (e) {
    notify.fail(e, 'A kimutatások betöltése nem sikerült')
  } finally {
    loading.value = false
  }
}
onMounted(load)
watch(period, load)

const maxRentals = computed(() => Math.max(1, ...(popular.value?.types.map((t) => t.rental_count) ?? [1])))
const kpis = computed(() => {
  const u = util.value
  if (!u) return null
  const rented = u.types.reduce((s, t) => s + t.rented_hours, 0)
  const rentals = u.types.reduce((s, t) => s + t.rental_count, 0)
  const avg = u.types.length ? u.types.reduce((s, t) => s + t.utilization * t.item_count, 0) / Math.max(1, u.types.reduce((s, t) => s + t.item_count, 0)) : 0
  return { revenue: u.total_revenue, rentals, rentedDays: Math.round(rented / 24), avg }
})
</script>

<template>
  <div class="page">
    <div class="page-head">
      <h1>Kimutatások</h1>
      <div class="row">
        <SelectButton v-model="preset" :options="presets" option-label="label" option-value="value" :allow-empty="false"
          aria-label="Időszak" />
        <DatePicker v-if="preset === 0" v-model="custom" selection-mode="range" :max-date="new Date()" show-icon
          placeholder="Kezdete – vége" :manual-input="false" />
        <Button icon="pi pi-refresh" severity="secondary" outlined aria-label="Frissítés" :loading="loading" @click="load" />
      </div>
    </div>
    <p v-if="popular" class="muted window">
      {{ date(popular.window.from) }} – {{ date(addDays(new Date(popular.window.to), -1).toISOString()) }}
    </p>

    <div v-if="kpis" class="kpis">
      <Card><template #content><div class="kpi-label">Bevétel</div><div class="kpi">{{ money(kpis.revenue) }}</div></template></Card>
      <Card><template #content><div class="kpi-label">Kölcsönzések száma</div><div class="kpi">{{ kpis.rentals }}</div></template></Card>
      <Card><template #content><div class="kpi-label">Kölcsönzési napok</div><div class="kpi">{{ kpis.rentedDays }}</div></template></Card>
      <Card><template #content><div class="kpi-label">Átlagos kihasználtság</div><div class="kpi">{{ percent(kpis.avg) }}</div></template></Card>
    </div>

    <Tabs v-model:value="tab">
      <TabList>
        <Tab value="popular">Leggyakrabban kölcsönzött</Tab>
        <Tab v-if="auth.isAdmin" value="util">Kihasználtság és bevétel</Tab>
        <Tab v-if="auth.isAdmin" value="proc">Beszerzési javaslatok</Tab>
      </TabList>
      <TabPanels>
        <TabPanel value="popular">
          <DataTable :value="popular?.types ?? []" :loading="loading" data-key="tool_type_id" striped-rows>
            <template #empty><div class="empty">Ebben az időszakban nem volt kölcsönzés.</div></template>
            <Column header="#"><template #body="{ index }">{{ index + 1 }}</template></Column>
            <Column field="tool_type_name" header="Eszköz" />
            <Column field="category_name" header="Kategória" />
            <Column header="Kölcsönzések" class="bar-col">
              <template #body="{ data }">
                <div class="bar-cell">
                  <div class="bar" :style="{ width: `${(data.rental_count / maxRentals) * 100}%` }" />
                  <span class="num">{{ data.rental_count }}</span>
                </div>
              </template>
            </Column>
            <Column header="Példányok" field="item_count" class="num" />
          </DataTable>
        </TabPanel>

        <TabPanel v-if="auth.isAdmin" value="util">
          <h3 class="sub">Eszköztípusonként</h3>
          <DataTable :value="util?.types ?? []" :loading="loading" data-key="tool_type_id" striped-rows
            sort-field="utilization" :sort-order="-1">
            <Column field="tool_type_name" header="Eszköz" sortable />
            <Column field="item_count" header="Példányok" class="num" />
            <Column field="rental_count" header="Kölcsönzések" class="num" sortable />
            <Column header="Kihasználtság" field="utilization" sortable class="bar-col">
              <template #body="{ data }">
                <div class="util">
                  <ProgressBar :value="Math.round(data.utilization * 100)" :show-value="false" class="pb" />
                  <span class="num">{{ percent(data.utilization) }}</span>
                </div>
              </template>
            </Column>
            <Column header="Bevétel" field="revenue" sortable class="num"><template #body="{ data }">{{ money(data.revenue) }}</template></Column>
          </DataTable>

          <h3 class="sub">Példányonként</h3>
          <DataTable :value="util?.items ?? []" :loading="loading" data-key="tool_item_id" striped-rows paginator :rows="15"
            sort-field="utilization" :sort-order="-1">
            <Column header="Példány" field="tool_item_id" sortable><template #body="{ data }">#{{ data.tool_item_id }}</template></Column>
            <Column field="tool_type_name" header="Eszköz" sortable />
            <Column header="Állapot"><template #body="{ data }"><StatusTag :value="data.status" /></template></Column>
            <Column field="rental_count" header="Kölcsönzések" class="num" sortable />
            <Column header="Kölcsönzött órák" field="rented_hours" class="num" sortable />
            <Column header="Kihasználtság" field="utilization" sortable class="bar-col">
              <template #body="{ data }">
                <div class="util">
                  <ProgressBar :value="Math.round(data.utilization * 100)" :show-value="false" class="pb" />
                  <span class="num">{{ percent(data.utilization) }}</span>
                </div>
              </template>
            </Column>
            <Column header="Bevétel" field="revenue" sortable class="num"><template #body="{ data }">{{ money(data.revenue) }}</template></Column>
          </DataTable>
        </TabPanel>

        <TabPanel v-if="auth.isAdmin" value="proc">
          <div class="row thr">
            <label for="thr">Beszerzési javaslat, ha a kihasználtság legalább</label>
            <InputNumber input-id="thr" v-model="threshold" :min="1" :max="100" suffix="%" show-buttons class="thr-in" />
            <Button label="Alkalmaz" severity="secondary" outlined :loading="loading" @click="load" />
          </div>
          <Message v-if="proc && !proc.suggestions.length" severity="success" :closable="false">
            Ebben az időszakban egy eszköz sem éri el a(z) {{ threshold }}%-os kihasználtságot, nincs szükség beszerzésre.
          </Message>
          <div class="sugg">
            <Card v-for="s in proc?.suggestions ?? []" :key="s.tool_type_id" class="sugg-card">
              <template #title>{{ s.tool_type_name }}</template>
              <template #subtitle>{{ s.category_name }}</template>
              <template #content>
                <div class="buy">Javasolt: <strong>+{{ s.suggested_additional }}</strong> példány</div>
                <div class="util">
                  <ProgressBar :value="Math.round(s.utilization * 100)" :show-value="false" class="pb" />
                  <span class="num">{{ percent(s.utilization) }}</span>
                </div>
                <dl class="facts">
                  <dt>Jelenlegi példányok</dt><dd>{{ s.item_count }}</dd>
                  <dt>Kölcsönzések</dt><dd>{{ s.rental_count }}</dd>
                  <dt>Bevétel</dt><dd>{{ money(s.revenue) }}</dd>
                  <dt>Ár / példány</dt><dd>{{ money(s.purchase_price) }}</dd>
                  <dt>Megtérülés</dt><dd>{{ s.payback_days != null ? `kb. ${s.payback_days} nap` : '—' }}</dd>
                </dl>
              </template>
            </Card>
          </div>
        </TabPanel>
      </TabPanels>
    </Tabs>
  </div>
</template>

<style scoped>
.window {
  margin: -0.75rem 0 1rem;
}
.kpis {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 1rem;
  margin-bottom: 1rem;
}
.kpi-label {
  color: var(--app-muted);
  font-size: 0.85rem;
}
.kpi {
  font-size: 1.6rem;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}
.bar-col {
  min-width: 12rem;
}
.bar-cell {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.bar {
  height: 10px;
  border-radius: 4px;
  background: var(--p-primary-color);
  min-width: 2px;
}
.util {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.pb {
  flex: 1;
  height: 8px;
  min-width: 6rem;
}
.sub {
  margin: 1rem 0 0.5rem;
  font-size: 1rem;
}
.thr {
  margin-bottom: 1rem;
}
.thr-in {
  width: 9rem;
}
.sugg {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 1rem;
}
.sugg-card {
  border: 1px solid var(--app-border);
}
.buy {
  font-size: 1.2rem;
  margin-bottom: 0.5rem;
}
.facts {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 0.3rem 1rem;
  margin: 0.75rem 0 0;
  font-size: 0.9rem;
}
.facts dt {
  color: var(--app-muted);
}
.facts dd {
  margin: 0;
  text-align: right;
}
</style>
