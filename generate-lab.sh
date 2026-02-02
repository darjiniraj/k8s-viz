#!/bin/bash

# 1. Create 10 Namespaces
namespaces=("finance" "hr" "engineering" "marketing" "ops" "security" "data-science" "legal" "sales" "it-support")

for ns in "${namespaces[@]}"; do
  kubectl create ns $ns --dry-run=client -o yaml | kubectl apply -f -
done

# 2. Create Service Accounts & Roles (One-to-Many Mappings)
for ns in "${namespaces[@]}"; do
  # Each namespace gets a 'worker' and a 'manager'
  kubectl create sa app-worker -n $ns
  kubectl create sa app-manager -n $ns

  # Add AWS IAM Annotations for the SA view
  kubectl annotate sa app-worker -n $ns eks.amazonaws.com/role-arn=arn:aws:iam::123456789012:role/$ns-worker-role
  kubectl annotate sa app-manager -n $ns eks.amazonaws.com/role-arn=arn:aws:iam::123456789012:role/$ns-manager-role

  # Create multiple roles in each namespace
  cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata: { name: pod-reader, namespace: $ns }
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata: { name: secret-accessor, namespace: $ns }
rules:
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["*"]
EOF

  # Map app-manager to BOTH roles (One-to-Many)
  kubectl create rolebinding manager-rb-1 -n $ns --role=pod-reader --serviceaccount=$ns:app-manager
  kubectl create rolebinding manager-rb-2 -n $ns --role=secret-accessor --serviceaccount=$ns:app-manager
done

# 3. Create Complex Group Mappings (The "Many" part)
# Group: "platform-admins" gets ClusterAdmin + specific namespace roles
kubectl create clusterrolebinding platform-admin-global --clusterrole=cluster-admin --group=platform-admins

for ns in "security" "ops" "it-support"; do
  kubectl create rolebinding platform-ns-access -n $ns --clusterrole=edit --group=platform-admins
done

# Group: "audit-team" (Read-only everywhere)
kubectl create clusterrolebinding audit-global --clusterrole=view --group=audit-team

# Group: "data-science-group" (Many roles in one namespace)
ns="data-science"
kubectl create rolebinding ds-pods -n $ns --role=pod-reader --group=data-science-group
kubectl create rolebinding ds-secrets -n $ns --role=secret-accessor --group=data-science-group

echo "✅ Lab data generated! Check your dashboard."