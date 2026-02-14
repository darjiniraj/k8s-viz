<script setup lang="ts">
import { computed } from 'vue'
import { ShieldAlert } from 'lucide-vue-next'
import RBACDetailsCard from './RBACDetailsCard.vue'
import type { IAMSecurityRecord } from '../types'

const props = defineProps<{
  records: IAMSecurityRecord[]
}>()

const hasRecords = computed(() => props.records.length > 0)

const metadataSections = computed(() => {
  const accessTypes = new Set<string>()
  const bindingKinds = new Set<string>()
  const roleKinds = new Set<string>()
  const policyNames = new Set<string>()

  for (const record of props.records) {
    if (record.attachment_type) accessTypes.add(record.attachment_type)
    const policies = Array.isArray(record.access_policies) ? record.access_policies : []
    for (const policy of policies) {
      if (policy?.trim()) policyNames.add(policy.trim())
    }

    const details = Array.isArray(record.rbac_details) ? record.rbac_details : []
    for (const detail of details) {
      if (detail.binding_kind) bindingKinds.add(detail.binding_kind)
      if (detail.role_kind) roleKinds.add(detail.role_kind)
    }
  }

  return [
    { title: 'Access Types', items: Array.from(accessTypes.values()).sort() },
    { title: 'Binding Types', items: Array.from(bindingKinds.values()).sort() },
    { title: 'Role Types', items: Array.from(roleKinds.values()).sort() },
    { title: 'Access Policies', items: Array.from(policyNames.values()).sort() },
  ].filter((section) => section.items.length > 0)
})
</script>

<template>
  <section class="space-y-4">
    <div v-if="metadataSections.length > 0" class="rounded-xl border border-white/10 bg-slate-950/30 p-3">
      <p class="text-[10px] uppercase tracking-widest text-slate-400 font-black">Metadata</p>
      <div class="mt-2 space-y-2">
        <div v-for="section in metadataSections" :key="section.title">
          <p class="text-[10px] uppercase tracking-wider text-slate-500 mb-1">{{ section.title }}</p>
          <div class="flex flex-wrap gap-2 text-[10px]">
            <span v-for="item in section.items" :key="`${section.title}-${item}`" class="legend-chip border-white/20 bg-slate-900/60 text-slate-200">
              {{ item }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <div v-if="!hasRecords" class="rounded-2xl border border-white/10 bg-slate-900/20 p-8 text-center text-slate-400">
      <ShieldAlert :size="26" class="mx-auto mb-3" />
      <p>No RBAC bindings found for this identity.</p>
    </div>

    <RBACDetailsCard
      v-for="record in records"
      v-else
      :key="`${record.iam_principal}-${record.attachment_type}-${record.k8s_subject}`"
      :record="record"
    />
  </section>
</template>

<style scoped>
.legend-chip {
  border-width: 1px;
  border-radius: 9999px;
  padding: 0.2rem 0.55rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}
</style>
