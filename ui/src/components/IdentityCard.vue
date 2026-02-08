<!-- <script setup>
import { computed } from 'vue'
import { Box, Users, ChevronRight, ArrowRight, Globe, Shield } from 'lucide-vue-next'

const props = defineProps(['item', 'type', 'isSelected'])
defineEmits(['select'])

const name = computed(() => props.item.sa || props.item.group_name)

// Helper to pair up the items for Group view
const yamlPairs = computed(() => {
  if (!props.item.all_yamls) return []
  const pairs = []
  for (let i = 0; i < props.item.all_yamls.length; i += 2) {
    pairs.push([props.item.all_yamls[i], props.item.all_yamls[i + 1]])
  }
  return pairs
})
</script>

<template>
  <div @click="$emit('select')"
       :class="isSelected ? 'border-blue-500/50 bg-blue-500/10 ring-1 ring-blue-500/20 shadow-blue-500/10' : 'border-white/5 bg-[#1e293b]/40 hover:bg-[#1e293b]/60 shadow-sm'"
       class="p-5 rounded-2xl border transition-all cursor-pointer group relative overflow-hidden">
    
    <div v-if="isSelected" class="absolute inset-0 bg-gradient-to-br from-blue-500/5 to-transparent pointer-events-none"></div>

    <div class="flex justify-between items-start mb-4 relative z-10">
      <div class="flex items-center gap-3">
        <div :class="type === 'sa' ? 'bg-blue-500/20 text-blue-400' : 'bg-amber-500/20 text-amber-400'" class="p-2 rounded-lg">
          <component :is="type === 'sa' ? Box : Users" :size="18" />
        </div>
        <div>
          <h3 class="font-semibold text-slate-100 text-base flex items-center gap-2">
            {{ name }}
          </h3>
          <p v-if="type === 'sa'" class="text-[10px] text-slate-500 font-mono mt-0.5">{{ item.iam_role }}</p>
        </div>
      </div>
      <ChevronRight :size="16" class="text-slate-600 group-hover:translate-x-1 transition-transform" />
    </div>

    <div class="flex flex-wrap gap-2 relative z-10">
      
      <template v-if="type === 'sa'">
        <div class="flex items-center gap-2 bg-slate-950/60 px-3 py-2 rounded-xl border border-white/5">
          
          <div class="flex items-center gap-1.5 mr-1">
            <template v-if="item.binding_type === 'ClusterRoleBinding'">
              <Globe :size="12" class="text-purple-400" />
              <span class="text-[10px] font-black text-purple-400 uppercase tracking-tighter">Cluster Access</span>
            </template>
            <template v-else>
              <Shield :size="12" class="text-emerald-500" />
              <span class="text-[10px] font-bold text-emerald-500 uppercase">{{ item.namespace }}</span>
            </template>
          </div>

          <ArrowRight :size="10" class="text-slate-600" />
          
          <span class="text-[10px] font-mono text-blue-400/80 truncate max-w-[140px]">{{ item.binding_name }}</span>
          
          <ArrowRight :size="10" class="text-slate-600" />
          
          <span class="text-[10px] font-bold text-amber-500">
            <span class="opacity-40 font-normal mr-1 text-[9px]">{{ item.role_kind }}</span>
            {{ item.role_name || item.role }}
          </span>
        </div>
      </template>

      <template v-else>
        <div v-for="(pair, idx) in yamlPairs" :key="idx" 
             class="flex items-center gap-2 bg-slate-950/60 px-3 py-2 rounded-xl border border-white/5 mb-1">
          
          <div class="flex items-center gap-1.5 mr-1">
            <Globe v-if="pair[0].kind === 'ClusterRoleBinding'" :size="12" class="text-purple-400" />
            <Shield v-else :size="12" class="text-emerald-500" />
            <span class="text-[10px] font-bold uppercase" :class="pair[0].kind === 'ClusterRoleBinding' ? 'text-purple-400' : 'text-emerald-500'">
              {{ pair[0].kind === 'ClusterRoleBinding' ? 'Cluster' : pair[0].namespace }}
            </span>
          </div>

          <ArrowRight :size="10" class="text-slate-600" />
          
          <span class="text-[10px] font-mono text-slate-400">{{ pair[0].name }}</span>
          
          <ArrowRight :size="10" class="text-slate-600" />
          
          <span class="text-[10px] font-bold text-amber-500">
            <span class="opacity-40 font-normal mr-1 text-[9px]">{{ pair[1]?.kind || 'Role' }}</span>
            {{ pair[1]?.name || 'Role' }}
          </span>
        </div>
        <div v-if="yamlPairs.length === 0" class="text-[10px] text-slate-500 italic px-2">No mappings found</div>
      </template>
    </div>
  </div>
