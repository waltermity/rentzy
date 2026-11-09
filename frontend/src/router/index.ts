import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import type { Role } from '@/api/types'

declare module 'vue-router' {
  interface RouteMeta {
    title?: string
    /** Logged-in users with one of these roles; empty list = any logged-in user. */
    roles?: Role[]
    /** Only for visitors who are not logged in (login, register). */
    guestOnly?: boolean
  }
}

const STAFF: Role[] = ['STAFF', 'ADMIN']
const ADMIN: Role[] = ['ADMIN']

const routes: RouteRecordRaw[] = [
  { path: '/', redirect: '/catalog' },
  // public
  { path: '/catalog', component: () => import('@/views/CatalogView.vue'), meta: { title: 'Katalógus' } },
  { path: '/catalog/:id(\\d+)', component: () => import('@/views/ToolDetailView.vue'), meta: { title: 'Eszköz' } },
  { path: '/login', component: () => import('@/views/LoginView.vue'), meta: { title: 'Bejelentkezés', guestOnly: true } },
  { path: '/register', component: () => import('@/views/RegisterView.vue'), meta: { title: 'Regisztráció', guestOnly: true } },
  // any logged-in user
  { path: '/profile', component: () => import('@/views/ProfileView.vue'), meta: { title: 'Profil', roles: [] } },
  // staff + admin
  { path: '/staff/rentals', component: () => import('@/views/staff/RentalsView.vue'), meta: { title: 'Kölcsönzések', roles: STAFF } },
  { path: '/staff/maintenance', component: () => import('@/views/staff/MaintenanceView.vue'), meta: { title: 'Karbantartás', roles: STAFF } },
  { path: '/staff/customers', component: () => import('@/views/staff/CustomersView.vue'), meta: { title: 'Ügyfelek', roles: STAFF } },
  { path: '/admin/reports', component: () => import('@/views/admin/ReportsView.vue'), meta: { title: 'Kimutatások', roles: STAFF } },
  // admin
  { path: '/admin/tools', component: () => import('@/views/admin/ToolsView.vue'), meta: { title: 'Eszközök', roles: ADMIN } },
  { path: '/admin/users', component: () => import('@/views/admin/UsersView.vue'), meta: { title: 'Felhasználók', roles: ADMIN } },
  { path: '/:pathMatch(.*)*', component: () => import('@/views/NotFoundView.vue'), meta: { title: 'Nem található' } },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior: () => ({ top: 0 }),
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.guestOnly && auth.isLoggedIn) return auth.home()
  if (to.meta.roles) {
    if (!auth.isLoggedIn) return { path: '/login', query: { redirect: to.fullPath } }
    if (to.meta.roles.length && !auth.hasRole(to.meta.roles)) return auth.home()
  }
})

router.afterEach((to) => {
  document.title = to.meta.title ? `${to.meta.title} · Rentzy` : 'Rentzy'
})
