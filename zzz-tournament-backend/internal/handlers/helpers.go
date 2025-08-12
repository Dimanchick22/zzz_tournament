// internal/handlers/helpers.go
package handlers

import "strings"

// joinStrings объединяет строки с разделителем
func joinStrings(strings []string, separator string) string {
	if len(strings) == 0 {
		return ""
	}
	if len(strings) == 1 {
		return strings[0]
	}

	result := strings[0]
	for i := 1; i < len(strings); i++ {
		result += separator + strings[i]
	}
	return result
}

// contains проверяет, содержит ли строка подстроку
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// trimSpace удаляет пробелы в начале и конце строки
func trimSpace(s string) string {
	return strings.TrimSpace(s)
}

// validateStringLength проверяет длину строки
func validateStringLength(s string, min, max int) bool {
	length := len(s)
	return length >= min && length <= max
}

// isEmptyString проверяет, является ли строка пустой
func isEmptyString(s string) bool {
	return strings.TrimSpace(s) == ""
}
