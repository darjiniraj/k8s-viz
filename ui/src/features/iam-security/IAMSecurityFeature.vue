<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { Sparkles, ShieldAlert } from 'lucide-vue-next'
import YamlInspector from '../../components/YamlInspector.vue'
import GlobalSelectionBar from './components/GlobalSelectionBar.vue'
import IdentityFlowMap from './components/IdentityFlowMap.vue'
import { iamSecurityStore } from './iamSecurityStore'
import type { AttachmentType, IAMRBACDetail, IAMSecurityPrincipalGroup, IAMSecurityRecord } from './types'

type DetailTab = 'map' | 'yaml'

const iamRecords = computed<IAMSecurityRecord[]>(() => iamSecurityStore.records.value)
const isLoading = computed(() => iamSecurityStore.isLoading.value)
const lastSyncedAt = computed(() => iamSecurityStore.lastSyncedAt.value)
const searchQuery = ref('')
const selectedNamespace = ref('all')
const selectedRoleKeys = ref<string[]>([])
const selectedRoleKey = ref('')
const selectedSubject = ref('')
const detailTab = ref<DetailTab>('map')

const getRoleKey = (arn: string): string => {
  const chunks = arn.split('/')
  const raw = chunks[chunks.length - 1] || arn
  return raw.trim().toLowerCase()
}

const truncateArn = (arn: string): string => {
  if (arn.length <= 72) return arn
  return `${arn.slice(0, 38)}...${arn.slice(-28)}`
}

const attachmentClass = (type: AttachmentType): string => {
  if (type === 'IRSA') return 'bg-emerald-500/20 text-emerald-300 border-emerald-500/40'
  if (type === 'PodIdentity') return 'bg-sky-500/20 text-sky-300 border-sky-500/40'
  return 'bg-violet-500/20 text-violet-300 border-violet-500/40'
}

const safeRBACDetails = (record: IAMSecurityRecord): IAMRBACDetail[] => {
  return Array.isArray(record?.rbac_details) ? record.rbac_details : []
}

const detailNamespace = (detail: IAMRBACDetail): string => {
  return detail.binding_namespace || detail.role_namespace || 'Cluster-Wide'
}

const buildGroups = (records: IAMSecurityRecord[]): IAMSecurityPrincipalGroup[] => {
  const grouped = new Map<string, IAMSecurityPrincipalGroup>()

  for (const record of records) {
    const existing = grouped.get(record.iam_principal)
    if (!existing) {
      grouped.set(record.iam_principal, {
        iamPrincipal: record.iam_principal,
        roleKey: getRoleKey(record.iam_principal),
        attachments: [record.attachment_type],
        records: [record],
      })
      continue
    }

    existing.records.push(record)
    if (!existing.attachments.includes(record.attachment_type)) {
      existing.attachments.push(record.attachment_type)
    }
  }

  return Array.from(grouped.values()).sort((a, b) => a.roleKey.localeCompare(b.roleKey))
}

const namespaceOptions = computed<string[]>(() => {
  const namespaces = new Set<string>()
  for (const record of iamRecords.value) {
    for (const detail of safeRBACDetails(record)) {
      namespaces.add(detailNamespace(detail))
    }
  }
  return Array.from(namespaces.values()).sort()
})

const recordsAfterNamespaceAndSearch = computed<IAMSecurityRecord[]>(() => {
  const query = searchQuery.value.trim().toLowerCase()

  return iamRecords.value.filter((record) => {
    const details = safeRBACDetails(record)
    const matchesNamespace =
      selectedNamespace.value === 'all' ||
      details.some((detail) => detailNamespace(detail) === selectedNamespace.value)

    if (!matchesNamespace) return false
    if (!query) return true
    if (record.iam_principal.toLowerCase().includes(query)) return true
    if (getRoleKey(record.iam_principal).includes(query)) return true
    return record.k8s_subject.toLowerCase().includes(query)
  })
})

const roleOptions = computed(() =>
  buildGroups(recordsAfterNamespaceAndSearch.value).map((group) => ({
    iamPrincipal: group.iamPrincipal,
    roleKey: group.roleKey,
  })),
)

