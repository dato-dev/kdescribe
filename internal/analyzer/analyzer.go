package analyzer

import (
	"strings"
)

type Finding struct {
	Severity Severity
	Message  string
	Line     string
}

func Analyze(input string) []Finding {
	lines := strings.Split(input, "\n")

	var findings []Finding

	for _, line := range lines {
		for _, rule := range Rules {
			if strings.Contains(line, rule.Pattern) {
				findings = append(findings, Finding{
					Severity: rule.Severity,
					Message:  rule.Message,
					Line:     strings.TrimSpace(line),
				})
			}
		}
	}

	return findings
}
