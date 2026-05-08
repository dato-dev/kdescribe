package renderer

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/dato-dev/kdescribe/internal/analyzer"

	"github.com/charmbracelet/lipgloss"
)

const (
	FormatHuman    = "human"
	FormatJSON     = "json"
	FormatMarkdown = "markdown"
)

type Options struct {
	Format  string
	NoColor bool
	ShowAll bool
}

var (
	criticalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6"))

	mutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))
)

func Render(w io.Writer, analysis analyzer.Analysis, opts Options) error {
	switch opts.Format {
	case "", FormatHuman:
		renderHuman(w, analysis, opts)
	case FormatJSON:
		return renderJSON(w, analysis)
	case FormatMarkdown:
		renderMarkdown(w, analysis)
	default:
		return fmt.Errorf("unsupported output format %q", opts.Format)
	}

	return nil
}

func renderHuman(w io.Writer, analysis analyzer.Analysis, opts Options) {
	findings := visibleFindings(analysis.Findings, opts)
	if len(findings) == 0 {
		fmt.Fprintln(w, "No obvious issues found")
		return
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Summary")
	fmt.Fprintf(w, "  Cluster health risk: %s", renderRisk(analysis.Risk, opts))
	if len(analysis.Risk.Reasons) > 0 {
		fmt.Fprintf(w, " because %s", strings.Join(analysis.Risk.Reasons, ", "))
	}
	fmt.Fprintln(w)
	fmt.Fprintf(w, "  Findings: %d\n", len(analysis.Findings))
	fmt.Fprintln(w)

	fmt.Fprintln(w, "Findings")

	for _, f := range findings {
		fmt.Fprintf(
			w,
			"%s %s %s\n",
			severityLabel(f.Severity, opts),
			f.Message,
			muted(fmt.Sprintf("(score %d, %s)", f.Score, f.Confidence), opts),
		)

		if f.Section != "" {
			fmt.Fprintf(w, "  %s %s\n", muted("section:", opts), f.Section)
		}
		if f.Workload != "" {
			fmt.Fprintf(w, "  %s %s\n", muted("workload:", opts), f.Workload)
		}

		if f.LineNo > 0 {
			fmt.Fprintf(w, "  %s\n", muted(fmt.Sprintf("line %d:", f.LineNo), opts))
		}
		if f.Line != "" {
			fmt.Fprintf(w, "  %s\n", f.Line)
		}
		if f.Hint != "" {
			fmt.Fprintf(w, "  %s %s\n", muted("hint:", opts), f.Hint)
		}
		fmt.Fprintln(w)
	}

	if !opts.ShowAll && len(analysis.Findings) > len(findings) {
		fmt.Fprintf(w, "Showing top %d findings. Re-run with --show-all to print everything.\n\n", len(findings))
	}

	fmt.Fprintln(w, "Next checks")
	for _, hint := range nextChecks(findings) {
		fmt.Fprintf(w, "  - %s\n", hint)
	}
}

func renderJSON(w io.Writer, analysis analyzer.Analysis) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(analysis)
}

func renderMarkdown(w io.Writer, analysis analyzer.Analysis) {
	fmt.Fprintln(w, "# kdescribe report")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "- Risk: **%s** (score %d)\n", analysis.Risk.Level, analysis.Risk.Score)
	fmt.Fprintf(w, "- Findings: %d\n", len(analysis.Findings))
	fmt.Fprintln(w)
	fmt.Fprintln(w, "## Findings")
	fmt.Fprintln(w)

	if len(analysis.Findings) == 0 {
		fmt.Fprintln(w, "No obvious issues found.")
		return
	}

	for _, f := range analysis.Findings {
		fmt.Fprintf(w, "### %s: %s\n\n", f.Severity, f.Message)
		fmt.Fprintf(w, "- Score: `%d`\n", f.Score)
		fmt.Fprintf(w, "- Domain: `%s`\n", f.Domain)
		if f.Workload != "" {
			fmt.Fprintf(w, "- Workload: `%s`\n", f.Workload)
		}
		if f.Section != "" {
			fmt.Fprintf(w, "- Section: `%s`\n", f.Section)
		}
		if f.LineNo > 0 {
			fmt.Fprintf(w, "- Line: `%d`\n", f.LineNo)
		}
		if f.Hint != "" {
			fmt.Fprintf(w, "- Hint: %s\n", f.Hint)
		}
		if f.Line != "" {
			fmt.Fprintf(w, "\n```text\n%s\n```\n", f.Line)
		}
		fmt.Fprintln(w)
	}
}

func visibleFindings(findings []analyzer.Finding, opts Options) []analyzer.Finding {
	if opts.ShowAll || len(findings) <= 10 {
		return findings
	}

	return findings[:10]
}

func severityLabel(severity analyzer.Severity, opts Options) string {
	label := "[" + string(severity) + "]"
	if opts.NoColor {
		return label
	}

	switch severity {
	case analyzer.Critical:
		return criticalStyle.Render(label)
	case analyzer.Warning:
		return warningStyle.Render(label)
	default:
		return infoStyle.Render(label)
	}
}

func renderRisk(risk analyzer.Risk, opts Options) string {
	if opts.NoColor {
		if risk.Score == 0 {
			return risk.Level
		}
		return fmt.Sprintf("%s (score %d)", risk.Level, risk.Score)
	}

	level := risk.Level
	if risk.Score > 0 {
		level = fmt.Sprintf("%s (score %d)", risk.Level, risk.Score)
	}

	switch risk.Level {
	case "HIGH":
		return criticalStyle.Render(level)
	case "MEDIUM":
		return warningStyle.Render(level)
	default:
		return infoStyle.Render(level)
	}
}

func muted(text string, opts Options) string {
	if opts.NoColor {
		return text
	}

	return mutedStyle.Render(text)
}

func nextChecks(findings []analyzer.Finding) []string {
	seen := make(map[string]bool)
	checks := make([]string, 0, 3)

	for _, finding := range findings {
		if finding.Hint == "" || seen[finding.Hint] {
			continue
		}

		seen[finding.Hint] = true
		checks = append(checks, finding.Hint)
		if len(checks) == 3 {
			break
		}
	}

	if len(checks) == 0 {
		return []string{"Review Events, container state, and recent logs for correlated failures."}
	}

	return checks
}
