package schema

import (
	"fmt"
	"regexp"
	"strings"
)

// BriefSchema defines required fields and their validation rules.
type BriefSchema struct {
	Required []FieldRule
}

// FieldRule describes one required field.
type FieldRule struct {
	Name        string
	Pattern     *regexp.Regexp
	AllowEmpty  bool
}

// DefaultBriefSchema is the Iron Tree brief contract.
var DefaultBriefSchema = BriefSchema{
	Required: []FieldRule{
		{Name: "Task Key", Pattern: regexp.MustCompile(`Task Key:\s*(\S+)`), AllowEmpty: false},
		{Name: "Issue Number", Pattern: regexp.MustCompile(`Issue Number:\s*(\d+)`), AllowEmpty: false},
		{Name: "Spec Revision", Pattern: regexp.MustCompile(`Spec Revision:\s*([a-f0-9]+)`), AllowEmpty: false},
		{Name: "Current Stage", Pattern: regexp.MustCompile(`Current Stage:\s*(\w+)`), AllowEmpty: false},
		{Name: "Next Stage", Pattern: regexp.MustCompile(`Next Stage:\s*(\w+)`), AllowEmpty: false},
		{Name: "Attempt Count", Pattern: regexp.MustCompile(`Attempt Count:\s*(\d+)`), AllowEmpty: false},
		{Name: "Max Retry Count", Pattern: regexp.MustCompile(`Max Retry Count:\s*(\d+)`), AllowEmpty: false},
	},
}

// ValidateBrief checks raw markdown against the schema.
// Returns a slice of error messages; empty slice means valid.
func ValidateBrief(raw string) []string {
	var errs []string
	for _, rule := range DefaultBriefSchema.Required {
		if !rule.Pattern.MatchString(raw) {
			errs = append(errs, fmt.Sprintf("missing or malformed field: %s", rule.Name))
			continue
		}
		if !rule.AllowEmpty {
			m := rule.Pattern.FindStringSubmatch(raw)
			if len(m) > 1 && strings.TrimSpace(m[1]) == "" {
				errs = append(errs, fmt.Sprintf("empty value for field: %s", rule.Name))
			}
		}
	}
	return errs
}
