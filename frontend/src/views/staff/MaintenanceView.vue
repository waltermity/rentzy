<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import Button from 'primevue/button'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import Dialog from 'primevue/dialog'
import InputNumber from 'primevue/inputnumber'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Textarea from 'primevue/textarea'
import Tabs from 'primevue/tabs'
import TabList from 'primevue/tablist'
import Tab from 'primevue/tab'
import TabPanels from 'primevue/tabpanels'
import TabPanel from 'primevue/tabpanel'
import StartMaintenanceDialog from '@/components/StartMaintenanceDialog.vue'
import StatusTag from '@/components/StatusTag.vue'
import { maintenanceApi } from '@/api'
import { CONDITIONS, type Condition, type Maintenance, type MaintenanceDue, type ToolStatus } from '@/api/types'
import { useAuthStore } from '@/stores/auth'
import { date, dateTime, label, money } from '@/utils/format'
import { useNotify } from '@/utils/notify'

const auth = useAuthStore()
const notify = useNotify()

const due = ref<MaintenanceDue[]>([])
const open = ref<Maintenance[]>([])
const done = ref<Maintenance[]>([])
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const [d, o, c] = await Promise.all([
      maintenanceApi.due(),
      maintenanceApi.list({ open: true }),
      maintenanceApi.list({ open: false }),
    ])
    due.value = d
    open.value = o
    done.value = c
  } catch (e) {
    notify.fail(e, 'A karbantartások betöltése nem sikerült')
  } finally {
    loading.value = false
  }
}
onMounted(load)

const reasonText: Record<string, string> = {
  RENTAL_COUNT: 'Kölcsönzések',
  ELAPSED_DAYS: 'Idő',
  OPERATING_HOURS: 'Üzemóra',
}
const usage = (d: MaintenanceDue) => {
  const parts = [`${d.rentals_since}${d.maintenance_every_rentals ? ` / ${d.maintenance_every_rentals}` : ''} kölcsönzés`]
  parts.push(`${d.days_since}${d.maintenance_every_days ? ` / ${d.maintenance_every_days}` : ''} nap`)
  if (d.maintenance_every_hours) parts.push(`${d.operating_hours_since} / ${d.maintenance_every_hours} üzemóra`)
  return parts.join(' · ')
}

// ---------- start ----------
const startItem = ref<number | null>(null)

// ---------- complete ----------
const resultOptions: { label: string; value: ToolStatus }[] = [
  { label: 'Újra elérhető', value: 'AVAILABLE' },
  { label: 'Továbbra is meghibásodott', value: 'BROKEN' },
  { label: 'Kivonás a forgalomból', value: 'RETIRED' },
]
const cp = reactive({
  record: null as Maintenance | null,
  cost: 0,
  result: 'AVAILABLE' as ToolStatus,
  condition: 'GOOD' as Condition,
  description: '',
  performedBy: '',
  busy: false,
})

function openComplete(m: Maintenance) {
  Object.assign(cp, {
    record: m, cost: 0, result: 'AVAILABLE', condition: 'GOOD',
    description: m.description, performedBy: m.performed_by,
  })
}

async function complete() {
  if (!cp.record) return
  cp.busy = true
  try {
    await maintenanceApi.complete(cp.record.id, {
      cost: cp.cost ?? 0,
      result_status: cp.result,
      condition: cp.condition,
      description: cp.description,
      performed_by: cp.performedBy,
    })
    notify.ok('Karbantartás lezárva')
    cp.record = null
    await load()
  } catch (e) {
    notify.fail(e, 'A lezárás nem sikerült')
  } finally {
    cp.busy = false
  }
}
</script>

