import { readonly, ref } from 'vue'
import { k8sService } from '../../services/k8sService'
import type { IAMSecurityRecord } from './types'

const records = ref<IAMSecurityRecord[]>([])
const isLoading = ref(false)
const hasLoadedOnce = ref(false)
const lastSyncedAt = ref('')

const toSyncTime = (): string => {
  return new Date().toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  })
}

const fetchAndCache = async (refresh: boolean) => {
  if (isLoading.value) return
  isLoading.value = true
  try {
    const data = await k8sService.fetchTableData('iam', refresh)
    records.value = Array.isArray(data) ? data : []
    hasLoadedOnce.value = true
    lastSyncedAt.value = toSyncTime()
  } catch (err) {
    console.error('IAM security fetch failed:', err)
    records.value = []
  } finally {
    isLoading.value = false
  }
}

const ensureLoaded = async () => {
  if (hasLoadedOnce.value) return
  await fetchAndCache(false)
}

const refresh = async () => {
  await fetchAndCache(true)
}

export const iamSecurityStore = {
  records: readonly(records),
  isLoading: readonly(isLoading),
  hasLoadedOnce: readonly(hasLoadedOnce),
  lastSyncedAt: readonly(lastSyncedAt),
  ensureLoaded,
  refresh,
}
