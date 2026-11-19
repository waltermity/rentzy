<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Card from 'primevue/card'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import Message from 'primevue/message'
import { useAuthStore } from '@/stores/auth'
import { errorMessage } from '@/api/client'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

const username = ref('')
const password = ref('')
const error = ref('')
const busy = ref(false)

async function submit() {
  error.value = ''
  busy.value = true
  try {
    await auth.login(username.value.trim(), password.value)
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : null
    router.replace(redirect && redirect.startsWith('/') ? redirect : auth.home())
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div class="page auth-page">
    <Card class="auth-card">
      <template #title>Bejelentkezés</template>
      <template #subtitle>Üdvözöljük a Rentzy szerszámkölcsönzőben</template>
      <template #content>
        <form class="stack" @submit.prevent="submit">
          <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>
          <div class="field">
            <label for="username">Felhasználónév</label>
            <InputText id="username" v-model="username" autocomplete="username" autofocus required />
          </div>
          <div class="field">
            <label for="password">Jelszó</label>
            <Password input-id="password" v-model="password" :feedback="false" toggle-mask fluid
              autocomplete="current-password" required />
          </div>
          <Button type="submit" label="Bejelentkezés" icon="pi pi-sign-in" :loading="busy" />
          <p class="muted switch">
            Még nincs fiókja?
            <RouterLink :to="{ path: '/register', query: route.query }">Regisztráció</RouterLink>
          </p>
        </form>
      </template>
    </Card>
  </div>
</template>

<style scoped>
.auth-page {
  display: grid;
  place-items: start center;
  padding-top: 3rem;
}
.auth-card {
  width: min(420px, 100%);
}
.switch {
  margin: 0;
  text-align: center;
}
</style>
