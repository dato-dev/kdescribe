package analyzer

import (
	"sort"
	"strings"
)

type Analysis struct {
	Findings []Finding `json:"findings"`
	Risk     Risk      `json:"risk"`
}

type Risk struct {
	Level   string   `json:"level"`
	Score   int      `json:"score"`
	Reasons []string `json:"reasons"`
}

type Finding struct {
	ID         string     `json:"id"`
	Domain     Domain     `json:"domain"`
	Workload   string     `json:"workload,omitempty"`
	Severity   Severity   `json:"severity"`
	Score      int        `json:"score"`
	Message    string     `json:"message"`
	Hint       string     `json:"hint,omitempty"`
	Line       string     `json:"line,omitempty"`
	LineNo     int        `json:"lineNo,omitempty"`
	Section    string     `json:"section,omitempty"`
	Confidence Confidence `json:"confidence"`
	Evidence   []Evidence `json:"evidence,omitempty"`
}

type Evidence struct {
	LineNo int    `json:"lineNo"`
	Text   string `json:"text"`
}

func Analyze(input string) []Finding {
	return AnalyzeDocument(input).Findings
}

func AnalyzeDocument(input string) Analysis {
	doc := Parse(input)
	findings := collectRuleFindings(doc)
	findings = append(findings, collectStructuredFindings(doc)...)
	findings = deduplicate(findings)
	sortFindings(findings)

	return Analysis{
		Findings: findings,
		Risk:     CalculateRisk(findings),
	}
}

func FilterFindings(findings []Finding, minScore int) []Finding {
	if minScore <= 0 {
		return findings
	}

	filtered := make([]Finding, 0, len(findings))
	for _, finding := range findings {
		if finding.Score >= minScore {
			filtered = append(filtered, finding)
		}
	}

	return filtered
}

func CalculateRisk(findings []Finding) Risk {
	if len(findings) == 0 {
		return Risk{Level: "LOW"}
	}

	maxScore := 0
	var reasons []string
	for _, finding := range findings {
		if finding.Score > maxScore {
			maxScore = finding.Score
		}
	}

	for _, finding := range findings {
		if finding.Score == maxScore {
			reasons = append(reasons, finding.Message)
		}
		if len(reasons) == 3 {
			break
		}
	}

	level := "LOW"
	switch {
	case maxScore >= 90:
		level = "HIGH"
	case maxScore >= 60:
		level = "MEDIUM"
	}

	return Risk{
		Level:   level,
		Score:   maxScore,
		Reasons: reasons,
	}
}

func RiskLevel(findings []Finding) string {
	return CalculateRisk(findings).Level
}

func collectRuleFindings(doc Document) []Finding {
	var findings []Finding
	seen := make(map[string]bool)

	for _, line := range doc.Lines {
		trimmedLine := strings.TrimSpace(line.Text)
		if trimmedLine == "" {
			continue
		}

		for _, rule := range Rules {
			if rule.Pattern.MatchString(line.Text) {
				key := rule.ID + "\x00" + trimmedLine
				if seen[key] {
					continue
				}

				seen[key] = true
				findings = append(findings, Finding{
					ID:         rule.ID,
					Domain:     rule.Domain,
					Workload:   workloadLabel(doc.WorkloadForLine(line.No)),
					Severity:   rule.Severity,
					Score:      rule.Score,
					Message:    rule.Message,
					Hint:       rule.Hint,
					Line:       trimmedLine,
					LineNo:     line.No,
					Section:    doc.SectionForLine(line.No),
					Confidence: RuleBased,
					Evidence:   doc.EvidenceAround(line.No, 1),
				})
			}
		}
	}

	return findings
}

