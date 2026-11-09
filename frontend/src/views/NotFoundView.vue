<script setup lang="ts">
import Button from 'primevue/button'
</script>

<template>
  <div class="page empty">
    <h1>Az oldal nem található</h1>
    <p>A keresett oldal nem létezik.</p>
    <Button as="router-link" to="/catalog" label="Vissza a katalógushoz" icon="pi pi-arrow-left" />
  </div>
</template>
