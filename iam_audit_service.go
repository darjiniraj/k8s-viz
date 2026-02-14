package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const iamAuditPageLimit int64 = 500

type iamSubjectKey struct {
	kind      string
	name      string
	namespace string
}

type principalAttachment struct {
	iamPrincipal   string
	attachmentType string
	accessPolicies []string
	k8sSubject     string
	subjectKey     iamSubjectKey
}

type roleResolved struct {
	yaml      string
	rules     []rbacv1.PolicyRule
	kind      string
	name      string
	namespace string
}

func BuildIAMRBACMap(ctx context.Context, svc *K8sService, clusterName, region string) ([]IAMRBACMapRow, error) {
	clusterName = strings.TrimSpace(clusterName)
	if clusterName == "" {
		clusterName = strings.TrimSpace(os.Getenv("EKS_CLUSTER_NAME"))
	}
	if clusterName == "" {
		return nil, fmt.Errorf("missing EKS cluster name: set query param 'cluster' or EKS_CLUSTER_NAME")
	}

	region = strings.TrimSpace(region)
	if region == "" {
		region = strings.TrimSpace(os.Getenv("AWS_REGION"))
	}
	if region == "" {
		region = strings.TrimSpace(os.Getenv("AWS_DEFAULT_REGION"))
	}
	if region == "" {
		return nil, fmt.Errorf("missing AWS region: set query param 'region', AWS_REGION, or AWS_DEFAULT_REGION")
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	eksClient := eks.NewFromConfig(cfg)

	accessEntries, err := fetchAccessEntryAttachments(ctx, eksClient, clusterName)
	if err != nil {
		return nil, err
	}

	podIdentityEntries, err := fetchPodIdentityAttachments(ctx, eksClient, clusterName)
	if err != nil {
		return nil, err
	}

	irsaEntries, err := fetchIRSAAttachments(ctx, svc)
	if err != nil {
		return nil, err
	}

	allAttachments := make([]principalAttachment, 0, len(accessEntries)+len(podIdentityEntries)+len(irsaEntries))
	allAttachments = append(allAttachments, accessEntries...)
	allAttachments = append(allAttachments, podIdentityEntries...)
	allAttachments = append(allAttachments, irsaEntries...)

	if len(allAttachments) == 0 {
		return []IAMRBACMapRow{}, nil
	}

	groupTargets := map[string]struct{}{}
	saTargets := map[string]struct{}{}

	for _, entry := range allAttachments {
		if entry.subjectKey.kind == "Group" {
			groupTargets[entry.subjectKey.name] = struct{}{}
			continue
		}
		saTargets[entry.subjectKey.namespace+"/"+entry.subjectKey.name] = struct{}{}
	}

	detailsBySubject, err := buildRBACDetailsBySubject(ctx, svc, groupTargets, saTargets)
	if err != nil {
		return nil, err
	}

	rows := make([]IAMRBACMapRow, 0, len(allAttachments))
	for _, attachment := range allAttachments {
		details := detailsBySubject[attachment.subjectKey]
		row := IAMRBACMapRow{
			IAMPrincipal:       attachment.iamPrincipal,
			AttachmentType:     attachment.attachmentType,
			AccessPolicies:     attachment.accessPolicies,
			K8sSubject:         attachment.k8sSubject,
			RBACDetails:        details,
			SummaryPlaceholder: "",
		}
		rows = append(rows, row)
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].IAMPrincipal != rows[j].IAMPrincipal {
			return rows[i].IAMPrincipal < rows[j].IAMPrincipal
		}
		if rows[i].AttachmentType != rows[j].AttachmentType {
			return rows[i].AttachmentType < rows[j].AttachmentType
		}
		return rows[i].K8sSubject < rows[j].K8sSubject
	})

	return rows, nil
}

