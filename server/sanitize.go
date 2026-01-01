package main

import (
	"regexp"
	"strings"
)

var htmlTagRegex = regexp.MustCompile(`<[^>]*>`)

// SanitizeInput removes HTML tags and trims whitespace from the input string.
func SanitizeInput(input string) string {
	// Remove HTML tags
	sanitized := htmlTagRegex.ReplaceAllString(input, "")
	
	// Trim whitespace
	return strings.TrimSpace(sanitized)
}

// SanitizeMultiline preserves newlines but removes HTML tags.
func SanitizeMultiline(input string) string {
	// Remove HTML tags
	sanitized := htmlTagRegex.ReplaceAllString(input, "")
	return strings.TrimSpace(sanitized)
}
