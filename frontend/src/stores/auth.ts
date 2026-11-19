import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { authApi } from '@/api'
import { getToken, setToken, setUnauthorizedHandler } from '@/api/client'
import type { Customer, Role, Session, User } from '@/api/types'

/** JWT, the logged-in user and their role (RBAC). */
export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(getToken())
  const user = ref<User | null>(null)
  const customer = ref<Customer | null>(null)
  const ready = ref(false)

  const isLoggedIn = computed(() => !!token.value && !!user.value)
  const role = computed<Role | null>(() => user.value?.role ?? null)
  const isAdmin = computed(() => role.value === 'ADMIN')
  const isStaff = computed(() => role.value === 'STAFF' || role.value === 'ADMIN')
  const isCustomer = computed(() => role.value === 'CUSTOMER')
  const displayName = computed(() => customer.value?.name ?? user.value?.username ?? '')

  function apply(s: Session) {
    token.value = s.token
    setToken(s.token)
    user.value = s.user
    customer.value = s.customer ?? null
  }

  function logout() {
    token.value = null
    setToken(null)
    user.value = null
    customer.value = null
  }

  /** Restores the session from the stored token (on page load). */
  async function init() {
    setUnauthorizedHandler(logout)
    if (token.value) {
      try {
        const me = await authApi.me()
        user.value = me.user
        customer.value = me.customer ?? null
      } catch {
        logout()
      }
    }
    ready.value = true
  }

  async function login(username: string, password: string) {
    apply(await authApi.login(username, password))
  }

  async function register(body: Parameters<typeof authApi.register>[0]) {
    apply(await authApi.register(body))
  }

  /** Landing page after login, by role. */
  function home(): string {
    if (role.value === 'ADMIN') return '/admin/reports'
    if (role.value === 'STAFF') return '/staff/rentals'
    return '/catalog'
  }

  function hasRole(roles?: Role[]) {
    return !roles || (role.value != null && roles.includes(role.value))
  }

  return {
    token, user, customer, ready,
    isLoggedIn, role, isAdmin, isStaff, isCustomer, displayName,
    init, login, register, logout, home, hasRole,
  }
})