<template>
  <div class="page">
    <div class="page-head">
      <h1>Karbantartás</h1>
      <Button label="Frissítés" icon="pi pi-refresh" severity="secondary" outlined :loading="loading" @click="load" />
    </div>

    <Tabs value="due">
      <TabList>
        <Tab value="due">Esedékes ({{ due.length }})</Tab>
        <Tab value="open">Folyamatban ({{ open.length }})</Tab>
        <Tab value="done">Előzmények</Tab>
      </TabList>
      <TabPanels>
        <TabPanel value="due">
          <p class="muted hint">
            Azok a példányok, amelyek az utolsó lezárt karbantartásuk óta elérték az eszköztípusuk valamelyik
            karbantartási küszöbét (kölcsönzések száma, eltelt napok vagy üzemóra).
          </p>
          <DataTable :value="due" :loading="loading" data-key="tool_item_id" striped-rows>
            <template #empty><div class="empty">Nincs esedékes karbantartás.</div></template>
            <Column header="Példány"><template #body="{ data }">#{{ data.tool_item_id }}</template></Column>
            <Column field="tool_type_name" header="Eszköz" />
            <Column header="Állapot"><template #body="{ data }"><StatusTag :value="data.status" /></template></Column>
            <Column header="Ok">
              <template #body="{ data }">
                <div class="tags">
                  <Tag v-for="r in data.reasons" :key="r" :value="reasonText[r]" severity="warn" />
                </div>
              </template>
            </Column>
            <Column header="Utolsó karbantartás óta">
              <template #body="{ data }">
                <span class="muted">{{ usage(data) }}</span>
                <div class="small muted">{{ data.last_maintenance_at ? `Utoljára: ${date(data.last_maintenance_at)}` : `Még nem volt karbantartva (beszerzés: ${date(data.since)})` }}</div>
              </template>
            </Column>
            <Column v-if="auth.isAdmin" header="" class="num">
              <template #body="{ data }">
                <Button v-if="data.status !== 'MAINTENANCE' && data.status !== 'RENTED'" label="Indítás" icon="pi pi-wrench"
                  size="small" @click="startItem = data.tool_item_id" />
                <span v-else-if="data.status === 'RENTED'" class="small muted">Ki van kölcsönözve</span>
              </template>
            </Column>
          </DataTable>
        </TabPanel>

        <TabPanel value="open">
          <DataTable :value="open" :loading="loading" data-key="id" striped-rows>
            <template #empty><div class="empty">Nincs folyamatban lévő karbantartás.</div></template>
            <Column header="Példány"><template #body="{ data }">#{{ data.tool_item_id }}</template></Column>
            <Column field="tool_type_name" header="Eszköz" />
            <Column header="Típus"><template #body="{ data }">{{ label(data.type) }}</template></Column>
            <Column field="description" header="Leírás" />
            <Column header="Kezdete"><template #body="{ data }">{{ dateTime(data.started_at) }}</template></Column>
            <Column field="performed_by" header="Végzi" />
            <Column v-if="auth.isAdmin" header="" class="num">
              <template #body="{ data }">
                <Button label="Lezárás" icon="pi pi-check" size="small" @click="openComplete(data)" />
              </template>
            </Column>
          </DataTable>
        </TabPanel>

        <TabPanel value="done">
          <DataTable :value="done" :loading="loading" data-key="id" striped-rows paginator :rows="15">
            <template #empty><div class="empty">Még nincs lezárt karbantartás.</div></template>
            <Column header="Példány"><template #body="{ data }">#{{ data.tool_item_id }}</template></Column>
            <Column field="tool_type_name" header="Eszköz" />
            <Column header="Típus"><template #body="{ data }">{{ label(data.type) }}</template></Column>
            <Column field="description" header="Leírás" />
            <Column header="Lezárva"><template #body="{ data }">{{ dateTime(data.completed_at) }}</template></Column>
            <Column field="performed_by" header="Végzi" />
            <Column header="Költség" class="num"><template #body="{ data }">{{ money(data.cost) }}</template></Column>
          </DataTable>
        </TabPanel>
      </TabPanels>
    </Tabs>

    <StartMaintenanceDialog :item-id="startItem" @close="startItem = null" @started="load" />

    <Dialog :visible="!!cp.record" modal :header="`Karbantartás lezárása (#${cp.record?.id})`"
      :style="{ width: 'min(520px, 95vw)' }" @update:visible="cp.record = null">
      <div class="stack">
        <div class="form-grid">
          <div class="field">
            <label for="cp-result">Eredmény</label>
            <Select input-id="cp-result" v-model="cp.result" :options="resultOptions" option-label="label"
              option-value="value" />
          </div>
          <div class="field">
            <label for="cp-cond">Állapot</label>
            <Select input-id="cp-cond" v-model="cp.condition"
              :options="CONDITIONS.map((c) => ({ label: label(c), value: c }))" option-label="label"
              option-value="value" />
          </div>
          <div class="field">
            <label for="cp-cost">Költség</label>
            <InputNumber input-id="cp-cost" v-model="cp.cost" :min="0" mode="currency" currency="HUF" locale="hu-HU"
              :max-fraction-digits="0" />
          </div>
          <div class="field">
            <label for="cp-by">Végezte</label>
            <InputText id="cp-by" v-model="cp.performedBy" />
          </div>
        </div>
        <div class="field">
          <label for="cp-desc">Megjegyzés</label>
          <Textarea id="cp-desc" v-model="cp.description" rows="3" auto-resize />
        </div>
      </div>
      <template #footer>
        <Button label="Bezárás" severity="secondary" text @click="cp.record = null" />
        <Button label="Lezárás" icon="pi pi-check" :loading="cp.busy" @click="complete" />
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.hint {
  margin: 0.25rem 0 0.75rem;
  font-size: 0.9rem;
}
.tags {
  display: flex;
  gap: 0.3rem;
  flex-wrap: wrap;
}
.small {
  font-size: 0.8rem;
}
</style>
