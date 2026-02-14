<script setup lang="ts">
import { computed } from 'vue'
import { Search, Filter, Layers, RefreshCw } from 'lucide-vue-next'

interface RoleOption {
  iamPrincipal: string
  roleKey: string
}

const props = defineProps<{
  searchQuery: string
  selectedNamespace: string
  namespaceOptions: string[]
  roleOptions: RoleOption[]
  selectedRoleKeys: string[]
  isSyncing: boolean
  lastSyncedAt: string
}>()

const emit = defineEmits<{
  'update:searchQuery': [value: string]
  'update:selectedNamespace': [value: string]
  'update:selectedRoleKeys': [value: string[]]
  refresh: []
}>()

const selectedRoleSet = computed(() => new Set(props.selectedRoleKeys))

const updateRole = (roleKey: string, checked: boolean) => {
  const next = new Set(props.selectedRoleKeys)
  if (checked) next.add(roleKey)
  else next.delete(roleKey)
  emit('update:selectedRoleKeys', Array.from(next.values()))
}

const clearRoles = () => emit('update:selectedRoleKeys', [])
</script>

<template>
  <div class="workspace-toolbar flex-wrap">
    <div class="relative flex-1 min-w-[260px]">
      <Search class="absolute left-4 top-3 text-slate-500" :size="18" />
      <input
        :value="searchQuery"
        type="text"
        placeholder="Search IAM Role or K8s Subject..."
        class="w-full bg-slate-900/40 border border-white/10 rounded-2xl pl-12 pr-4 py-3 text-sm text-slate-200 outline-none focus:border-blue-500/50"
        @input="emit('update:searchQuery', ($event.target as HTMLInputElement).value)"
      />
    </div>

    <div class="flex items-center gap-2 min-w-[220px]">
      <Filter :size="16" class="text-blue-300" />
      <select
        :value="selectedNamespace"
        class="w-full !rounded-2xl"
        @change="emit('update:selectedNamespace', ($event.target as HTMLSelectElement).value)"
      >
        <option value="all">All Namespaces</option>
        <option v-for="namespace in namespaceOptions" :key="namespace" :value="namespace">
          {{ namespace }}
        </option>
      </select>
    </div>

    <details class="min-w-[260px] flex-1 rounded-2xl border border-white/10 bg-slate-900/35">
      <summary class="list-none cursor-pointer px-4 py-3 text-xs font-semibold text-slate-200 flex items-center justify-between">
        <span class="flex items-center gap-2">
          <Layers :size="15" class="text-indigo-300" />
          IAM Role Selector
        </span>
        <span class="text-[10px] text-slate-400">
          {{ selectedRoleKeys.length === 0 ? 'All Roles' : `${selectedRoleKeys.length} Selected` }}
        </span>
      </summary>
      <div class="px-4 pb-4 border-t border-white/10">
        <div class="pt-3 flex items-center justify-between">
          <span class="text-[10px] uppercase tracking-wider text-slate-400">
            Multi-select Roles
          </span>
          <button
            type="button"
            class="text-[10px] text-blue-300 hover:text-blue-200"
            @click="clearRoles"
          >
            Clear
          </button>
        </div>
        <div class="mt-3 max-h-44 overflow-y-auto custom-scroll space-y-2 pr-1">
          <label
            v-for="role in roleOptions"
            :key="role.iamPrincipal"
            class="flex items-start gap-2 rounded-lg border border-white/10 bg-slate-950/35 px-2 py-2"
          >
            <input
              :checked="selectedRoleSet.has(role.roleKey)"
              type="checkbox"
              class="mt-0.5 accent-blue-500"
              @change="updateRole(role.roleKey, ($event.target as HTMLInputElement).checked)"
            />
            <span class="text-[11px] text-slate-300 font-mono break-all">{{ role.iamPrincipal }}</span>
          </label>
          <div v-if="roleOptions.length === 0" class="text-[11px] text-slate-500 py-2">
            No IAM roles available for this namespace filter.
          </div>
        </div>
      </div>
    </details>

    <div class="flex items-center gap-2 ml-auto">
      <span v-if="lastSyncedAt" class="text-[10px] uppercase tracking-wider text-slate-500">
        Last Sync: {{ lastSyncedAt }}
      </span>
      <button
        type="button"
        class="inline-flex items-center gap-2 rounded-2xl border border-blue-500/35 bg-blue-500/15 px-4 py-2 text-xs font-black uppercase tracking-wide text-blue-200 hover:bg-blue-500/25 disabled:opacity-60"
        :disabled="isSyncing"
        @click="emit('refresh')"
      >
        <RefreshCw :size="14" :class="isSyncing ? 'animate-spin' : ''" />
        {{ isSyncing ? 'Syncing' : 'Sync' }}
      </button>
    </div>
  </div>
</template>
