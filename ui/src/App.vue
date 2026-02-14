<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import {
  ChevronDown,
  ChevronRight,
  Download,
  Fingerprint,
  Network,
  RefreshCw,
  Search,
  ShieldCheck,
  Users,
  Box
} from 'lucide-vue-next'
import { k8sService } from './services/k8sService'
import IdentityCard from './components/IdentityCard.vue'
import YamlInspector from './components/YamlInspector.vue'
import IAMSecurityFeature from './features/iam-security/IAMSecurityFeature.vue'
import { FEATURE_FLAGS, K8S_RESOURCE_TABS } from './config/features.config'

const enabledResourceTabs = computed(() =>
  K8S_RESOURCE_TABS.filter((tab) => FEATURE_FLAGS[tab.flag])
)
const areK8sResourcesVisible = computed(() =>
  FEATURE_FLAGS.SHOW_K8S_RESOURCES && enabledResourceTabs.value.length > 0
)

const getDefaultTab = () => {
  if (window.location.pathname === '/iam-security') return 'iam-security'
  if (FEATURE_FLAGS.SHOW_K8S_RESOURCES && enabledResourceTabs.value.length > 0) {
    return enabledResourceTabs.value[0].key
  }
  return 'iam-security'
}

const currentTab = ref(getDefaultTab())
const isK8sResourcesExpanded = ref(true)
const detailTab = ref('map')

const saData = ref([])
const groupData = ref([])
const ciliumData = ref([])
const selectedItem = ref(null)
const searchQuery = ref('')
const namespaceFilter = ref('')
const isLoading = ref(false)
const lastUpdated = ref(new Date().toLocaleTimeString())

const isIAMTab = computed(() => currentTab.value === 'iam-security')
const isResourceTab = computed(() =>
  FEATURE_FLAGS.SHOW_K8S_RESOURCES &&
  enabledResourceTabs.value.some((tab) => tab.key === currentTab.value)
)

const currentTabLabel = computed(() => {
  if (currentTab.value === 'iam-security') return 'IAM Navigator'
  const match = enabledResourceTabs.value.find((tab) => tab.key === currentTab.value)
  return match?.label || 'Resources'
})

const sourceData = computed(() => {
  if (currentTab.value === 'sa') return saData.value
  if (currentTab.value === 'groups') return groupData.value
  if (currentTab.value === 'cilium') return ciliumData.value
  return []
})

const namespaces = computed(() => {
  if (!isResourceTab.value) return []
  const data = sourceData.value
  if (!Array.isArray(data)) return []

  const set = new Set()
  for (const item of data) {
    if (item.namespace) set.add(item.namespace)
    if (Array.isArray(item.namespaces)) {
      item.namespaces.forEach((ns) => { if (ns) set.add(ns) })
    }
  }

  if (currentTab.value === 'cilium' && data.some((item) => item.is_cluster_wide)) {
    set.add('Global')
  }
  return [...set].sort()
})

const handleTab = (tab) => {
  if (!FEATURE_FLAGS.SHOW_K8S_RESOURCES && tab !== 'iam-security') return
  if (currentTab.value === tab) return

  currentTab.value = tab
  selectedItem.value = null
  searchQuery.value = ''
  namespaceFilter.value = ''
  detailTab.value = 'map'

  if (tab === 'iam-security') {
    window.history.replaceState({}, '', '/iam-security')
    return
  }

  window.history.replaceState({}, '', '/')
  fetchData(false)
}

const fetchData = async (refresh = false) => {
  if (!isResourceTab.value) return

  isLoading.value = true
  try {
    const rawData = await k8sService.fetchTableData(currentTab.value, refresh)

    if (currentTab.value === 'sa') {
      const grouped = rawData.reduce((acc, curr) => {
        const key = `${curr.sa}-${curr.namespace}`
        if (!acc[key]) {
          acc[key] = { ...curr, roles: [curr.role], bindings: [curr.binding_name] }
        } else {
          if (!acc[key].roles.includes(curr.role)) acc[key].roles.push(curr.role)
          if (!acc[key].bindings.includes(curr.binding_name)) acc[key].bindings.push(curr.binding_name)
        }
        return acc
      }, {})
      saData.value = Object.values(grouped)
    } else if (currentTab.value === 'groups') {
      groupData.value = [...rawData]
    } else if (currentTab.value === 'cilium') {
      ciliumData.value = [...rawData]
    }

    lastUpdated.value = new Date().toLocaleTimeString()
  } catch (err) {
    console.error('Fetch failed:', err)
  } finally {
    isLoading.value = false
  }
}

