export type AttachmentType = 'AccessEntry' | 'PodIdentity' | 'IRSA'

export interface IAMRBACDetail {
  binding_kind: string
  binding_name: string
  binding_namespace?: string
  binding_yaml: string
  role_kind: string
  role_name: string
  role_namespace?: string
  role_yaml: string
  rules: Array<Record<string, unknown>>
}

export interface IAMSecurityRecord {
  iam_principal: string
  attachment_type: AttachmentType
  access_policies?: string[]
  k8s_subject: string
  rbac_details?: IAMRBACDetail[] | null
  summary_placeholder: string
}

export interface IAMSecurityPrincipalGroup {
  iamPrincipal: string
  roleKey: string
  attachments: AttachmentType[]
  records: IAMSecurityRecord[]
}
