package main

// Aggregate performs the "Reverse Lookup": Grouping all K8s RBAC by AWS IAM ARN
func Aggregate(saRows []SecurityRow, groupRows []GroupSecurityRow, accessEntries map[string][]string) []IAMAuditRow {
	iamMap := make(map[string]*IAMAuditRow)

	// 1. Process Service Account data (IRSA)
	for _, sa := range saRows {
		// Skip rows that don't have an IAM Role assigned
		if sa.IAMRole == "" || sa.IAMRole == "None" {
			continue
		}

		if _, exists := iamMap[sa.IAMRole]; !exists {
			iamMap[sa.IAMRole] = &IAMAuditRow{
				IAMRole:  sa.IAMRole,
				Type:     "iam",
				AllYAMLs: []YamlBlock{},
			}
		}

		// Add both the Binding and the Role to maintain the "Pair" logic in the UI
		iamMap[sa.IAMRole].AllYAMLs = append(iamMap[sa.IAMRole].AllYAMLs,
			YamlBlock{Kind: sa.BindingType, Name: sa.BindingName, Data: sa.BindingYAML, Namespace: sa.Namespace},
			YamlBlock{Kind: "Role", Name: sa.RoleName, Data: sa.RoleYAML, Namespace: sa.Namespace},
		)
	}

	// 2. Process Group data (Access Entries)
	for _, g := range groupRows {
		if g.GroupName == "data-science-group" {
			arn := "arn:aws:iam::123456789012:role/EKS-DataScience-Access"

			if _, exists := iamMap[arn]; !exists {
				// Initialize with a pointer to a new struct
				iamMap[arn] = &IAMAuditRow{
					IAMRole:  arn,
					Type:     "iam",
					AllYAMLs: []YamlBlock{}, // Initialize the slice
				}
			}

			// Now you can append directly because iamMap[arn] is a pointer
			iamMap[arn].AllYAMLs = append(iamMap[arn].AllYAMLs, g.AllYAMLs...)
		}
	}
	// NEW REAL LOGIC: Process Access Entries
	for iamArn, k8sGroups := range accessEntries {
		if _, exists := iamMap[iamArn]; !exists {
			iamMap[iamArn] = &IAMAuditRow{
				IAMRole:  iamArn,
				Type:     "iam",
				AllYAMLs: []YamlBlock{},
			}
		}

		// For every group this IAM ARN belongs to...
		for _, groupName := range k8sGroups {
			// ...find the matching Group security data we got from K8s RBAC
			for _, gRow := range groupRows {
				if gRow.GroupName == groupName {
					iamMap[iamArn].AllYAMLs = append(iamMap[iamArn].AllYAMLs, gRow.AllYAMLs...)
				}
			}
		}
	}

	// 3. Convert Map to Slice for the API
	var result []IAMAuditRow
	for _, v := range iamMap {
		// Placeholder for the AI Summary button logic
		v.AISummary = "Identity found. Click 'Analyze' to evaluate permissions."
		result = append(result, *v)
	}
	return result
}