const recordsAfterGlobalFilters = computed<IAMSecurityRecord[]>(() => {
  if (selectedRoleKeys.value.length === 0) return recordsAfterNamespaceAndSearch.value
  const selected = new Set(selectedRoleKeys.value)
  return recordsAfterNamespaceAndSearch.value.filter((record) => selected.has(getRoleKey(record.iam_principal)))
})

const filteredPrincipals = computed<IAMSecurityPrincipalGroup[]>(() => buildGroups(recordsAfterGlobalFilters.value))

const selectedGroup = computed<IAMSecurityPrincipalGroup | null>(() => {
  if (filteredPrincipals.value.length === 0) return null
  const found = filteredPrincipals.value.find((group) => group.roleKey === selectedRoleKey.value)
  return found || filteredPrincipals.value[0]
})

const uniqueSubjects = computed<string[]>(() => {
  if (!selectedGroup.value) return []
  const seen = new Set<string>()
  for (const record of selectedGroup.value.records) {
    seen.add(record.k8s_subject)
  }
  return Array.from(seen.values()).sort()
})

const selectedRecords = computed<IAMSecurityRecord[]>(() => {
  if (!selectedGroup.value) return []
  if (!selectedSubject.value) return selectedGroup.value.records
  return selectedGroup.value.records.filter((record) => record.k8s_subject === selectedSubject.value)
})

const selectedSummary = computed<string>(() => {
  const candidate = selectedRecords.value.find((record) => record.summary_placeholder?.trim().length > 0)
  return candidate?.summary_placeholder || 'Summary will be generated by AI in a future release.'
})

const inspectorItem = computed(() => {
  if (!selectedGroup.value) return null

  const allYamls: Array<{ kind: string; name: string; data: string; namespace: string }> = []

  for (const record of selectedRecords.value) {
    for (const detail of safeRBACDetails(record)) {
      allYamls.push({
        kind: detail.binding_kind,
        name: detail.binding_name,
        data: detail.binding_yaml,
        namespace: detail.binding_namespace || 'Cluster-Wide',
      })
      allYamls.push({
        kind: detail.role_kind,
        name: detail.role_name,
        data: detail.role_yaml,
        namespace: detail.role_namespace || 'Cluster-Wide',
      })
    }
  }

  return {
    name: selectedGroup.value.iamPrincipal,
    all_yamls: allYamls,
  }
})

const updateURL = (roleKey: string) => {
  if (!roleKey) return
  const url = new URL(window.location.href)
  url.pathname = '/iam-security'
  url.searchParams.set('role', roleKey)
  window.history.replaceState({}, '', `${url.pathname}${url.search}`)
}

const parseRoleFromURL = (): string => {
  const url = new URL(window.location.href)
  return (url.searchParams.get('role') || '').trim().toLowerCase()
}

const selectPrincipal = (group: IAMSecurityPrincipalGroup) => {
  selectedRoleKey.value = group.roleKey
}

const rbacLinkCount = (group: IAMSecurityPrincipalGroup): number => {
  return group.records.reduce((total, record) => total + safeRBACDetails(record).length, 0)
}

watch(selectedGroup, (group) => {
  if (!group) return
  selectedRoleKey.value = group.roleKey
  if (!uniqueSubjects.value.includes(selectedSubject.value)) {
    selectedSubject.value = uniqueSubjects.value[0] || ''
  }
  updateURL(group.roleKey)
})

watch(filteredPrincipals, (groups) => {
  if (groups.length === 0) {
    selectedRoleKey.value = ''
    selectedSubject.value = ''
    return
  }
  const stillExists = groups.some((group) => group.roleKey === selectedRoleKey.value)
  if (!stillExists) selectedRoleKey.value = groups[0].roleKey
})

watch(roleOptions, (options) => {
  if (options.length === 0) {
    selectedRoleKeys.value = []
    return
  }
  const allowed = new Set(options.map((option) => option.roleKey))
  selectedRoleKeys.value = selectedRoleKeys.value.filter((key) => allowed.has(key))
})

onMounted(async () => {
  selectedRoleKey.value = parseRoleFromURL()
  await iamSecurityStore.ensureLoaded()
})
</script>

