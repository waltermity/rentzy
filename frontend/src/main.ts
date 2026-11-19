import { createApp } from 'vue'
import { createPinia } from 'pinia'
import PrimeVue from 'primevue/config'
import Aura from '@primeuix/themes/aura'
import ToastService from 'primevue/toastservice'
import ConfirmationService from 'primevue/confirmationservice'
import Tooltip from 'primevue/tooltip'
import 'primeicons/primeicons.css'
import './style.css'

import App from './App.vue'
import { router } from './router'
import { useAuthStore } from './stores/auth'
import { primeLocale } from './i18n/hu'

const app = createApp(App)
app.use(createPinia())
app.use(PrimeVue, {
  theme: { preset: Aura, options: { darkModeSelector: 'system' } },
  locale: primeLocale,
})
app.use(ToastService)
app.use(ConfirmationService)
app.directive('tooltip', Tooltip)

// Restore the session before the first navigation, so route guards know the role.
await useAuthStore().init()
app.use(router)
app.mount('#app')