const refreshData = () => {
  fetchData(true)
}

const filteredData = computed(() => {
  if (!isResourceTab.value) return []

  const query = searchQuery.value.toLowerCase().trim()
  const ns = namespaceFilter.value
  const data = sourceData.value

  return data.filter((item) => {
    let matchesNS = false

    if (currentTab.value === 'sa') {
      matchesNS = !ns || item.namespace === ns
    } else if (currentTab.value === 'groups') {
      matchesNS = !ns || (item.namespaces && item.namespaces.includes(ns))
    } else {
      if (!ns) {
        matchesNS = true
      } else if (ns === 'Global') {
        matchesNS = item.is_cluster_wide
      } else {
        matchesNS = item.namespace === ns && !item.is_cluster_wide
      }
    }

    if (!matchesNS) return false
    if (!query) return true

    const searchStr = currentTab.value === 'cilium'
      ? `${item.name} ${item.target_selector}`
      : currentTab.value === 'sa'
        ? `${item.sa} ${item.role_name} ${item.iam_role}`
        : `${item.group_name} ${item.roles?.join(' ')}`

    return searchStr.toLowerCase().includes(query)
  })
})

const selectedTitle = computed(() => {
  if (!selectedItem.value) return ''
  return selectedItem.value.name || selectedItem.value.sa || selectedItem.value.group_name || 'Resource'
})

const selectedSubtitle = computed(() => {
  if (!selectedItem.value) return ''
  if (currentTab.value === 'sa') return `${selectedItem.value.namespace} namespace`
  if (currentTab.value === 'groups') return 'Group subject and RBAC bindings'
  if (selectedItem.value.is_cluster_wide) return 'Cluster-wide policy'
  return `${selectedItem.value.namespace || 'Namespace'} policy`
})

const aiInsight = computed(() => {
  if (!selectedItem.value) return 'Select a resource to inspect risk posture and policy scope.'

  if (currentTab.value === 'sa') {
    const scope = selectedItem.value.is_global ? 'cluster-wide' : `namespace ${selectedItem.value.namespace}`
    const iamRole = selectedItem.value.iam_role === 'None' ? 'without IAM role mapping' : 'with IAM role mapping'
    return `Service account ${selectedItem.value.sa} is bound to ${selectedItem.value.role_name || 'a role'} in ${scope}, ${iamRole}.`
  }

  if (currentTab.value === 'groups') {
    const nsCount = Array.isArray(selectedItem.value.namespaces) ? selectedItem.value.namespaces.length : 0
    return `Group ${selectedItem.value.group_name} inherits ${selectedItem.value.roles?.length || 0} RBAC role mappings across ${nsCount} scope(s).`
  }

  const clusterScope = selectedItem.value.is_cluster_wide ? 'cluster-wide' : `namespace ${selectedItem.value.namespace}`
  return `Cilium policy ${selectedItem.value.name} enforces ${selectedItem.value.type} controls at ${clusterScope} scope targeting ${selectedItem.value.target_selector}.`
})

const mapRows = computed(() => {
  if (!selectedItem.value) return []

  if (currentTab.value === 'sa') {
    return [{
      left: selectedItem.value.sa,
      middle: selectedItem.value.binding_name,
      right: selectedItem.value.role_name || selectedItem.value.role
    }]
  }

  if (currentTab.value === 'groups') {
    const roles = selectedItem.value.roles || []
    return roles.map((role) => ({
      left: selectedItem.value.group_name,
      middle: 'RoleBinding / ClusterRoleBinding',
      right: role
    }))
  }

  return [{
    left: selectedItem.value.name,
    middle: selectedItem.value.type,
    right: selectedItem.value.target_selector || 'All Endpoints'
  }]
})