<template>
  <div class="h-full flex flex-col">
    <GlobalSelectionBar
      :search-query="searchQuery"
      :selected-namespace="selectedNamespace"
      :namespace-options="namespaceOptions"
      :role-options="roleOptions"
      :selected-role-keys="selectedRoleKeys"
      :is-syncing="isLoading"
      :last-synced-at="lastSyncedAt"
      @update:searchQuery="searchQuery = $event"
      @update:selectedNamespace="selectedNamespace = $event"
      @update:selectedRoleKeys="selectedRoleKeys = $event"
      @refresh="iamSecurityStore.refresh"
    />

    <main class="workspace-main">
      <div class="workspace-split">
      <section class="workspace-list-pane custom-scroll p-4">
        <div v-if="isLoading" class="space-y-3">
          <div v-for="idx in 7" :key="idx" class="workspace-card animate-pulse">
            <div class="h-3 bg-slate-700/50 rounded w-3/4 mb-3"></div>
            <div class="h-2 bg-slate-800/80 rounded w-1/2"></div>
          </div>
        </div>

        <template v-else>
          <button
            v-for="group in filteredPrincipals"
            :key="group.iamPrincipal"
            @click="selectPrincipal(group)"
            :class="selectedGroup?.iamPrincipal === group.iamPrincipal
              ? 'border-blue-500/50 bg-blue-500/10 ring-1 ring-blue-500/20'
              : 'border-white/10 bg-slate-900/25 hover:bg-slate-900/45'"
            class="w-full text-left p-4 rounded-2xl border transition-all mb-3"
          >
            <p class="font-mono text-[11px] text-slate-100 truncate" :title="group.iamPrincipal">
              {{ truncateArn(group.iamPrincipal) }}
            </p>
            <p class="text-[10px] text-slate-500 mt-1">
              {{ group.records.length }} subjects, {{ rbacLinkCount(group) }} bindings
            </p>
            <div class="mt-3 flex flex-wrap gap-2">
              <span
                v-for="attachment in group.attachments"
                :key="attachment"
                :class="attachmentClass(attachment)"
                class="text-[9px] px-2 py-1 rounded-full border font-black tracking-wide uppercase"
              >
                {{ attachment }}
              </span>
            </div>
          </button>

          <div v-if="filteredPrincipals.length === 0" class="text-center text-slate-500 pt-20">
            No IAM identities matched your filter.
          </div>
        </template>
      </section>

      <section class="workspace-detail-pane overflow-y-auto custom-scroll">
        <div v-if="!selectedGroup && !isLoading" class="h-full flex flex-col items-center justify-center text-slate-500">
          <div class="p-6 rounded-full bg-slate-800/30 border border-white/10 mb-4">
            <ShieldAlert :size="36" />
          </div>
          <p class="text-sm">Select an IAM identity to begin</p>
        </div>

        <div v-else class="space-y-5 p-5">
          <header class="workspace-card">
            <h2 class="text-lg text-white font-bold tracking-tight break-all">
              {{ selectedGroup?.iamPrincipal }} -> {{ selectedSubject || 'K8s Subject' }}
            </h2>
            <div class="mt-3 flex flex-wrap gap-2">
              <button
                v-for="subject in uniqueSubjects"
                :key="subject"
                @click="selectedSubject = subject"
                :class="selectedSubject === subject
                  ? 'chip-btn bg-blue-500/20 border-blue-500/40 text-blue-300'
                  : 'chip-btn bg-slate-900/40 border-white/10 text-slate-300'"
              >
                {{ subject }}
              </button>
            </div>
          </header>

          <div class="insight-card">
            <div class="flex items-center gap-2 text-[10px] font-black tracking-widest uppercase text-indigo-200 mb-2">
              <Sparkles :size="12" /> AI Security Insight
            </div>
            <p class="text-sm text-slate-200 leading-relaxed">
              {{ selectedSummary }}
            </p>
          </div>

          <div class="rounded-2xl border border-white/10 bg-slate-900/30 overflow-hidden">
            <div class="flex border-b border-white/10 bg-slate-900/40">
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

            <div class="p-4 bg-slate-900/20">
              <IdentityFlowMap v-if="detailTab === 'map'" :records="selectedRecords" />

              <YamlInspector v-else :item="inspectorItem" :isLoading="isLoading" />
            </div>
          </div>
        </div>
      </section>
      </div>
    </main>
  </div>
</template>