func collectStructuredFindings(doc Document) []Finding {
	var findings []Finding

	for _, container := range doc.Containers {
		if container.RestartCount > 5 {
			score := 70
			if container.RestartCount > 10 {
				score = 85
			}

			findings = append(findings, Finding{
				ID:         "high-restart-count",
				Domain:     DomainResources,
				Workload:   workloadLabel(doc.WorkloadForLine(container.RestartLine.No)),
				Severity:   Warning,
				Score:      score,
				Message:    "Container has a high restart count",
				Hint:       "Inspect previous logs and correlate restarts with probes, OOM kills, or app exits.",
				Line:       strings.TrimSpace(container.RestartLine.Text),
				LineNo:     container.RestartLine.No,
				Section:    "Containers",
				Confidence: Structured,
				Evidence:   doc.EvidenceAround(container.RestartLine.No, 1),
			})
		}

		if container.Ready == "False" && container.ReadyLine.No > 0 {
			findings = append(findings, Finding{
				ID:         "container-not-ready",
				Domain:     DomainProbes,
				Workload:   workloadLabel(doc.WorkloadForLine(container.ReadyLine.No)),
				Severity:   Warning,
				Score:      55,
				Message:    "Container is not ready",
				Hint:       "Check readiness probe, app startup, and container state.",
				Line:       strings.TrimSpace(container.ReadyLine.Text),
				LineNo:     container.ReadyLine.No,
				Section:    "Containers",
				Confidence: Structured,
				Evidence:   doc.EvidenceAround(container.ReadyLine.No, 1),
			})
		}

		if container.StateReason != "" && container.StateLine.No > 0 {
			findings = append(findings, Finding{
				ID:         "container-state-" + strings.ToLower(container.StateReason),
				Domain:     DomainRuntime,
				Workload:   workloadLabel(doc.WorkloadForLine(container.StateLine.No)),
				Severity:   severityForReason(container.StateReason),
				Score:      scoreForReason(container.StateReason, 65),
				Message:    "Container state reason: " + container.StateReason,
				Hint:       "Use the state reason together with events and previous logs to narrow down the cause.",
				Line:       strings.TrimSpace(container.StateLine.Text),
				LineNo:     container.StateLine.No,
				Section:    "Containers",
				Confidence: Structured,
				Evidence:   doc.EvidenceAround(container.StateLine.No, 2),
			})
		}
	}

	for _, condition := range doc.Conditions {
		if condition.Status == "False" && condition.Type == "Ready" {
			findings = append(findings, Finding{
				ID:         "pod-ready-false",
				Domain:     DomainScheduling,
				Workload:   workloadLabel(doc.WorkloadForLine(condition.Line.No)),
				Severity:   Warning,
				Score:      60,
				Message:    "Pod Ready condition is False",
				Hint:       "Check container readiness, scheduling conditions, and recent events.",
				Line:       strings.TrimSpace(condition.Line.Text),
				LineNo:     condition.Line.No,
				Section:    "Conditions",
				Confidence: Structured,
				Evidence:   doc.EvidenceAround(condition.Line.No, 1),
			})
		}
	}

	return findings
}

func deduplicate(findings []Finding) []Finding {
	deduped := make([]Finding, 0, len(findings))
	seen := make(map[string]bool)

	for _, finding := range findings {
		key := finding.ID + "\x00" + finding.Line
		if seen[key] {
			continue
		}
		seen[key] = true
		deduped = append(deduped, finding)
	}

	return deduped
}

func sortFindings(findings []Finding) {
	sort.SliceStable(findings, func(i, j int) bool {
		if findings[i].Score == findings[j].Score {
			return findings[i].LineNo < findings[j].LineNo
		}

		return findings[i].Score > findings[j].Score
	})
}

func severityForReason(reason string) Severity {
	if scoreForReason(reason, 0) >= 90 {
		return Critical
	}

	return Warning
}

func scoreForReason(reason string, fallback int) int {
	switch reason {
	case "OOMKilled", "CrashLoopBackOff":
		return 100
	case "ImagePullBackOff", "ErrImagePull":
		return 90
	default:
		return fallback
	}
}

func workloadLabel(w Workload) string {
	if w.Name == "" {
		return ""
	}

	parts := make([]string, 0, 3)
	if w.Kind != "" {
		parts = append(parts, strings.ToLower(w.Kind))
	}
	if w.Namespace != "" {
		parts = append(parts, w.Namespace)
	}
	parts = append(parts, w.Name)

	return strings.Join(parts, "/")
}
