<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import Card from 'primevue/card'
import InputText from 'primevue/inputtext'
import IconField from 'primevue/iconfield'
import InputIcon from 'primevue/inputicon'
import Select from 'primevue/select'
import Skeleton from 'primevue/skeleton'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import { useToolStore } from '@/stores/tools'
import { errorMessage } from '@/api/client'
import { money } from '@/utils/format'

const tools = useToolStore()
const search = ref('')
const categoryId = ref<number | null>(null)
const error = ref('')

onMounted(async () => {
  try {
    await tools.loadCatalog(true)
  } catch (e) {
    error.value = errorMessage(e)
  }
})


const visible = computed(() => {
  const q = search.value.trim().toLowerCase()
  return tools.types.filter(
    (t) =>
      (categoryId.value == null || t.category_id === categoryId.value) &&
      (!q || [t.name, t.manufacturer, t.model].some((s) => s.toLowerCase().includes(q))),
  )
})

/** Copies that can be booked at all (not in maintenance, broken or retired). */
const bookable = (t: { item_counts?: Record<string, number | undefined> }) =>
  (t.item_counts?.AVAILABLE ?? 0) + (t.item_counts?.RENTED ?? 0)
</script>

<template>
  <div class="page">
    <div class="page-head">
      <h1>Szerszámkatalógus</h1>
      <div class="row filters">
        <IconField>
          <InputIcon class="pi pi-search" />
          <InputText v-model="search" placeholder="Keresés: név, gyártó, típus" aria-label="Szerszám keresése" />
        </IconField>
        <Select v-model="categoryId" :options="tools.categories" option-label="name" option-value="id"
          placeholder="Minden kategória" show-clear aria-label="Kategória" class="cat" />
      </div>
    </div>

    <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>

    <div v-if="tools.loading && !tools.types.length" class="grid">
      <Skeleton v-for="i in 6" :key="i" height="190px" border-radius="12px" />
    </div>

    <div v-else-if="visible.length" class="grid">
      <RouterLink v-for="t in visible" :key="t.id" :to="`/catalog/${t.id}`" class="card-link">
        <Card class="tool-card">
          <template #title>{{ t.name }}</template>
          <template #subtitle>{{ [t.manufacturer, t.model].filter(Boolean).join(' · ') || '—' }}</template>
          <template #content>
            <Tag :value="t.category_name" severity="secondary" class="cat-tag" />
            <div class="price">
              <span class="rate">{{ money(t.daily_rate) }}</span>
              <span class="muted"> / nap</span>
            </div>
            <div class="muted small">Kaució: {{ money(t.deposit) }}</div>
          </template>
          <template #footer>
            <span v-if="(t.item_counts?.AVAILABLE ?? 0) > 0" class="avail ok">
              <i class="pi pi-check-circle" /> {{ t.item_counts?.AVAILABLE }} db elérhető most
            </span>
            <span v-else-if="bookable(t) > 0" class="avail">
              <i class="pi pi-clock" /> Most mind kint van, későbbre foglalható
            </span>
            <span v-else class="avail off"><i class="pi pi-ban" /> Nem elérhető</span>
          </template>
        </Card>
      </RouterLink>
    </div>

    <div v-else-if="!tools.loading" class="empty">Nincs a keresésnek megfelelő szerszám.</div>
  </div>
</template>

<style scoped>
.filters .cat {
  min-width: 13rem;
}
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 1rem;
}
.card-link {
  text-decoration: none;
  color: inherit;
}
.tool-card {
  height: 100%;
  border: 1px solid var(--app-border);
  transition: border-color 0.15s, transform 0.15s;
}
.card-link:hover .tool-card,
.card-link:focus-visible .tool-card {
  border-color: var(--p-primary-color);
  transform: translateY(-2px);
}
.cat-tag {
  margin-bottom: 0.75rem;
}
.price .rate {
  font-size: 1.35rem;
  font-weight: 700;
}
.small {
  font-size: 0.85rem;
}
.avail {
  font-size: 0.85rem;
  color: var(--app-muted);
  display: inline-flex;
  gap: 0.35rem;
  align-items: center;
}
.avail.ok {
  color: var(--p-green-500);
}
.avail.off {
  color: var(--p-red-400);
}
@media (max-width: 520px) {
  .filters,
  .filters > * {
    width: 100%;
  }
}
</style>
