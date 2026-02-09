<script setup>
import { ref, onMounted, computed } from 'vue'
import { Search, Download, ShieldCheck, Box, Users } from 'lucide-vue-next'
import { k8sService } from './services/k8sService'
import IdentityCard from './components/IdentityCard.vue'
import YamlInspector from './components/YamlInspector.vue'
import { watch } from 'vue';
import { Network } from 'lucide-vue-next'



const currentTab = ref('sa')
const saData = ref([])
const groupData = ref([])
const ciliumData = ref([])
const selectedItem = ref(null)
const searchQuery = ref('')
const namespaceFilter = ref('')
const isLoading = ref(false)
const lastUpdated = ref(new Date().toLocaleTimeString())

const handleTab = (tab) => {
  if (currentTab.value === tab) return;

  // 1. Reset selection and filters
  currentTab.value = tab;
  selectedItem.value = null;
  searchQuery.value = '';


  if (!namespaces.value.includes(namespaceFilter.value)) {
    namespaceFilter.value = '';
  }

  groupData.value = [];
  saData.value = [];
  ciliumData.value = [];

  fetchData(false);
};

const namespaces = computed(() => {
  let data = []
  if (currentTab.value === 'sa') data = saData.value
  else if (currentTab.value === 'groups') data = groupData.value
  else data = ciliumData.value // For Cilium

  if (!Array.isArray(data)) return []

  const set = new Set()
  for (const item of data) {
    if (item.namespace) set.add(item.namespace)

    // Groups
    if (Array.isArray(item.namespaces)) {
      item.namespaces.forEach(ns => { if (ns) set.add(ns) })
    }
  }

  if (currentTab.value == 'cilium' && data.some(i => i.is_cluster_wide)) {
    set.add('Global')
  }
  return [...set].sort()
})


const fetchData = async (refresh = false) => {
  isLoading.value = true;
  try {
    const rawData = await k8sService.fetchTableData(currentTab.value, refresh);

    if (currentTab.value === 'sa') {
      // Group SAs by name + namespace so they don't appear as duplicates
      const grouped = rawData.reduce((acc, curr) => {
        const key = `${curr.sa}-${curr.namespace}`;
        if (!acc[key]) {
          acc[key] = { ...curr, roles: [curr.role], bindings: [curr.binding_name] };
        } else {
          if (!acc[key].roles.includes(curr.role)) acc[key].roles.push(curr.role);
          if (!acc[key].bindings.includes(curr.binding_name)) acc[key].bindings.push(curr.binding_name);
        }
        return acc;
      }, {});
      saData.value = Object.values(grouped);
    } else if (currentTab.value === 'groups') {
      groupData.value = [...rawData];
    } else {
      ciliumData.value = [...rawData];
    }

    lastUpdated.value = new Date().toLocaleTimeString();
  } catch (err) {
    console.error("Fetch failed:", err);
  } finally {
    isLoading.value = false;
  }
};

// ADD THIS FUNCTION
const refreshData = () => {
  fetchData(true)
}



const filteredData = computed(() => {
  const query = searchQuery.value.toLowerCase().trim()
  const ns = namespaceFilter.value

  // SELECT SOURCE BASED ON TAB
  let data = []
  if (currentTab.value === 'sa') data = saData.value
  else if (currentTab.value === 'groups') data = groupData.value
  else if (currentTab.value === 'cilium') data = ciliumData.value
  else return []

  return data.filter(item => {
    // 1. NAMESPACE GATE
    let matchesNS = false;
    if (currentTab.value === 'sa') {
      matchesNS = !ns || item.namespace === ns;
    } else if (currentTab.value === 'groups') {
      matchesNS = !ns || (item.namespaces && item.namespaces.includes(ns));
    } else if (currentTab.value === 'cilium') {
      if (!ns) {
        matchesNS = true
      } else if (ns === 'Global') {
        matchesNS = item.is_cluster_wide
      } else {
        matchesNS = item.namespace === ns && !item.is_cluster_wide
      }

    }

    if (!matchesNS) return false;

    // 2. SEARCH GATE
    if (!query) return true;

    const searchStr = currentTab.value === 'cilium'
      ? `${item.name} ${item.target_selector}`
      : currentTab.value === 'sa'
        ? `${item.sa} ${item.role_name} ${item.iam_role}`
        : `${item.group_name} ${item.roles?.join(' ')}`;

    return searchStr.toLowerCase().includes(query);
  })
})