watch(filteredData, (newList) => {
  if (!selectedItem.value || !isResourceTab.value) return

  let currentId
  if (currentTab.value === 'sa') {
    currentId = `${selectedItem.value.sa}-${selectedItem.value.namespace}`
  } else if (currentTab.value === 'groups') {
    currentId = selectedItem.value.group_name
  } else {
    currentId = `${selectedItem.value.name}-${selectedItem.value.namespace}`
  }

  const exists = newList.some((item) => {
    if (currentTab.value === 'sa') return `${item.sa}-${item.namespace}` === currentId
    if (currentTab.value === 'groups') return item.group_name === currentId
    return `${item.name}-${item.namespace}` === currentId
  })

  if (!exists) selectedItem.value = null
})

watch(namespaceFilter, () => {
  searchQuery.value = ''
})

watch(namespaces, (newNamespaces) => {
  if (namespaceFilter.value && !newNamespaces.includes(namespaceFilter.value)) {
    namespaceFilter.value = ''
  }
})

onMounted(() => {
  if (currentTab.value !== 'iam-security' && isResourceTab.value) {
    fetchData(false)
  }
})
</script>

<template>
  <div class="app-shell !flex-row">
    <aside class="w-[260px] shrink-0 border-r border-white/10 bg-[#0b1326] p-4 flex flex-col">
      <div class="header-brand mb-5">
        <ShieldCheck class="text-blue-400" :size="22" /> FleetOps
      </div>

      <div class="space-y-2">
        <button
          @click="handleTab('iam-security')"
          :class="currentTab === 'iam-security' ? 'tab-button tab-button-active' : 'tab-button tab-button-inactive'"
          class="w-full justify-start"
        >
          <Fingerprint :size="14" class="mr-2" /> IAM Navigator
        </button>

        <div v-if="areK8sResourcesVisible" class="rounded-xl border border-white/10 overflow-hidden">
          <button
            @click="isK8sResourcesExpanded = !isK8sResourcesExpanded"
            class="w-full flex items-center justify-between px-3 py-2 text-[11px] font-black uppercase tracking-wider text-slate-300 bg-slate-900/40"
          >
            <span>K8s Resources</span>
            <component :is="isK8sResourcesExpanded ? ChevronDown : ChevronRight" :size="14" />
          </button>

          <div v-if="isK8sResourcesExpanded" class="p-2 space-y-1 bg-slate-950/30">
            <button
              v-for="tab in enabledResourceTabs"
              :key="tab.key"
              @click="handleTab(tab.key)"
              :class="currentTab === tab.key ? 'tab-button tab-button-active' : 'tab-button tab-button-inactive'"
              class="w-full justify-start !py-2 !px-3"
            >
              <Box v-if="tab.key === 'sa'" :size="14" class="mr-2" />
              <Users v-else-if="tab.key === 'groups'" :size="14" class="mr-2" />
              <Network v-else :size="14" class="mr-2" />
              {{ tab.label }}
            </button>
          </div>
        </div>
      </div>
    </aside>

    <div class="flex-1 min-w-0 flex flex-col">
      <header class="app-header">
        <div class="text-sm font-black uppercase tracking-wider text-slate-300">
          {{ currentTabLabel }}
        </div>
        <div class="flex items-center gap-3">
          <span v-if="!isIAMTab" class="text-[10px] uppercase tracking-widest text-slate-500 font-bold">Last Sync: {{ lastUpdated }}</span>
          <button
            v-if="!isIAMTab"
            @click="k8sService.downloadExport(filteredData, currentTab)"
            class="p-2 hover:bg-white/5 rounded-full text-slate-400"
          >
            <Download :size="18" />
          </button>
        </div>
      </header>

      <IAMSecurityFeature v-if="isIAMTab" class="flex-1 min-h-0" />

      <template v-else>
        <div class="workspace-toolbar">
          <div class="relative flex-1 max-w-xl">
            <Search class="absolute left-4 top-3 text-slate-500" :size="18" />
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Filter resources..."
              class="w-full bg-slate-900/40 border border-white/10 rounded-2xl pl-12 pr-4 py-3 text-sm text-slate-200 outline-none"
            />
          </div>

          <select v-model="namespaceFilter">
            <option value="">{{ isLoading ? 'Loading...' : 'All Namespaces' }}</option>
            <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
          </select>

          <button
            @click="refreshData"
            :disabled="isLoading"
            class="flex items-center gap-2 px-6 py-3 bg-blue-500/10 border border-blue-500/20 rounded-2xl text-blue-400 text-xs font-black hover:bg-blue-500/20 transition-all disabled:opacity-50"
          >
            <RefreshCw :size="14" :class="{ 'animate-spin': isLoading }" />
            {{ isLoading ? 'SYNCING...' : 'REFRESH' }}
          </button>
        </div>

        <main class="workspace-main">
          <div class="workspace-split">
            <section class="workspace-list-pane custom-scroll p-4 space-y-3 relative">
            <div v-if="isLoading" class="loading-overlay">
              <div class="spinner"></div>
            </div>

            <IdentityCard
              v-for="item in filteredData"
              :key="`${currentTab}-${item.sa || item.group_name || item.name}-${item.namespace || 'global'}-${item.binding_name || ''}`"
              :item="item"
              :type="currentTab"
              :isSelected="selectedItem === item"
              @select="selectedItem = item"
            />
            </section>

            <section class="workspace-detail-pane">
            <div v-if="!selectedItem" class="h-full flex flex-col items-center justify-center text-slate-500">
              <div class="p-6 rounded-full bg-slate-800/20 mb-4 border border-white/5">
                <Fingerprint :size="34" />
              </div>
              <p class="text-[10px] font-black tracking-[0.3em] uppercase text-center">
                Select A Resource<br />To Inspect
              </p>
            </div>

            <div v-else class="h-full flex flex-col">
              <div class="p-6 border-b border-white/10 bg-slate-900/40">
                <h2 class="text-xl font-bold text-white tracking-tight">{{ selectedTitle }}</h2>
                <p class="text-[10px] text-slate-500 uppercase tracking-[0.2em] mt-1">{{ selectedSubtitle }}</p>
              </div>

              <div class="p-5 border-b border-white/10">
                <div class="insight-card">
                <div class="text-[10px] font-black tracking-widest uppercase text-indigo-300 mb-2">AI Security Insight</div>
                <p class="text-sm text-slate-200">{{ aiInsight }}</p>
                </div>
              </div>

              <div class="px-5 pt-4 flex gap-2 border-b border-white/10">
                <button
                  @click="detailTab = 'map'"
                  :class="detailTab === 'map' ? 'detail-tab-btn text-blue-300 border-b-2 border-blue-400' : 'detail-tab-btn text-slate-400'"
                >
                  Visual Map
                </button>
                <button
                  @click="detailTab = 'yaml'"
                  :class="detailTab === 'yaml' ? 'detail-tab-btn text-blue-300 border-b-2 border-blue-400' : 'detail-tab-btn text-slate-400'"
                >
                  YAML Inspector
                </button>
              </div>

              <div class="flex-1 min-h-0 overflow-y-auto custom-scroll p-5">
                <div v-if="detailTab === 'map'" class="space-y-3">
                  <div
                    v-for="(row, index) in mapRows"
                    :key="index"
                    class="workspace-card text-xs text-slate-300"
                  >
                    <span class="font-mono text-slate-100">{{ row.left }}</span>
                    <span class="mx-2 text-slate-500">-></span>
                    <span>{{ row.middle }}</span>
                    <span class="mx-2 text-slate-500">-></span>
                    <span class="text-orange-300">{{ row.right }}</span>
                  </div>
                </div>
                <YamlInspector v-else :item="selectedItem" :isLoading="isLoading" />
              </div>
            </div>
            </section>
          </div>
        </main>
      </template>
    </div>
  </div>
</template>
