package utils

import (
	"regexp"
	"strings"
)

func ToScreamingSnakeCase(s string) string {
	// Replace spaces and hyphens with underscores
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	
	// Insert underscore before uppercase letters (for camelCase)
	re := regexp.MustCompile("([a-z0-9])([A-Z])")
	s = re.ReplaceAllString(s, "${1}_${2}")
	
	// Convert to uppercase
	s = strings.ToUpper(s)
	
	// Remove multiple consecutive underscores
	re = regexp.MustCompile("_+")
	s = re.ReplaceAllString(s, "_")
	
	// Trim underscores from start and end
	s = strings.Trim(s, "_")
	
	return s
}

func ToScreamingSnakeCasePtr(s *string) *string {
	if s == nil {
		return nil
	}
	result := ToScreamingSnakeCase(*s)
	return &result
}