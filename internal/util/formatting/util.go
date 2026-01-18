package formatting

import (
	"strings"
	"unicode/utf8"
)

// ToSnakeCase converts a string from CamelCase to snake_case.
func ToSnakeCase(str string) string {
	var result strings.Builder
	for i, r := range str {
		if i > 0 && (r >= 'A' && r <= 'Z') {
			// Check if previous char was not an uppercase letter
			prev, _ := utf8.DecodeLastRuneInString(str[:i])
			if prev != '_' && !(prev >= 'A' && prev <= 'Z') {
				result.WriteRune('_')
			}
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
