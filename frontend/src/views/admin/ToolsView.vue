<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import Button from 'primevue/button'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import DatePicker from 'primevue/datepicker'
import Dialog from 'primevue/dialog'
import InputNumber from 'primevue/inputnumber'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Tabs from 'primevue/tabs'
import TabList from 'primevue/tablist'
import Tab from 'primevue/tab'
import TabPanels from 'primevue/tabpanels'
import TabPanel from 'primevue/tabpanel'
import StartMaintenanceDialog from '@/components/StartMaintenanceDialog.vue'
import StatusTag from '@/components/StatusTag.vue'
import { maintenanceApi, toolApi, type ToolTypeInput } from '@/api'
import { CONDITIONS, type Category, type Condition, type Maintenance, type ToolItem, type ToolStatus, type ToolType } from '@/api/types'
import { useToolStore } from '@/stores/tools'
import { date, dateTime, label, money } from '@/utils/format'
import { useNotify } from '@/utils/notify'

const tools = useToolStore()
const notify = useNotify()
const tab = ref('types')

const items = ref<ToolItem[]>([])
const itemsLoading = ref(false)
const itemTypeFilter = ref<number | null>(null)
const itemStatusFilter = ref<ToolStatus | null>(null)

async function loadTypes() {
  try {
    await tools.loadCatalog(true)
  } catch (e) {
    notify.fail(e, 'Az eszközök betöltése nem sikerült')
  }
}
async function loadItems() {
  itemsLoading.value = true
  try {
    items.value = await toolApi.items()
  } catch (e) {
    notify.fail(e, 'A példányok betöltése nem sikerült')
  } finally {
    itemsLoading.value = false
  }
}
const reloadAll = () => Promise.all([loadTypes(), loadItems()])
onMounted(reloadAll)

const conditionOptions = CONDITIONS.map((c) => ({ label: label(c), value: c }))
const statusFilterOptions = (['AVAILABLE', 'RENTED', 'MAINTENANCE', 'BROKEN', 'RETIRED'] as ToolStatus[]).map((s) => ({
  label: label(s),
  value: s,
}))
const visibleItems = computed(() =>
  items.value.filter(
    (i) =>
      (itemTypeFilter.value == null || i.tool_type_id === itemTypeFilter.value) &&
      (itemStatusFilter.value == null || i.status === itemStatusFilter.value),
  ),
)
const count = (t: ToolType, s: ToolStatus) => t.item_counts?.[s] ?? 0

// ---------- tool type create / edit ----------
const emptyType = (): ToolTypeInput => ({
  name: '', category_id: tools.categories[0]?.id ?? 0, manufacturer: '', model: '',
  purchase_price: 0, daily_rate: 0, deposit: 0,
  maintenance_every_rentals: null, maintenance_every_days: null, maintenance_every_hours: null,
})
const te = reactive({ open: false, id: null as number | null, form: emptyType(), busy: false })

function openType(t?: ToolType) {
  te.id = t?.id ?? null
  te.form = t
    ? {
        name: t.name, category_id: t.category_id, manufacturer: t.manufacturer, model: t.model,
        purchase_price: t.purchase_price, daily_rate: t.daily_rate, deposit: t.deposit,
        maintenance_every_rentals: t.maintenance_every_rentals,
        maintenance_every_days: t.maintenance_every_days,
        maintenance_every_hours: t.maintenance_every_hours,
      }
    : emptyType()
  te.open = true
}

async function saveType() {
  te.busy = true
  try {
    if (te.id) await toolApi.updateType(te.id, te.form)
    else await toolApi.createType(te.form)
    notify.ok(te.id ? 'Eszköz módosítva' : 'Eszköz létrehozva', te.id ? undefined : 'Most adjon hozzá példányokat.')
    te.open = false
    await loadTypes()
  } catch (e) {
    notify.fail(e, 'A mentés nem sikerült')
  } finally {
    te.busy = false
  }
}

// ---------- add copies ----------
const ad = reactive({ type: null as ToolType | null, count: 1, condition: 'NEW' as Condition, acquired: new Date(), busy: false })
function openAdd(t: ToolType) {
  Object.assign(ad, { type: t, count: 1, condition: 'NEW', acquired: new Date() })
}
async function addCopies() {
  if (!ad.type) return
  ad.busy = true
  try {
    const created = await toolApi.addItems(ad.type.id, {
      count: ad.count, condition: ad.condition, acquired_at: ad.acquired.toISOString(),
    })
    notify.ok(`${created.length} példány hozzáadva`, created.map((i) => `#${i.id}`).join(', '))
    ad.type = null
    await Promise.all([loadTypes(), loadItems()])
  } catch (e) {
    notify.fail(e, 'A példányok hozzáadása nem sikerült')
  } finally {
    ad.busy = false
  }
}

