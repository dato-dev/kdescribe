package renderer

import (
	"fmt"

	"github.com/dato-dev/kdescribe/internal/analyzer"

	"github.com/charmbracelet/lipgloss"
)

var (
	criticalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6"))
)

func Render(findings []analyzer.Finding) {
	if len(findings) == 0 {
		fmt.Println("✅ No obvious issues found")
		return
	}

	fmt.Println()
	fmt.Println("⚠ Important findings")
	fmt.Println()

	for _, f := range findings {
		switch f.Severity {
		case analyzer.Critical:
			fmt.Printf(
				"%s %s\n",
				criticalStyle.Render("[CRITICAL]"),
				f.Message,
			)

		case analyzer.Warning:
			fmt.Printf(
				"%s %s\n",
				warningStyle.Render("[WARNING]"),
				f.Message,
			)

		default:
			fmt.Printf(
				"%s %s\n",
				infoStyle.Render("[INFO]"),
				f.Message,
			)
		}

		fmt.Printf("  %s\n\n", f.Line)
	}
}
