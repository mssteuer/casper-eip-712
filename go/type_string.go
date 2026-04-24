package eip712

import (
	"fmt"
	"sort"
	"strings"
)

// BuildTypeString returns the EIP-712 type string for a single type.
// Format: "TypeName(field1Type field1Name,field2Type field2Name,...)"
func BuildTypeString(typeName string, fields []TypedField) string {
	var sb strings.Builder
	sb.WriteString(typeName)
	sb.WriteByte('(')
	for i, f := range fields {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(f.Type)
		sb.WriteByte(' ')
		sb.WriteString(f.Name)
	}
	sb.WriteByte(')')
	return sb.String()
}

// BuildCanonicalTypeString returns the canonical EIP-712 type string for
// primaryType, including all transitively referenced struct types sorted
// alphabetically (per the EIP-712 spec).
//
// Example for nested types:
//
//	"Mail(Person from,Person to,string contents)Person(string name,address wallet)"
func BuildCanonicalTypeString(primaryType string, types TypeDefinitions) (string, error) {
	deps, err := collectDeps(primaryType, types, make(map[string]bool))
	if err != nil {
		return "", err
	}

	// Sort dependencies alphabetically, excluding the primary type.
	sorted := make([]string, 0, len(deps)-1)
	for name := range deps {
		if name != primaryType {
			sorted = append(sorted, name)
		}
	}
	sort.Strings(sorted)

	var sb strings.Builder
	fields, ok := types[primaryType]
	if !ok {
		return "", fmt.Errorf("eip712: primary type %q not found in type definitions", primaryType)
	}
	sb.WriteString(BuildTypeString(primaryType, fields))
	for _, dep := range sorted {
		sb.WriteString(BuildTypeString(dep, types[dep]))
	}
	return sb.String(), nil
}

// collectDeps recursively collects all struct types referenced by typeName.
func collectDeps(typeName string, types TypeDefinitions, visited map[string]bool) (map[string]bool, error) {
	if visited[typeName] {
		return visited, nil
	}
	fields, ok := types[typeName]
	if !ok {
		// Primitive type - no fields to recurse into.
		return visited, nil
	}
	visited[typeName] = true
	for _, f := range fields {
		if _, isStruct := types[f.Type]; isStruct {
			if _, err := collectDeps(f.Type, types, visited); err != nil {
				return visited, err
			}
		}
	}
	return visited, nil
}
