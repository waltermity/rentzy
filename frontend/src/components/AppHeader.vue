<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import Menu from 'primevue/menu'
import Tag from 'primevue/tag'
import { useAuthStore } from '@/stores/auth'
import { label } from '@/utils/format'
import { severity } from '@/utils/status'

const auth = useAuthStore()
const router = useRouter()

// Navigation links by role.
const links = computed(() => {
  const l = [{ to: '/catalog', text: 'Katalógus', icon: 'pi pi-th-large' }]
  if (auth.isStaff) {
    l.push(
      { to: '/staff/rentals', text: 'Kölcsönzések', icon: 'pi pi-arrow-right-arrow-left' },
      { to: '/staff/customers', text: 'Ügyfelek', icon: 'pi pi-id-card' },
      { to: '/staff/maintenance', text: 'Karbantartás', icon: 'pi pi-wrench' },
    )
  }
  if (auth.isAdmin) {
    l.push(
      { to: '/admin/tools', text: 'Eszközök', icon: 'pi pi-box' },
      { to: '/admin/users', text: 'Felhasználók', icon: 'pi pi-users' },
    )
  }
  if (auth.isStaff) l.push({ to: '/admin/reports', text: 'Kimutatások', icon: 'pi pi-chart-bar' })
  return l
})

const menu = ref<InstanceType<typeof Menu>>()
const menuItems = [
  { label: 'Profil', icon: 'pi pi-user', command: () => router.push('/profile') },
  { separator: true },
  {
    label: 'Kijelentkezés',
    icon: 'pi pi-sign-out',
    command: () => {
      auth.logout()
      router.push('/catalog')
    },
  },
]
</script>

<template>
  <header class="header">
    <div class="inner">
      <RouterLink to="/catalog" class="brand" aria-label="Rentzy kezdőlap">
        <img src="/favicon.svg" alt="" width="28" height="28" />
        <span>Rentzy</span>
      </RouterLink>

      <nav class="nav" :class="{ many: links.length > 5 }" aria-label="Főmenü">
        <RouterLink v-for="l in links" :key="l.to" :to="l.to" class="nav-link" :title="l.text" :aria-label="l.text">
          <i :class="l.icon" aria-hidden="true" />
          <span>{{ l.text }}</span>
        </RouterLink>
      </nav>

      <div class="account">
        <template v-if="auth.isLoggedIn">
          <button class="user-btn" aria-haspopup="true" aria-controls="user-menu" @click="menu?.toggle($event)">
            <span class="avatar" aria-hidden="true"><i class="pi pi-user" /></span>
            <span class="who">
              <span class="name">{{ auth.displayName }}</span>
              <Tag :value="label(auth.role!)" :severity="severity(auth.role!)" class="role" />
            </span>
            <i class="pi pi-angle-down" aria-hidden="true" />
          </button>
          <Menu id="user-menu" ref="menu" :model="menuItems" popup />
        </template>
        <Button v-else as="router-link" to="/login" label="Bejelentkezés" icon="pi pi-sign-in" size="small" />
      </div>
    </div>
  </header>
</template>

<style scoped>
.header {
  position: sticky;
  top: 0;
  z-index: 100;
  background: var(--app-header-bg);
  border-bottom: 1px solid var(--app-border);
}
.inner {
  max-width: var(--app-max);
  margin: 0 auto;
  padding: 0.5rem 1rem;
  display: flex;
  align-items: center;
  gap: 1rem;
}
.brand {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 700;
  font-size: 1.1rem;
  color: var(--p-text-color);
  text-decoration: none;
  flex-shrink: 0;
}
.nav {
  display: flex;
  gap: 0.25rem;
  flex: 1;
  min-width: 0;
  overflow-x: auto;
  scrollbar-width: none;
}
.nav-link {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.45rem 0.7rem;
  border-radius: 8px;
  color: var(--app-muted);
  text-decoration: none;
  white-space: nowrap;
  font-weight: 500;
}
.nav-link:hover {
  background: var(--p-content-hover-background);
  color: var(--p-text-color);
}
.nav-link.router-link-active {
  color: var(--p-primary-color);
  background: var(--p-highlight-background);
}
.account {
  flex-shrink: 0;
}
.user-btn {
  display: flex;
  align-items: center;
  gap: 0.55rem;
  padding: 0.3rem 0.5rem;
  border: 0;
  border-radius: 8px;
  background: transparent;
  color: var(--p-text-color);
  cursor: pointer;
  font: inherit;
}
.user-btn:hover {
  background: var(--p-content-hover-background);
}
.avatar {
  display: grid;
  place-items: center;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: var(--p-highlight-background);
  color: var(--p-primary-color);
}
.who {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}
.name {
  font-weight: 600;
  max-width: 14rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.role {
  font-size: 0.7rem;
}
/* Admins have 7 links: show icons only (with tooltips) when they would not fit. */
@media (max-width: 1180px) {
  .nav.many .nav-link span {
    display: none;
  }
}
@media (max-width: 720px) {
  .nav-link span,
  .role {
    display: none;
  }
  .name {
    max-width: 8rem;
  }
}
@media (max-width: 420px) {
  .brand span,
  .name {
    display: none;
  }
}
</style>
