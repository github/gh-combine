package common

import "strings"

// Normalize an array of strings to lowercase
func NormalizeArray(array []string) []string {
	normalized := make([]string, len(array))
	for i, label := range array {
		normalized[i] = strings.ToLower(label)
	}
	return normalized
}
