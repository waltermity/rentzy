<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import Button from 'primevue/button'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import Select from 'primevue/select'
import SelectButton from 'primevue/selectbutton'
import StatusTag from '@/components/StatusTag.vue'
import { customerApi, userApi } from '@/api'
import type { Customer, Role, User } from '@/api/types'
import { useAuthStore } from '@/stores/auth'
import { date } from '@/utils/format'
import { useNotify } from '@/utils/notify'

const auth = useAuthStore()
const notify = useNotify()

const users = ref<User[]>([])
const customers = ref<Customer[]>([])
const loading = ref(false)
const roleFilter = ref<Role | 'ALL'>('ALL')

async function load() {
  loading.value = true
  try {
    const [u, c] = await Promise.all([userApi.list(), customerApi.list()])
    users.value = u
    customers.value = c
  } catch (e) {
    notify.fail(e, 'A felhasználók betöltése nem sikerült')
  } finally {
    loading.value = false
  }
}
onMounted(load)

const visible = computed(() => users.value.filter((u) => roleFilter.value === 'ALL' || u.role === roleFilter.value))
const customerName = (id: number | null) => customers.value.find((c) => c.id === id)?.name ?? ''
const roleOptions = [
  { label: 'Mind', value: 'ALL' },
  { label: 'Adminisztrátor', value: 'ADMIN' },
  { label: 'Személyzet', value: 'STAFF' },
  { label: 'Ügyfél', value: 'CUSTOMER' },
]
const roleChoices: { label: string; value: Role }[] = [
  { label: 'Adminisztrátor', value: 'ADMIN' },
  { label: 'Személyzet', value: 'STAFF' },
  { label: 'Ügyfél', value: 'CUSTOMER' },
]
/** Customers that have no login yet. */
const unlinked = computed(() => customers.value.filter((c) => !users.value.some((u) => u.customer_id === c.id)))

// ---------- create ----------
const cr = reactive({
  open: false, busy: false,
  username: '', email: '', password: '', role: 'STAFF' as Role,
  link: 'new' as 'new' | 'existing', customerId: null as number | null,
  name: '', phone: '', address: '',
})
function openCreate() {
  Object.assign(cr, {
    open: true, username: '', email: '', password: '', role: 'STAFF', link: 'new', customerId: null,
    name: '', phone: '', address: '',
  })
}
async function create() {
  cr.busy = true
  try {
    const body: Parameters<typeof userApi.create>[0] = {
      username: cr.username.trim(), email: cr.email.trim(), password: cr.password, role: cr.role,
    }
    if (cr.role === 'CUSTOMER') {
      if (cr.link === 'existing') body.customer_id = cr.customerId ?? undefined
      else body.customer = { name: cr.name, email: cr.email.trim(), phone: cr.phone, address: cr.address }
    }
    await userApi.create(body)
    notify.ok('Felhasználó létrehozva', cr.username)
    cr.open = false
    await load()
  } catch (e) {
    notify.fail(e, 'A felhasználó létrehozása nem sikerült')
  } finally {
    cr.busy = false
  }
}

// ---------- edit ----------
const ed = reactive({ user: null as User | null, email: '', role: 'STAFF' as Role, password: '', busy: false })
function openEdit(u: User) {
  Object.assign(ed, { user: u, email: u.email, role: u.role, password: '' })
}
async function save() {
  if (!ed.user) return
  ed.busy = true
  try {
    await userApi.update(ed.user.id, {
      email: ed.email !== ed.user.email ? ed.email : undefined,
      role: ed.role !== ed.user.role ? ed.role : undefined,
      password: ed.password || undefined,
    })
    notify.ok('Felhasználó módosítva')
    ed.user = null
    await load()
  } catch (e) {
    notify.fail(e, 'A módosítás nem sikerült')
  } finally {
    ed.busy = false
  }
}
</script>

