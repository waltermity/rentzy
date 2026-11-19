<script setup lang="ts">
import Tag from 'primevue/tag'
import { label } from '@/utils/format'
import { severity } from '@/utils/status'

defineProps<{ value: string }>()
</script>

<template>
  <Tag :value="label(value)" :severity="severity(value)" />
</template>