// 1. Clear selection if filtered out
watch(filteredData, (newList) => {
  if (!selectedItem.value) return


  let currentId
  if (currentTab.value === 'sa') {
    currentId = `${selectedItem.value.sa}-${selectedItem.value.namespace}`
  } else if (currentTab.value === 'groups') {
    currentId = selectedItem.value.group_name
  } else {
    currentId = `${selectedItem.value.name}-${selectedItem.value.namespace}`
  }

  const exists = newList.some(item => {
    if (currentTab.value === 'sa') {
      return `${item.sa}-${item.namespace}` === currentId
    } else if (currentTab.value === 'groups') {
      return item.group_name === currentId
    } else {
      return `${item.name}-${item.namespace}` === currentId
    }
  })

  if (!exists) {
    selectedItem.value = null
  }
})

// 2. UX: reset search when namespace changes
watch(namespaceFilter, () => {
  searchQuery.value = ''
})

// 3. CRITICAL: reset namespace if option disappears
watch(namespaces, (newNamespaces) => {
  if (
    namespaceFilter.value &&
    !newNamespaces.includes(namespaceFilter.value)
  ) {
    namespaceFilter.value = ''
  }
})



watch(namespaceFilter, () => {
  searchQuery.value = '';
  if (currentTab.value === 'cilium') {
    fetchData(false)
  }
})


onMounted(() => {
  fetchData(false)
})

</script>


<template>
  <div class="app-shell">
    <header class="app-header">
      <div class="flex items-center gap-10">
        <div class="header-brand">
          <ShieldCheck class="text-blue-400" :size="24" /> FleetOps
        </div>
        <nav class="tab-nav">
          <button @click="handleTab('sa')"
            :class="currentTab === 'sa' ? 'tab-button tab-button-active' : 'tab-button tab-button-inactive'">
            <Box :size="14" class="inline mr-2" /> Service Accounts
          </button>
          <button @click="handleTab('groups')"
            :class="currentTab === 'groups' ? 'tab-button tab-button-active' : 'tab-button tab-button-inactive'">
            <Users :size="14" class="inline mr-2" /> User Groups
          </button>
          <button @click="handleTab('cilium')"
            :class="currentTab === 'cilium' ? 'tab-button tab-button-active' : 'tab-button tab-button-inactive'">
            <Network :size="14" class="inline mr-2" /> Cilium Policies
          </button>
        </nav>
      </div>
      <div class="flex items-center gap-4">
        <span class="text-[10px] uppercase tracking-widest text-slate-500 font-bold">MVP 1: Local Cluster</span>
        <button @click="k8sService.downloadExport(filteredData, currentTab)"
          class="p-2 hover:bg-white/5 rounded-full text-slate-400">
          <Download :size="20" />
        </button>
      </div>
    </header>

    <div class="px-8 py-6 flex items-center gap-4 bg-gradient-to-b from-[#1e293b]/30 to-transparent">
      <div class="relative flex-1 max-w-xl">
        <Search class="absolute left-4 top-3 text-slate-500" :size="18" />
        <input v-model="searchQuery" type="text" placeholder="Filter..."
          class="w-full bg-slate-900/40 border border-white/10 rounded-2xl pl-12 pr-4 py-3 text-sm text-slate-200 outline-none" />
      </div>

      <select v-model="namespaceFilter">
        <option value="">{{ isLoading ? 'Loading...' : 'All Namespaces' }}</option>
        <option v-for="ns in namespaces" :key="ns" :value="ns">
          {{ ns }}
        </option>
      </select>

      <button @click="refreshData" :disabled="isLoading"
        class="flex items-center gap-2 px-6 py-3 bg-blue-500/10 border border-blue-500/20 rounded-2xl text-blue-400 text-xs font-black hover:bg-blue-500/20 transition-all disabled:opacity-50">
        <RefreshCw :size="14" :class="{ 'animate-spin': isLoading }" />
        {{ isLoading ? 'SYNCING...' : 'REFRESH' }}
      </button>
    </div>

    <div class="px-8 pb-2 text-[9px] text-slate-500 uppercase tracking-widest text-right">
      Last Sync: {{ lastUpdated }}
    </div>

    <main class="main-layout flex flex-row">
      <section class="min-w-[550px] w-[35%] overflow-y-auto space-y-3 custom-scroll relative">
        <div v-if="isLoading" class="loading-overlay">
          <div class="spinner"></div>
        </div>


        <!-- <IdentityCard v-for="item in filteredData"
          :key="`${currentTab}-${item.sa || item.group_name}-${item.namespace || 'global'}-${item.binding_name || ''}`"
          :item="item" :type="currentTab" :isSelected="selectedItem === item" @select="selectedItem = item" />  -->

          <IdentityCard v-for="item in filteredData"
          :key="`${currentTab}-${item.sa || item.group_name || item.name}-${item.namespace || 'global'}-${item.binding_name || ''}`"
          :item="item" :type="currentTab" :isSelected="selectedItem === item" @select="selectedItem = item" />
      </section>

      <YamlInspector :item="selectedItem" :isLoading="isLoading" class="flex-1" />
    </main>
  </div>
</template>