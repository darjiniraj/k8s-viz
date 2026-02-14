<script setup lang="ts">
import { computed } from 'vue'
import { ArrowRight, Link2 } from 'lucide-vue-next'
import type { IAMRBACDetail, IAMSecurityRecord } from '../types'

const props = defineProps<{
  record: IAMSecurityRecord
}>()

const rbacDetails = computed<IAMRBACDetail[]>(() => {
  const details = props.record?.rbac_details
  return Array.isArray(details) ? details : []
})

const toNamespace = (detail: IAMRBACDetail): string => {
  return detail.binding_namespace || detail.role_namespace || 'Cluster-Wide'
}

const extractAnnotationLabel = (yaml: string): string | null => {
  if (!yaml) return null
  const lines = yaml.split('\n')
  const annotationStart = lines.findIndex((line) => /^\s+annotations:\s*$/.test(line))
  if (annotationStart < 0) return null

  for (let idx = annotationStart + 1; idx < lines.length; idx += 1) {
    const line = lines[idx]
    if (!/^\s+[^:\s][^:]*:\s*.+$/.test(line)) break
    const match = line.match(/^\s+([^:]+):\s*(.+)\s*$/)
    if (!match) continue
    const key = match[1].trim().toLowerCase()
    const value = match[2].trim().replace(/^['"]|['"]$/g, '')
    if (!value) continue
    if (key.includes('description') || key.includes('purpose') || key.includes('summary') || key.includes('display-name')) {
      return value
    }
  }

  return null
}

const smartRoleLabel = (roleName: string): string | null => {
  const normalized = roleName.trim().toLowerCase()
  if (!normalized) return null
  if (normalized === 'view' || normalized.includes('readonly') || normalized.includes('read-only') || normalized.includes('reader')) {
    return 'View-Only Access'
  }
  if (normalized === 'admin' || normalized === 'cluster-admin' || normalized.endsWith('-admin') || normalized.includes('administrator')) {
    return 'Administrative Access'
  }
  if (normalized === 'edit' || normalized.endsWith('-edit')) {
    return 'Edit Access'
  }
  return null
}

const accessPolicyLabel = computed(() => {
  const policies = Array.isArray(props.record.access_policies) ? props.record.access_policies : []
  return policies.length > 0 ? policies.join(', ') : null
})

const accessNodeLabel = computed(() => {
  if (props.record.attachment_type === 'PodIdentity') return 'Pod Identity'
  if (props.record.attachment_type === 'IRSA') return 'IRSA'
  return 'Access Entry'
})

const metadataLabel = (detail: IAMRBACDetail): string | null => {
  const annotation = extractAnnotationLabel(detail.role_yaml)
  if (annotation) return annotation
  const smart = smartRoleLabel(detail.role_name)
  if (smart) return smart
  if (accessPolicyLabel.value) return `Access Policy: ${accessPolicyLabel.value}`
  return null
}

const badgeClass = (type: 'identity' | 'access' | 'binding' | 'role') => {
  if (type === 'identity') return 'bg-cyan-500/15 border-cyan-400/30 text-cyan-200'
  if (type === 'access') return 'bg-indigo-500/15 border-indigo-400/30 text-indigo-200'
  if (type === 'binding') return 'bg-amber-500/15 border-amber-400/30 text-amber-200'
  return 'bg-emerald-500/15 border-emerald-400/30 text-emerald-200'
}
</script>

<template>
  <article class="workspace-card">
    <header class="flex items-center justify-between gap-3">
      <p class="font-mono text-[11px] text-slate-100 break-all">{{ record.iam_principal }}</p>
      <span class="text-[10px] px-2 py-1 rounded-full border border-blue-400/30 bg-blue-500/10 text-blue-200 uppercase tracking-wider">
        {{ record.attachment_type }}
      </span>
    </header>

    <p class="mt-2 text-[11px] text-slate-400">
      Subject: <span class="text-slate-200">{{ record.k8s_subject }}</span>
    </p>

    <div v-if="rbacDetails.length === 0" class="mt-4 rounded-xl border border-amber-400/20 bg-amber-500/10 p-3 text-sm text-amber-100">
      No RBAC bindings found for this identity.
    </div>

    <div class="mt-4 space-y-3">
      <div v-for="detail in rbacDetails" :key="`${detail.binding_kind}-${detail.binding_name}-${detail.role_kind}-${detail.role_name}`" class="rounded-xl border border-white/10 bg-slate-950/40 p-3">
        <div class="flow-grid">
          <div class="flow-node" :title="`IAM principal ${record.iam_principal}`">
            <span :class="badgeClass('identity')" class="flow-badge">Identity</span>
            <p class="flow-title">{{ record.iam_principal }}</p>
            <p class="flow-sub">IAM Role/User</p>
          </div>
          <ArrowRight class="flow-arrow" :size="14" />

          <div class="flow-node" :title="`Attachment ${record.attachment_type}`">
            <span :class="badgeClass('access')" class="flow-badge">{{ accessNodeLabel }}</span>
            <p class="flow-title">{{ record.attachment_type }}</p>
            <p v-if="accessPolicyLabel" class="flow-sub">{{ accessPolicyLabel }}</p>
          </div>
          <ArrowRight class="flow-arrow" :size="14" />

          <div class="flow-node" :title="`Binding ${detail.binding_kind} ${detail.binding_name}`">
            <span :class="badgeClass('binding')" class="flow-badge">Binding</span>
            <p class="flow-title">{{ detail.binding_kind }}</p>
            <p class="flow-sub">{{ detail.binding_name }} ({{ toNamespace(detail) }})</p>
          </div>
          <ArrowRight class="flow-arrow" :size="14" />

          <div class="flow-node" :title="`Role ${detail.role_kind} ${detail.role_name}`">
            <span :class="badgeClass('role')" class="flow-badge">Role</span>
            <p class="flow-title">{{ detail.role_kind }}</p>
            <p class="flow-sub">{{ detail.role_name }}</p>
          </div>
        </div>

        <div v-if="metadataLabel(detail)" class="mt-3 flex items-start gap-2 text-[11px] text-slate-300">
          <Link2 :size="13" class="text-indigo-300 mt-0.5 shrink-0" />
          <p>{{ metadataLabel(detail) }}</p>
        </div>
      </div>
    </div>
  </article>
</template>

<style scoped>
.flow-grid {
  display: grid;
  grid-template-columns: 1fr auto 1fr auto 1fr auto 1fr;
  gap: 0.5rem;
  align-items: center;
}

.flow-node {
  border: 1px solid rgb(255 255 255 / 0.08);
  border-radius: 0.75rem;
  background: rgb(2 6 23 / 0.45);
  padding: 0.5rem;
  min-height: 80px;
}

.flow-badge {
  display: inline-block;
  border-width: 1px;
  border-radius: 9999px;
  padding: 0.1rem 0.5rem;
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.flow-title {
  color: #e2e8f0;
  font-size: 11px;
  margin-top: 0.45rem;
  line-height: 1.2;
  word-break: break-word;
}

.flow-sub {
  color: #94a3b8;
  font-size: 10px;
  margin-top: 0.15rem;
  line-height: 1.3;
  word-break: break-word;
}

.flow-arrow {
  color: #64748b;
}

@media (max-width: 1024px) {
  .flow-grid {
    grid-template-columns: 1fr;
  }

  .flow-arrow {
    transform: rotate(90deg);
    justify-self: center;
  }
}
</style>
