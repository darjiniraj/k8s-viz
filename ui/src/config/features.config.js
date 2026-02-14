const parseBool = (value, defaultValue = true) => {
  if (value === undefined || value === null || value === '') return defaultValue
  const normalized = String(value).trim().toLowerCase()
  return ['1', 'true', 'yes', 'y', 'on'].includes(normalized)
}

export const FEATURE_FLAGS = {
  ENABLE_SERVICE_ACCOUNT_VIEW: parseBool(import.meta.env.VITE_ENABLE_SERVICE_ACCOUNT_VIEW, true),
  ENABLE_USER_GROUP_VIEW: parseBool(import.meta.env.VITE_ENABLE_USER_GROUP_VIEW, true),
  ENABLE_CILIUM_VIEW: parseBool(import.meta.env.VITE_ENABLE_CILIUM_VIEW, true),
  SHOW_K8S_RESOURCES: parseBool(import.meta.env.VITE_SHOW_K8S_RESOURCES, false),
}

export const K8S_RESOURCE_TABS = [
  {
    key: 'sa',
    label: 'Service Accounts',
    flag: 'ENABLE_SERVICE_ACCOUNT_VIEW',
  },
  {
    key: 'groups',
    label: 'User Groups',
    flag: 'ENABLE_USER_GROUP_VIEW',
  },
  {
    key: 'cilium',
    label: 'Cilium Policies',
    flag: 'ENABLE_CILIUM_VIEW',
  },
]
