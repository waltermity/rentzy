<script setup lang="ts">
import { reactive, ref } from 'vue'
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

const form = reactive({ username: '', email: '', password: '', password2: '', name: '', phone: '', address: '' })
const error = ref('')
const busy = ref(false)

async function submit() {
  error.value = ''
  if (form.password !== form.password2) {
    error.value = 'A két jelszó nem egyezik.'
    return
  }
  busy.value = true
  try {
    const { password2: _, ...body } = form
    await auth.register({ ...body, username: body.username.trim(), email: body.email.trim() })
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
      <template #title>Regisztráció</template>
      <template #subtitle>Regisztráljon, és foglaljon szerszámot online</template>
      <template #content>
        <form class="stack" @submit.prevent="submit">
          <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>
          <div class="form-grid">
            <div class="field">
              <label for="name">Teljes név</label>
              <InputText id="name" v-model="form.name" autocomplete="name" required />
            </div>
            <div class="field">
              <label for="email">E-mail cím</label>
              <InputText id="email" v-model="form.email" type="email" autocomplete="email" required />
            </div>
            <div class="field">
              <label for="username">Felhasználónév</label>
              <InputText id="username" v-model="form.username" autocomplete="username" minlength="3" required />
            </div>
            <div class="field">
              <label for="phone">Telefonszám (nem kötelező)</label>
              <InputText id="phone" v-model="form.phone" type="tel" autocomplete="tel" />
            </div>
          </div>
          <div class="field">
            <label for="address">Cím (nem kötelező)</label>
            <InputText id="address" v-model="form.address" autocomplete="street-address" />
          </div>
          <div class="form-grid">
            <div class="field">
              <label for="password">Jelszó</label>
              <Password input-id="password" v-model="form.password" toggle-mask fluid autocomplete="new-password"
                :input-props="{ minlength: 8 }" required />
              <small class="muted">Legalább 8 karakter.</small>
            </div>
            <div class="field">
              <label for="password2">Jelszó még egyszer</label>
              <Password input-id="password2" v-model="form.password2" :feedback="false" toggle-mask fluid
                autocomplete="new-password" required />
            </div>
          </div>
          <Button type="submit" label="Fiók létrehozása" icon="pi pi-user-plus" :loading="busy" />
          <p class="muted switch">
            Már van fiókja?
            <RouterLink :to="{ path: '/login', query: route.query }">Bejelentkezés</RouterLink>
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
  padding-top: 2rem;
}
.auth-card {
  width: min(620px, 100%);
}
.switch {
  margin: 0;
  text-align: center;
}
</style>