</template> -->

<script setup>
import { computed } from 'vue'
import { Box, Users, ChevronRight, ArrowRight, Globe, Home, ShieldCheck, Link } from 'lucide-vue-next'

const props = defineProps(['item', 'type', 'isSelected'])
defineEmits(['select'])

const name = computed(() => props.item.sa || props.item.group_name)

const yamlPairs = computed(() => {
  if (!props.item.all_yamls) return []
  const pairs = []
  for (let i = 0; i < props.item.all_yamls.length; i += 2) {
    pairs.push({ binding: props.item.all_yamls[i], role: props.item.all_yamls[i + 1] })
  }
  return pairs
})
</script>

<template>
  <div @click="$emit('select')"
       :class="isSelected ? 'border-blue-500/50 bg-blue-500/10 ring-1 ring-blue-500/20 shadow-xl' : 'border-white/5 bg-[#1e293b]/40 hover:bg-[#1e293b]/60'"
       class="p-5 rounded-2xl border transition-all cursor-pointer group relative overflow-hidden mb-3">
    
    <div class="flex justify-between items-center mb-4 relative z-10">
      <div class="flex items-center gap-3">
        <div :class="type === 'sa' ? 'bg-blue-500/20 text-blue-400' : 'bg-amber-500/20 text-amber-400'" class="p-2 rounded-xl">
          <component :is="type === 'sa' ? Box : Users" :size="18" />
        </div>
        <h3 class="font-bold text-slate-100 tracking-tight text-base">{{ name }}</h3>
      </div>
      <ChevronRight :size="16" class="text-slate-600 group-hover:translate-x-1 transition-transform" />
    </div>

    <div class="space-y-2 relative z-10">
      <template v-for="(pair, idx) in (type === 'sa' ? [{binding: {kind: item.binding_type, name: item.binding_name, namespace: item.namespace}, role: {kind: item.role_kind, name: item.role_name || item.role}}] : yamlPairs)" :key="idx">
        <div class="flex items-stretch bg-slate-950/60 rounded-xl border border-white/5 overflow-hidden h-11">
          
          <div :class="pair.binding.kind === 'ClusterRoleBinding' ? 'bg-purple-500/10 border-purple-500/20' : 'bg-emerald-500/10 border-emerald-500/20'" 
               class="w-20 flex flex-col items-center justify-center border-r border-white/5 shrink-0">
             <component :is="pair.binding.kind === 'ClusterRoleBinding' ? Globe : ShieldCheck" :size="12" 
                        :class="pair.binding.kind === 'ClusterRoleBinding' ? 'text-purple-400' : 'text-emerald-400'" />
             <span class="text-[7px] font-black uppercase mt-0.5 truncate px-1 w-full text-center" 
                   :class="pair.binding.kind === 'ClusterRoleBinding' ? 'text-purple-400' : 'text-emerald-400'">
               {{ pair.binding.kind === 'ClusterRoleBinding' ? 'Cluster' : (pair.binding.namespace || 'Local') }}
             </span>
          </div>

          <div class="flex-1 min-w-0 flex flex-col justify-center px-3 border-r border-white/5">
            <span class="text-[7px] text-slate-500 font-bold uppercase tracking-widest leading-none mb-1">{{ pair.binding.kind }}</span>
            <span class="text-[10px] font-mono text-slate-300 truncate pr-4" :title="pair.binding.name">
              {{ pair.binding.name }}
            </span>
          </div>

          <div class="w-40 flex flex-col justify-center px-3 bg-slate-900/40 shrink-0">
            <span class="text-[7px] text-slate-500 font-bold uppercase tracking-widest leading-none mb-1">{{ pair.role?.kind || 'Role' }}</span>
            <span class="text-[10px] font-bold text-orange-400 truncate" :title="pair.role?.name || 'Role'">
              {{ pair.role?.name || 'Role' }}
            </span>
          </div>
          
        </div>
      </template>
      <div v-if="type !== 'sa' && yamlPairs.length === 0" class="text-[10px] text-slate-500 italic px-2">No mappings found</div>
    </div>
  </div>
</template>