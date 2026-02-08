<script setup>
import { ref } from 'vue' // Added missing ref import
import { ShieldAlert, Box, Users, Network, Copy, Check } from 'lucide-vue-next'

defineProps(['item', 'isLoading'])

const copiedIndex = ref(null)

const copyToClipboard = async (text, index) => {
    try {
        await navigator.clipboard.writeText(text)
        copiedIndex.value = index
        setTimeout(() => {
            copiedIndex.value = null
        }, 2000)
    } catch (err) {
        console.error('Failed to copy!', err)
    }
}
</script>

<template>
    <section
        class="bg-slate-900/40 border border-white/5 rounded-3xl overflow-hidden flex flex-col shadow-2xl relative">
        <div v-if="!item" class="flex flex-col items-center justify-center h-full text-slate-700 min-h-[400px]">
            <div class="p-6 rounded-full bg-slate-800/20 mb-4 border border-white/5">
                <ShieldAlert :size="40" stroke-width="1" />
            </div>
            <p class="text-[10px] font-black tracking-[0.3em] uppercase text-center">Select an Identity<br />to Audit
                Manifest</p>
        </div>

        <div v-else class="flex flex-col h-full animate-in fade-in duration-300">
            <div class="p-8 border-b border-white/5 flex justify-between items-center bg-slate-900/20">
                <div>
                    <h2 class="text-2xl font-bold text-white tracking-tight">{{ item.name || item.sa || item.group_name
                        }}</h2>
                    <p class="text-[10px] text-slate-500 uppercase tracking-[0.2em] mt-1">
                        {{ item.yaml ? 'Cilium Network Security' : 'RBAC Security' }} Manifest Data
                    </p>
                </div>

                <div :class="{
                    'bg-blue-500/10 text-blue-400 border-blue-500/20': item.sa,
                    'bg-amber-500/10 text-amber-400 border-amber-500/20': item.group_name,
                    'bg-indigo-500/10 text-indigo-400 border-indigo-500/20': item.yaml
                }" class="flex items-center gap-2 px-3 py-1 border rounded-full text-[9px] font-black uppercase tracking-widest">
                    <component :is="item.sa ? Box : (item.group_name ? Users : Network)" :size="10" />
                    {{ item.sa ? 'Service Account' : (item.group_name ? 'User Group' : item.kind) }}
                </div>
            </div>

            <div class="flex-1 overflow-y-auto p-8 space-y-8 custom-scroll">

                <div v-if="item.yaml" class="space-y-3">
                    <div class="flex items-center justify-between">
                        <div class="flex items-center gap-3">
                            <span
                                class="text-indigo-400 bg-indigo-400/10 text-[9px] px-2 py-0.5 rounded font-black uppercase border border-white/5">
                                {{ item.kind }}
                            </span>
                            <span class="text-xs font-mono text-slate-500">{{ item.namespace || 'Cluster-Wide' }}</span>
                        </div>
                        <button @click="copyToClipboard(item.yaml, 'cilium')" class="copy-btn">
                            <component :is="copiedIndex === 'cilium' ? Check : Copy" :size="12" />
                            <span>{{ copiedIndex === 'cilium' ? 'COPIED' : 'COPY' }}</span>
                        </button>
                    </div>
                    <div class="yaml-block">
                        <pre class="yaml-code custom-scroll text-indigo-300/80">{{ item.yaml }}</pre>
                    </div>
                </div>

                <template v-else-if="item.all_yamls && item.all_yamls.length">
                    <div v-for="(y, i) in item.all_yamls" :key="i" class="space-y-3">
                        <div class="flex items-center justify-between">
                            <div class="flex items-center gap-3">
                                <span
                                    class="text-blue-400 bg-blue-400/10 text-[9px] px-2 py-0.5 rounded font-black uppercase border border-white/5">
                                    {{ y.kind }}
                                </span>
                                <span class="text-xs font-mono text-slate-500">{{ y.name }}</span>
                            </div>
                            <button @click="copyToClipboard(y.data, i)" class="copy-btn">
                                <component :is="copiedIndex === i ? Check : Copy" :size="12" />
                                <span>{{ copiedIndex === i ? 'COPIED' : 'COPY' }}</span>
                            </button>
                        </div>
                        <div class="yaml-block">
                            <pre class="text-blue-300/80">{{ y.data }}</pre>
                        </div>
                    </div>
                </template>

                <template v-else>
                    <div v-if="item.binding_yaml" class="space-y-3">
                        <div class="flex items-center justify-between">
                            <div class="flex items-center gap-3">
                                <span
                                    class="text-blue-400 bg-blue-400/10 text-[9px] px-2 py-0.5 rounded font-black uppercase border border-white/5">
                                    {{ item.binding_type || 'Binding' }}
                                </span>
                                <span class="text-xs font-mono text-slate-500">{{ item.binding_name }}</span>
                            </div>
                            <button @click="copyToClipboard(item.binding_yaml, 'bind')" class="copy-btn">
                                <component :is="copiedIndex === 'bind' ? Check : Copy" :size="12" />
                                <span>{{ copiedIndex === 'bind' ? 'COPIED' : 'COPY' }}</span>
                            </button>
                        </div>
                        <div class="yaml-block">
                            <pre class="text-blue-300/80">{{ item.binding_yaml }}</pre>
                        </div>
                    </div>

                    <div v-if="item.role_yaml" class="space-y-3 mt-6">
                        <div class="flex items-center justify-between">
                            <div class="flex items-center gap-3">
                                <span
                                    class="text-orange-400 bg-orange-400/10 text-[9px] px-2 py-0.5 rounded font-black uppercase border border-white/5">
                                    {{ item.role_kind || 'Role' }}
                                </span>
                                <span class="text-xs font-mono text-slate-500">{{ item.role_name || item.role }}</span>
                            </div>
                            <button @click="copyToClipboard(item.role_yaml, 'role')" class="copy-btn">
                                <component :is="copiedIndex === 'role' ? Check : Copy" :size="12" />
                                <span>{{ copiedIndex === 'role' ? 'COPIED' : 'COPY' }}</span>
                            </button>
                        </div>
                        <div class="yaml-block">
                            <pre class="text-orange-300/80">{{ item.role_yaml }}</pre>
                        </div>
                    </div>
                </template>
            </div>
        </div>
    </section>
</template>
