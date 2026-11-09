<script setup lang="ts">
import { reactive, watch } from 'vue'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Textarea from 'primevue/textarea'
import { maintenanceApi } from '@/api'
import { MAINTENANCE_TYPES, type MaintenanceType } from '@/api/types'
import { useAuthStore } from '@/stores/auth'
import { label } from '@/utils/format'
import { useNotify } from '@/utils/notify'

const props = defineProps<{ itemId: number | null; title?: string }>()
const emit = defineEmits<{ close: []; started: [] }>()
const auth = useAuthStore()
const notify = useNotify()

const form = reactive({ type: 'SCHEDULED' as MaintenanceType, description: '', performedBy: '', busy: false })
watch(
  () => props.itemId,
  (id) => {
    if (id != null) Object.assign(form, { type: 'SCHEDULED', description: '', performedBy: auth.user?.username ?? '' })
  },
)

async function submit() {
  if (props.itemId == null) return
  form.busy = true
  try {
    await maintenanceApi.start(props.itemId, {
      type: form.type,
      description: form.description,
      performed_by: form.performedBy,
    })
    notify.ok('Karbantartás elindítva', `A #${props.itemId} példány karbantartás alatt van`)
    emit('started')
    emit('close')
  } catch (e) {
    notify.fail(e, 'A karbantartás indítása nem sikerült')
  } finally {
    form.busy = false
  }
}
</script>

<template>
  <Dialog :visible="itemId != null" modal :header="title ?? `Karbantartás indítása: #${itemId} példány`"
    :style="{ width: 'min(480px, 95vw)' }" @update:visible="emit('close')">
    <div class="stack">
      <div class="field">
        <label for="sm-type">Típus</label>
        <Select input-id="sm-type" v-model="form.type"
          :options="MAINTENANCE_TYPES.map((t) => ({ label: label(t), value: t }))" option-label="label"
          option-value="value" />
      </div>
      <div class="field">
        <label for="sm-desc">Leírás</label>
        <Textarea id="sm-desc" v-model="form.description" rows="3" auto-resize />
      </div>
      <div class="field">
        <label for="sm-by">Végzi</label>
        <InputText id="sm-by" v-model="form.performedBy" />
      </div>
      <p class="muted note">A példány a karbantartás lezárásáig nem foglalható és nem kölcsönözhető.</p>
    </div>
    <template #footer>
      <Button label="Bezárás" severity="secondary" text @click="emit('close')" />
      <Button label="Indítás" icon="pi pi-wrench" :loading="form.busy" @click="submit" />
    </template>
  </Dialog>
</template>

<style scoped>
.note {
  margin: 0;
  font-size: 0.85rem;
}
</style>