// ---------- copy edit ----------
const ie = reactive({ item: null as ToolItem | null, status: 'AVAILABLE' as ToolStatus, condition: 'GOOD' as Condition, busy: false, history: [] as Maintenance[] })
async function openItem(i: ToolItem) {
  Object.assign(ie, { item: i, status: i.status, condition: i.current_condition, history: [] })
  try {
    ie.history = await maintenanceApi.history(i.id)
  } catch {
    /* history is optional */
  }
}
const itemStatusOptions = computed(() => {
  const s = ie.item?.status
  if (s === 'RENTED') return []
  if (s === 'MAINTENANCE') return [{ label: label('MAINTENANCE'), value: 'MAINTENANCE' }, { label: label('RETIRED'), value: 'RETIRED' }]
  return (['AVAILABLE', 'BROKEN', 'RETIRED'] as ToolStatus[]).map((v) => ({ label: label(v), value: v }))
})
async function saveItem() {
  if (!ie.item) return
  ie.busy = true
  try {
    await toolApi.updateItem(ie.item.id, {
      status: ie.status !== ie.item.status ? ie.status : undefined,
      condition: ie.condition !== ie.item.current_condition ? ie.condition : undefined,
    })
    notify.ok('Példány módosítva')
    ie.item = null
    await Promise.all([loadTypes(), loadItems()])
  } catch (e) {
    notify.fail(e, 'A módosítás nem sikerült')
  } finally {
    ie.busy = false
  }
}
const startItem = ref<number | null>(null)
function startFromItem() {
  startItem.value = ie.item?.id ?? null
  ie.item = null
}

// ---------- categories ----------
const ce = reactive({ open: false, id: null as number | null, name: '', busy: false })
function openCategory(c?: Category) {
  Object.assign(ce, { open: true, id: c?.id ?? null, name: c?.name ?? '' })
}
async function saveCategory() {
  ce.busy = true
  try {
    if (ce.id) await toolApi.updateCategory(ce.id, ce.name)
    else await toolApi.createCategory(ce.name)
    notify.ok('Kategória mentve')
    ce.open = false
    await loadTypes()
  } catch (e) {
    notify.fail(e, 'A mentés nem sikerült')
  } finally {
    ce.busy = false
  }
}
</script>

