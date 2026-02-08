<!-- <template>
  <div @click="$emit('select')"
       :class="isSelected ? 'border-blue-500/50 bg-blue-500/5 ring-1 ring-blue-500/20' : 'border-white/5 bg-[#1e293b]/40 hover:bg-[#1e293b]/60'"
       class="p-5 rounded-2xl border transition-all cursor-pointer group shadow-sm">
    
    <div class="flex justify-between items-start mb-4">
      <div class="flex items-center gap-3">
        <div :class="type === 'sa' ? 'bg-blue-500/20 text-blue-400' : 'bg-amber-500/20 text-amber-400'" class="p-2 rounded-lg">
          <component :is="type === 'sa' ? Box : Users" :size="18" />
        </div>
        <h3 class="font-semibold text-slate-100 text-base">{{ name }}</h3>
      </div>
      <ChevronRight :size="16" class="text-slate-600 group-hover:translate-x-1 transition-transform" />
    </div>

    <div class="flex flex-wrap gap-2">
      <template v-if="type === 'sa'">
        <div class="flex items-center gap-2 bg-slate-950/40 px-3 py-1.5 rounded-xl border border-white/5">
          <span class="text-[10px] font-bold text-emerald-500 uppercase">{{ item.namespace }}</span>
          <ArrowRight :size="10" class="text-slate-600" />
          <span class="text-[10px] font-mono text-blue-400 truncate max-w-[120px]">{{ item.binding_name }}</span>
          <ArrowRight :size="10" class="text-slate-600" />
          <span class="text-[10px] font-bold text-amber-500">{{ item.role }}</span>
        </div>
      </template>

      <template v-else>
        <div v-for="i in yamlsCount" :key="i" class="flex items-center gap-2 bg-slate-950/40 px-3 py-1.5 rounded-xl border border-white/5">
          <span class="text-[10px] font-bold text-emerald-500 uppercase">
            {{ item.all_yamls[(i-1)*2].kind === 'ClusterRoleBinding' ? 'Global' : item.all_yamls[(i-1)*2].namespace }}
          </span>
          <ArrowRight :size="10" class="text-slate-600" />
          <span class="text-[10px] font-mono text-slate-400">{{ item.all_yamls[(i-1)*2].name }}</span>
          <ArrowRight :size="10" class="text-slate-600" />
          <span class="text-[10px] font-bold text-amber-500">{{ item.all_yamls[(i-1)*2 + 1]?.name || 'Role' }}</span>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
import { Box, Users, ChevronRight, ArrowRight } from 'lucide-vue-next';
import { computed } from 'vue';

const props = defineProps(['item', 'type', 'isSelected']);
defineEmits(['select']);

const name = computed(() => props.item.sa || props.item.group_name);
const yamlsCount = computed(() => props.item.all_yamls ? Math.floor(props.item.all_yamls.length / 2) : 0);
</script> -->

<script setup>
import { computed } from 'vue'
import { Box, Users, ChevronRight, ArrowRight } from 'lucide-vue-next'

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
       :class="isSelected ? 'border-blue-500/50 bg-blue-500/5 ring-1 ring-blue-500/20' : 'border-white/5 bg-[#1e293b]/40 hover:bg-[#1e293b]/60'"
       class="p-5 rounded-2xl border transition-all cursor-pointer group shadow-sm">
    
    <div class="flex justify-between items-start mb-4">
      <div class="flex items-center gap-3">
        <div :class="type === 'sa' ? 'bg-blue-500/20 text-blue-400' : 'bg-amber-500/20 text-amber-400'" class="p-2 rounded-lg">
          <component :is="type === 'sa' ? Box : Users" :size="18" />
        </div>
        <h3 class="font-semibold text-slate-100 text-base">{{ name }}</h3>
      </div>
      <ChevronRight :size="16" class="text-slate-600 group-hover:translate-x-1 transition-transform" />
    </div>

    <div class="flex flex-wrap gap-2">
      <template v-if="type === 'sa'">
        <div class="flex items-center gap-2 bg-slate-950/40 px-3 py-1.5 rounded-xl border border-white/5">
          <span class="text-[10px] font-bold text-emerald-500 uppercase">{{ item.namespace }}</span>
          <ArrowRight :size="10" class="text-slate-600" />
          <span class="text-[10px] font-mono text-blue-400 truncate max-w-[120px]">{{ item.binding_name }}</span>
          <ArrowRight :size="10" class="text-slate-600" />
          <span class="text-[10px] font-bold text-amber-500">{{ item.role }}</span>
        </div>
      </template>

      <template v-else>
        <div v-for="(pair, idx) in yamlPairs" :key="idx" 
             class="flex items-center gap-2 bg-slate-950/40 px-3 py-1.5 rounded-xl border border-white/5">
          <span class="text-[10px] font-bold text-emerald-500 uppercase">
            {{ pair[0].kind === 'ClusterRoleBinding' ? 'Global' : pair[0].namespace }}
          </span>
          <ArrowRight :size="10" class="text-slate-600" />
          <span class="text-[10px] font-mono text-slate-400">{{ pair[0].name }}</span>
          <ArrowRight :size="10" class="text-slate-600" />
          <span class="text-[10px] font-bold text-amber-500">{{ pair[1]?.name || 'Role' }}</span>
        </div>
        <div v-if="yamlPairs.length === 0" class="text-[10px] text-slate-500 italic">No mappings found</div>
      </template>
    </div>
  </div>
</template>