func fetchAccessEntryAttachments(ctx context.Context, eksClient *eks.Client, clusterName string) ([]principalAttachment, error) {
	listInput := &eks.ListAccessEntriesInput{
		ClusterName: aws.String(clusterName),
		MaxResults:  aws.Int32(100),
	}

	pager := eks.NewListAccessEntriesPaginator(eksClient, listInput)
	attachments := []principalAttachment{}

	for pager.HasMorePages() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list access entries: %w", err)
		}

		for _, principalArn := range page.AccessEntries {
			describeOut, err := eksClient.DescribeAccessEntry(ctx, &eks.DescribeAccessEntryInput{
				ClusterName:  aws.String(clusterName),
				PrincipalArn: aws.String(principalArn),
			})
			if err != nil {
				log.Printf("describe access entry failed for %s: %v", principalArn, err)
				continue
			}

			if describeOut.AccessEntry == nil {
				continue
			}

			policyNames := fetchAccessPolicyNames(ctx, eksClient, clusterName, principalArn)
			for _, group := range describeOut.AccessEntry.KubernetesGroups {
				group = strings.TrimSpace(group)
				if group == "" {
					continue
				}

				attachments = append(attachments, principalAttachment{
					iamPrincipal:   principalArn,
					attachmentType: "AccessEntry",
					accessPolicies: policyNames,
					k8sSubject:     group,
					subjectKey: iamSubjectKey{
						kind: "Group",
						name: group,
					},
				})
			}
		}
	}

	return attachments, nil
}

func fetchPodIdentityAttachments(ctx context.Context, eksClient *eks.Client, clusterName string) ([]principalAttachment, error) {
	listInput := &eks.ListPodIdentityAssociationsInput{
		ClusterName: aws.String(clusterName),
		MaxResults:  aws.Int32(100),
	}

	pager := eks.NewListPodIdentityAssociationsPaginator(eksClient, listInput)
	attachments := []principalAttachment{}

	for pager.HasMorePages() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list pod identity associations: %w", err)
		}

		for _, assoc := range page.Associations {
			assocID := aws.ToString(assoc.AssociationId)
			if assocID == "" {
				continue
			}

			describeOut, err := eksClient.DescribePodIdentityAssociation(ctx, &eks.DescribePodIdentityAssociationInput{
				ClusterName:   aws.String(clusterName),
				AssociationId: aws.String(assocID),
			})
			if err != nil {
				log.Printf("describe pod identity association failed for %s: %v", assocID, err)
				continue
			}
			if describeOut.Association == nil {
				continue
			}

			roleArn := aws.ToString(describeOut.Association.RoleArn)
			if roleArn == "" {
				roleArn = aws.ToString(describeOut.Association.TargetRoleArn)
			}
			namespace := aws.ToString(describeOut.Association.Namespace)
			saName := aws.ToString(describeOut.Association.ServiceAccount)
			if roleArn == "" || namespace == "" || saName == "" {
				continue
			}

			attachments = append(attachments, principalAttachment{
				iamPrincipal:   roleArn,
				attachmentType: "PodIdentity",
				k8sSubject:     fmt.Sprintf("%s/%s", namespace, saName),
				subjectKey: iamSubjectKey{
					kind:      "ServiceAccount",
					name:      saName,
					namespace: namespace,
				},
			})
		}
	}

	return attachments, nil
}

func fetchIRSAAttachments(ctx context.Context, svc *K8sService) ([]principalAttachment, error) {
	sas, err := listAllServiceAccounts(ctx, svc.Clientset)
	if err != nil {
		return nil, fmt.Errorf("list service accounts for IRSA: %w", err)
	}

	attachments := []principalAttachment{}
	for _, sa := range sas {
		roleArn := strings.TrimSpace(sa.Annotations["eks.amazonaws.com/role-arn"])
		if roleArn == "" {
			continue
		}

		attachments = append(attachments, principalAttachment{
			iamPrincipal:   roleArn,
			attachmentType: "IRSA",
			k8sSubject:     fmt.Sprintf("%s/%s", sa.Namespace, sa.Name),
			subjectKey: iamSubjectKey{
				kind:      "ServiceAccount",
				name:      sa.Name,
				namespace: sa.Namespace,
			},
		})
	}

	return attachments, nil
}

