import { ref } from 'vue'
import { defineStore } from 'pinia'
import { toolApi } from '@/api'
import type { Category, ToolType } from '@/api/types'

/** Catalog cache: categories and tool types. */
export const useToolStore = defineStore('tools', () => {
  const categories = ref<Category[]>([])
  const types = ref<ToolType[]>([])
  const loaded = ref(false)
  const loading = ref(false)

  async function loadCatalog(force = false) {
    if (loaded.value && !force) return
    loading.value = true
    try {
      const [cats, ts] = await Promise.all([toolApi.categories(), toolApi.types()])
      categories.value = cats
      types.value = ts
      loaded.value = true
    } finally {
      loading.value = false
    }
  }

  /** One tool type, from the cache if present, else from the API. */
  async function getType(id: number, fresh = false): Promise<ToolType> {
    const cached = types.value.find((t) => t.id === id)
    if (cached && !fresh) return cached
    const t = await toolApi.type(id)
    const i = types.value.findIndex((x) => x.id === id)
    if (i >= 0) types.value[i] = t
    return t
  }

  function invalidate() {
    loaded.value = false
  }

  return { categories, types, loaded, loading, loadCatalog, getType, invalidate }
})