<template>
  <div class="page">
    <div class="page-head">
      <h1>Felhasználók</h1>
      <div class="row">
        <SelectButton v-model="roleFilter" :options="roleOptions" option-label="label" option-value="value"
          :allow-empty="false" aria-label="Szűrés szerepkörre" />
        <Button label="Új felhasználó" icon="pi pi-plus" @click="openCreate" />
      </div>
    </div>

    <DataTable :value="visible" :loading="loading" data-key="id" striped-rows paginator :rows="20">
      <Column field="id" header="#" />
      <Column field="username" header="Felhasználónév" sortable />
      <Column field="email" header="E-mail" />
      <Column header="Szerepkör" field="role" sortable><template #body="{ data }"><StatusTag :value="data.role" /></template></Column>
      <Column header="Ügyfél"><template #body="{ data }">{{ customerName(data.customer_id) }}</template></Column>
      <Column header="Létrehozva"><template #body="{ data }">{{ date(data.created_at) }}</template></Column>
      <Column header="" class="num">
        <template #body="{ data }">
          <Button icon="pi pi-pencil" text size="small" aria-label="Szerkesztés" @click="openEdit(data)" />
        </template>
      </Column>
    </DataTable>

    <Dialog v-model:visible="cr.open" modal header="Új felhasználó" :style="{ width: 'min(620px, 95vw)' }">
      <form id="user-form" class="stack" @submit.prevent="create">
        <div class="form-grid">
          <div class="field"><label for="u-name">Felhasználónév</label><InputText id="u-name" v-model="cr.username" required minlength="3" /></div>
          <div class="field"><label for="u-email">E-mail cím</label><InputText id="u-email" v-model="cr.email" type="email" required /></div>
          <div class="field">
            <label for="u-pw">Kezdeti jelszó</label>
            <Password input-id="u-pw" v-model="cr.password" toggle-mask fluid :feedback="false" autocomplete="new-password" required />
          </div>
          <div class="field">
            <label for="u-role">Szerepkör</label>
            <Select input-id="u-role" v-model="cr.role" :options="roleChoices" option-label="label" option-value="value" />
          </div>
        </div>
        <template v-if="cr.role === 'CUSTOMER'">
          <SelectButton v-model="cr.link" :allow-empty="false"
            :options="[{ label: 'Új ügyfél', value: 'new' }, { label: 'Meglévő ügyfél', value: 'existing' }]"
            option-label="label" option-value="value" />
          <div v-if="cr.link === 'existing'" class="field">
            <label for="u-cust">Fiók nélküli ügyfél</label>
            <Select input-id="u-cust" v-model="cr.customerId" :options="unlinked" option-label="name" option-value="id"
              filter placeholder="Válasszon ügyfelet" />
          </div>
          <div v-else class="form-grid">
            <div class="field"><label for="u-full">Teljes név</label><InputText id="u-full" v-model="cr.name" required /></div>
            <div class="field"><label for="u-phone">Telefonszám</label><InputText id="u-phone" v-model="cr.phone" /></div>
            <div class="field"><label for="u-addr">Cím</label><InputText id="u-addr" v-model="cr.address" /></div>
          </div>
        </template>
      </form>
      <template #footer>
        <Button label="Bezárás" severity="secondary" text @click="cr.open = false" />
        <Button type="submit" form="user-form" label="Létrehozás" icon="pi pi-check" :loading="cr.busy" />
      </template>
    </Dialog>

    <Dialog :visible="!!ed.user" modal :header="`${ed.user?.username} szerkesztése`" :style="{ width: 'min(480px, 95vw)' }"
      @update:visible="ed.user = null">
      <div v-if="ed.user" class="stack">
        <div class="field"><label for="e-email">E-mail cím</label><InputText id="e-email" v-model="ed.email" type="email" /></div>
        <div class="field">
          <label for="e-role">Szerepkör</label>
          <Select v-if="ed.user.role !== 'CUSTOMER'" input-id="e-role" v-model="ed.role" :options="roleChoices.filter((r) => r.value !== 'CUSTOMER')"
            option-label="label" option-value="value"
            :disabled="ed.user.id === auth.user?.id" />
          <span v-else class="muted">Ügyfél fiók szerepköre nem módosítható.</span>
          <small v-if="ed.user.id === auth.user?.id" class="muted">A saját szerepkörét nem módosíthatja.</small>
        </div>
        <div class="field">
          <label for="e-pw">Új jelszó <span class="muted">(üresen hagyva nem változik)</span></label>
          <Password input-id="e-pw" v-model="ed.password" toggle-mask fluid :feedback="false" autocomplete="new-password" />
        </div>
      </div>
      <template #footer>
        <Button label="Bezárás" severity="secondary" text @click="ed.user = null" />
        <Button label="Mentés" icon="pi pi-check" :loading="ed.busy" @click="save" />
      </template>
    </Dialog>
  </div>
</template>