func fetchAccessPolicyNames(ctx context.Context, eksClient *eks.Client, clusterName, principalArn string) []string {
	input := &eks.ListAssociatedAccessPoliciesInput{
		ClusterName:  aws.String(clusterName),
		PrincipalArn: aws.String(principalArn),
		MaxResults:   aws.Int32(100),
	}

	pager := eks.NewListAssociatedAccessPoliciesPaginator(eksClient, input)
	names := []string{}
	seen := map[string]struct{}{}

	for pager.HasMorePages() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Printf("list associated access policies failed for %s: %v", principalArn, err)
			break
		}

		for _, assoc := range page.AssociatedAccessPolicies {
			policyArn := strings.TrimSpace(aws.ToString(assoc.PolicyArn))
			if policyArn == "" {
				continue
			}
			chunks := strings.Split(policyArn, "/")
			name := strings.TrimSpace(chunks[len(chunks)-1])
			if name == "" {
				name = policyArn
			}
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			names = append(names, name)
		}
	}

	sort.Strings(names)
	return names
}

func buildRBACDetailsBySubject(ctx context.Context, svc *K8sService, groupTargets, saTargets map[string]struct{}) (map[iamSubjectKey][]IAMRBACDetail, error) {
	roleBindings, err := listAllRoleBindings(ctx, svc.Clientset)
	if err != nil {
		return nil, fmt.Errorf("list role bindings: %w", err)
	}
	clusterRoleBindings, err := listAllClusterRoleBindings(ctx, svc.Clientset)
	if err != nil {
		return nil, fmt.Errorf("list cluster role bindings: %w", err)
	}

	result := map[iamSubjectKey][]IAMRBACDetail{}
	roleCache := map[string]roleResolved{}

	for _, rb := range roleBindings {
		subjectKeys := matchSubjects(rb.Subjects, rb.Namespace, groupTargets, saTargets)
		if len(subjectKeys) == 0 {
			continue
		}

		bindingYAML := svc.MarshalToYaml(rb)
		roleData, err := resolveRoleRef(ctx, svc, roleCache, rb.RoleRef.Kind, rb.RoleRef.Name, rb.Namespace)
		if err != nil {
			log.Printf("resolve role ref failed for %s/%s: %v", rb.Namespace, rb.Name, err)
			continue
		}

		detail := IAMRBACDetail{
			BindingKind:      "RoleBinding",
			BindingName:      rb.Name,
			BindingNamespace: rb.Namespace,
			BindingYAML:      bindingYAML,
			RoleKind:         roleData.kind,
			RoleName:         roleData.name,
			RoleNamespace:    roleData.namespace,
			RoleYAML:         roleData.yaml,
			Rules:            roleData.rules,
		}
		for _, key := range subjectKeys {
			result[key] = append(result[key], detail)
		}
	}

	for _, crb := range clusterRoleBindings {
		subjectKeys := matchSubjects(crb.Subjects, "", groupTargets, saTargets)
		if len(subjectKeys) == 0 {
			continue
		}

		bindingYAML := svc.MarshalToYaml(crb)
		roleData, err := resolveRoleRef(ctx, svc, roleCache, crb.RoleRef.Kind, crb.RoleRef.Name, "")
		if err != nil {
			log.Printf("resolve cluster role ref failed for %s: %v", crb.Name, err)
			continue
		}

		detail := IAMRBACDetail{
			BindingKind: "ClusterRoleBinding",
			BindingName: crb.Name,
			BindingYAML: bindingYAML,
			RoleKind:    roleData.kind,
			RoleName:    roleData.name,
			RoleYAML:    roleData.yaml,
			Rules:       roleData.rules,
		}
		for _, key := range subjectKeys {
			result[key] = append(result[key], detail)
		}
	}

	for key := range result {
		sort.Slice(result[key], func(i, j int) bool {
			if result[key][i].BindingKind != result[key][j].BindingKind {
				return result[key][i].BindingKind < result[key][j].BindingKind
			}
			if result[key][i].BindingNamespace != result[key][j].BindingNamespace {
				return result[key][i].BindingNamespace < result[key][j].BindingNamespace
			}
			return result[key][i].BindingName < result[key][j].BindingName
		})
	}

	return result, nil
}