<template>
  <div class="page">
    <div class="page-head">
      <h1>Eszközök</h1>
      <div class="row">
        <Button v-if="tab === 'types'" label="Új eszköz" icon="pi pi-plus" @click="openType()" />
        <Button v-if="tab === 'categories'" label="Új kategória" icon="pi pi-plus" @click="openCategory()" />
      </div>
    </div>

    <Tabs v-model:value="tab">
      <TabList>
        <Tab value="types">Eszköztípusok ({{ tools.types.length }})</Tab>
        <Tab value="items">Példányok ({{ items.length }})</Tab>
        <Tab value="categories">Kategóriák ({{ tools.categories.length }})</Tab>
      </TabList>
      <TabPanels>
        <TabPanel value="types">
          <DataTable :value="tools.types" :loading="tools.loading" data-key="id" striped-rows sort-field="name" :sort-order="1">
            <Column field="name" header="Megnevezés" sortable />
            <Column field="category_name" header="Kategória" sortable />
            <Column header="Gyártó / típus"><template #body="{ data }">{{ [data.manufacturer, data.model].filter(Boolean).join(' ') }}</template></Column>
            <Column header="Napi díj" class="num" field="daily_rate" sortable><template #body="{ data }">{{ money(data.daily_rate) }}</template></Column>
            <Column header="Kaució" class="num"><template #body="{ data }">{{ money(data.deposit) }}</template></Column>
            <Column header="Beszerzési ár" class="num"><template #body="{ data }">{{ money(data.purchase_price) }}</template></Column>
            <Column header="Példányok">
              <template #body="{ data }">
                <span v-tooltip.top="'Elérhető / kikölcsönözve / karbantartás alatt / meghibásodott'" class="counts">
                  {{ count(data, 'AVAILABLE') }} · {{ count(data, 'RENTED') }} · {{ count(data, 'MAINTENANCE') }} · {{ count(data, 'BROKEN') }}
                </span>
              </template>
            </Column>
            <Column header="" class="num">
              <template #body="{ data }">
                <div class="actions">
                  <Button icon="pi pi-plus" text size="small" v-tooltip.top="'Példányok hozzáadása'" aria-label="Példányok hozzáadása" @click="openAdd(data)" />
                  <Button icon="pi pi-pencil" text size="small" v-tooltip.top="'Szerkesztés'" aria-label="Szerkesztés" @click="openType(data)" />
                </div>
              </template>
            </Column>
          </DataTable>
        </TabPanel>

        <TabPanel value="items">
          <div class="row filters">
            <Select v-model="itemTypeFilter" :options="tools.types" option-label="name" option-value="id" filter
              placeholder="Minden eszköz" show-clear aria-label="Eszköz" />
            <Select v-model="itemStatusFilter" :options="statusFilterOptions" option-label="label" option-value="value"
              placeholder="Minden állapot" show-clear aria-label="Állapot" />
          </div>
          <DataTable :value="visibleItems" :loading="itemsLoading" data-key="id" striped-rows paginator :rows="20"
            selection-mode="single" class="clickable" @row-click="openItem($event.data)">
            <Column field="id" header="Példány #" sortable />
            <Column field="tool_type_name" header="Eszköz" sortable />
            <Column header="Állapot" field="status" sortable><template #body="{ data }"><StatusTag :value="data.status" /></template></Column>
            <Column header="Fizikai állapot"><template #body="{ data }"><StatusTag :value="data.current_condition" /></template></Column>
            <Column header="Beszerezve"><template #body="{ data }">{{ date(data.acquired_at) }}</template></Column>
          </DataTable>
        </TabPanel>

        <TabPanel value="categories">
          <DataTable :value="tools.categories" data-key="id" striped-rows>
            <Column field="name" header="Megnevezés" />
            <Column header="Eszköztípusok"><template #body="{ data }">{{ tools.types.filter((t) => t.category_id === data.id).length }}</template></Column>
            <Column header="" class="num">
              <template #body="{ data }">
                <Button icon="pi pi-pencil" text size="small" aria-label="Átnevezés" @click="openCategory(data)" />
              </template>
            </Column>
          </DataTable>
        </TabPanel>
      </TabPanels>
    </Tabs>

    <!-- tool type -->
    <Dialog v-model:visible="te.open" modal :header="te.id ? 'Eszköz szerkesztése' : 'Új eszköz'" :style="{ width: 'min(680px, 95vw)' }">
      <form id="type-form" class="stack" @submit.prevent="saveType">
        <div class="form-grid">
          <div class="field"><label for="t-name">Megnevezés</label><InputText id="t-name" v-model="te.form.name" required /></div>
          <div class="field">
            <label for="t-cat">Kategória</label>
            <Select input-id="t-cat" v-model="te.form.category_id" :options="tools.categories" option-label="name" option-value="id" />
          </div>
          <div class="field"><label for="t-man">Gyártó</label><InputText id="t-man" v-model="te.form.manufacturer" /></div>
          <div class="field"><label for="t-mod">Típus</label><InputText id="t-mod" v-model="te.form.model" /></div>
          <div class="field">
            <label for="t-rate">Napi kölcsönzési díj</label>
            <InputNumber input-id="t-rate" v-model="te.form.daily_rate" :min="1" mode="currency" currency="HUF" locale="hu-HU" :max-fraction-digits="0" />
          </div>
          <div class="field">
            <label for="t-dep">Kaució</label>
            <InputNumber input-id="t-dep" v-model="te.form.deposit" :min="0" mode="currency" currency="HUF" locale="hu-HU" :max-fraction-digits="0" />
          </div>
          <div class="field">
            <label for="t-pp">Beszerzési ár</label>
            <InputNumber input-id="t-pp" v-model="te.form.purchase_price" :min="0" mode="currency" currency="HUF" locale="hu-HU" :max-fraction-digits="0" />
          </div>
        </div>
        <div>
          <h3 class="sub">Karbantartás esedékes ennyi után <span class="muted">(üresen hagyva a szabály nem él)</span></h3>
          <div class="form-grid">
            <div class="field">
              <label for="t-mr">Kölcsönzés</label>
              <InputNumber input-id="t-mr" v-model="te.form.maintenance_every_rentals" :min="1" suffix=" kölcsönzés" />
            </div>
            <div class="field">
              <label for="t-md">Nap</label>
              <InputNumber input-id="t-md" v-model="te.form.maintenance_every_days" :min="1" suffix=" nap" />
            </div>
            <div class="field">
              <label for="t-mh">Üzemóra</label>
              <InputNumber input-id="t-mh" v-model="te.form.maintenance_every_hours" :min="1" :max-fraction-digits="2" suffix=" üzemóra" />
            </div>
          </div>
        </div>
      </form>
      <template #footer>
        <Button label="Bezárás" severity="secondary" text @click="te.open = false" />
        <Button type="submit" form="type-form" label="Mentés" icon="pi pi-check" :loading="te.busy" />
      </template>
    </Dialog>

    <!-- add copies -->
    <Dialog :visible="!!ad.type" modal :header="`Példányok hozzáadása: ${ad.type?.name}`" :style="{ width: 'min(480px, 95vw)' }"
      @update:visible="ad.type = null">
      <div class="form-grid">
        <div class="field"><label for="a-n">Darabszám</label><InputNumber input-id="a-n" v-model="ad.count" :min="1" :max="100" show-buttons /></div>
        <div class="field">
          <label for="a-c">Állapot</label>
          <Select input-id="a-c" v-model="ad.condition" :options="conditionOptions.filter((c) => c.value !== 'DAMAGED')"
            option-label="label" option-value="value" />
        </div>
        <div class="field"><label for="a-d">Beszerzés dátuma</label><DatePicker input-id="a-d" v-model="ad.acquired" :max-date="new Date()" show-icon /></div>
      </div>
      <template #footer>
        <Button label="Bezárás" severity="secondary" text @click="ad.type = null" />
        <Button label="Hozzáadás" icon="pi pi-check" :loading="ad.busy" @click="addCopies" />
      </template>
    </Dialog>

    <!-- copy -->
    <Dialog :visible="!!ie.item" modal :header="`#${ie.item?.id} példány · ${ie.item?.tool_type_name}`"
      :style="{ width: 'min(620px, 95vw)' }" @update:visible="ie.item = null">
      <div v-if="ie.item" class="stack">
        <p v-if="ie.item.status === 'RENTED'" class="muted">Ez a példány ki van kölcsönözve. Előbb vegye vissza a Kölcsönzések oldalon.</p>
        <div v-else class="form-grid">
          <div class="field">
            <label for="i-s">Állapot</label>
            <Select input-id="i-s" v-model="ie.status" :options="itemStatusOptions" option-label="label" option-value="value" />
          </div>
          <div class="field">
            <label for="i-c">Fizikai állapot</label>
            <Select input-id="i-c" v-model="ie.condition" :options="conditionOptions" option-label="label" option-value="value" />
          </div>
        </div>
        <h3 class="sub">Karbantartási és javítási előzmények</h3>
        <DataTable :value="ie.history" size="small" data-key="id">
          <template #empty><div class="empty">Még nem volt karbantartás.</div></template>
          <Column header="Típus"><template #body="{ data }">{{ label(data.type) }}</template></Column>
          <Column field="description" header="Leírás" />
          <Column header="Kezdete"><template #body="{ data }">{{ dateTime(data.started_at) }}</template></Column>
          <Column header="Lezárva"><template #body="{ data }">{{ data.completed_at ? dateTime(data.completed_at) : 'Folyamatban' }}</template></Column>
          <Column header="Költség" class="num"><template #body="{ data }">{{ money(data.cost) }}</template></Column>
        </DataTable>
      </div>
      <template #footer>
        <Button v-if="ie.item && !['RENTED', 'RETIRED', 'MAINTENANCE'].includes(ie.item.status)" label="Karbantartás indítása"
          icon="pi pi-wrench" severity="secondary" outlined @click="startFromItem" />
        <Button label="Bezárás" severity="secondary" text @click="ie.item = null" />
        <Button v-if="ie.item?.status !== 'RENTED'" label="Mentés" icon="pi pi-check" :loading="ie.busy" @click="saveItem" />
      </template>
    </Dialog>

    <StartMaintenanceDialog :item-id="startItem" @close="startItem = null" @started="reloadAll" />

    <!-- category -->
    <Dialog v-model:visible="ce.open" modal :header="ce.id ? 'Kategória átnevezése' : 'Új kategória'" :style="{ width: 'min(420px, 95vw)' }">
      <form id="cat-form" class="field" @submit.prevent="saveCategory">
        <label for="c-n">Megnevezés</label>
        <InputText id="c-n" v-model="ce.name" required autofocus />
      </form>
      <template #footer>
        <Button label="Bezárás" severity="secondary" text @click="ce.open = false" />
        <Button type="submit" form="cat-form" label="Mentés" icon="pi pi-check" :loading="ce.busy" />
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.actions {
  display: inline-flex;
}
.counts {
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}
.filters {
  margin-bottom: 0.75rem;
}
.clickable :deep(tbody tr) {
  cursor: pointer;
}
.sub {
  margin: 0.5rem 0;
  font-size: 0.95rem;
  font-weight: 600;
}
</style>