func matchSubjects(subjects []rbacv1.Subject, defaultNamespace string, groupTargets, saTargets map[string]struct{}) []iamSubjectKey {
	keys := []iamSubjectKey{}
	for _, sub := range subjects {
		switch sub.Kind {
		case "Group":
			if _, ok := groupTargets[sub.Name]; !ok {
				continue
			}
			keys = append(keys, iamSubjectKey{
				kind: "Group",
				name: sub.Name,
			})
		case "ServiceAccount":
			ns := sub.Namespace
			if ns == "" {
				ns = defaultNamespace
			}
			targetKey := ns + "/" + sub.Name
			if _, ok := saTargets[targetKey]; !ok {
				continue
			}
			keys = append(keys, iamSubjectKey{
				kind:      "ServiceAccount",
				name:      sub.Name,
				namespace: ns,
			})
		}
	}
	return keys
}

func resolveRoleRef(ctx context.Context, svc *K8sService, cache map[string]roleResolved, roleKind, roleName, roleNS string) (roleResolved, error) {
	cacheKey := roleKind + "|" + roleNS + "|" + roleName
	if cached, ok := cache[cacheKey]; ok {
		return cached, nil
	}

	var (
		objYAML string
		rules   []rbacv1.PolicyRule
		kind    string
	)

	switch roleKind {
	case "ClusterRole":
		role, err := svc.Clientset.RbacV1().ClusterRoles().Get(ctx, roleName, metav1.GetOptions{})
		if err != nil {
			return roleResolved{}, err
		}
		objYAML = svc.MarshalToYaml(role)
		rules = role.Rules
		kind = "ClusterRole"
	default:
		role, err := svc.Clientset.RbacV1().Roles(roleNS).Get(ctx, roleName, metav1.GetOptions{})
		if err != nil {
			return roleResolved{}, err
		}
		objYAML = svc.MarshalToYaml(role)
		rules = role.Rules
		kind = "Role"
	}

	resolved := roleResolved{
		yaml:      objYAML,
		rules:     rules,
		kind:      kind,
		name:      roleName,
		namespace: roleNS,
	}
	cache[cacheKey] = resolved
	return resolved, nil
}

func listAllServiceAccounts(ctx context.Context, clientset *kubernetes.Clientset) ([]corev1.ServiceAccount, error) {
	items := []corev1.ServiceAccount{}
	continueToken := ""

	for {
		out, err := clientset.CoreV1().ServiceAccounts("").List(ctx, metav1.ListOptions{
			Limit:    iamAuditPageLimit,
			Continue: continueToken,
		})
		if err != nil {
			return nil, err
		}

		items = append(items, out.Items...)
		if out.Continue == "" {
			break
		}
		continueToken = out.Continue
	}

	return items, nil
}

func listAllRoleBindings(ctx context.Context, clientset *kubernetes.Clientset) ([]rbacv1.RoleBinding, error) {
	items := []rbacv1.RoleBinding{}
	continueToken := ""

	for {
		out, err := clientset.RbacV1().RoleBindings("").List(ctx, metav1.ListOptions{
			Limit:    iamAuditPageLimit,
			Continue: continueToken,
		})
		if err != nil {
			return nil, err
		}

		items = append(items, out.Items...)
		if out.Continue == "" {
			break
		}
		continueToken = out.Continue
	}

	return items, nil
}

func listAllClusterRoleBindings(ctx context.Context, clientset *kubernetes.Clientset) ([]rbacv1.ClusterRoleBinding, error) {
	items := []rbacv1.ClusterRoleBinding{}
	continueToken := ""

	for {
		out, err := clientset.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{
			Limit:    iamAuditPageLimit,
			Continue: continueToken,
		})
		if err != nil {
			return nil, err
		}

		items = append(items, out.Items...)
		if out.Continue == "" {
			break
		}
		continueToken = out.Continue
	}

	return items, nil
}
