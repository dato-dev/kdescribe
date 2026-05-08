package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dato-dev/kdescribe/internal/analyzer"
	"github.com/dato-dev/kdescribe/internal/renderer"
	"github.com/spf13/cobra"
)

var options struct {
	output   string
	noColor  bool
	minScore int
	exitCode bool
	showAll  bool
}

var rootCmd = &cobra.Command{
	Use:           "kdescribe",
	Short:         "Kubernetes describe highlighter",
	SilenceUsage:  true,
	SilenceErrors: true,
	Example:       "  kubectl describe pod nginx | kdescribe\n  kubectl describe pod nginx | kdescribe --output json",
	RunE: func(cmd *cobra.Command, args []string) error {
		if hasNoStdin() {
			_ = cmd.Help()
			return nil
		}

		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}

		if strings.TrimSpace(string(input)) == "" {
			_ = cmd.Help()
			return nil
		}

		analysis := analyzer.AnalyzeDocument(string(input))
		analysis.Findings = analyzer.FilterFindings(analysis.Findings, options.minScore)
		analysis.Risk = analyzer.CalculateRisk(analysis.Findings)

		err = renderer.Render(os.Stdout, analysis, renderer.Options{
			Format:  options.output,
			NoColor: options.noColor,
			ShowAll: options.showAll,
		})
		if err != nil {
			return err
		}

		if options.exitCode && len(analysis.Findings) > 0 {
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	rootCmd.Flags().StringVarP(&options.output, "output", "o", renderer.FormatHuman, "output format: human, json, markdown")
	rootCmd.Flags().BoolVar(&options.noColor, "no-color", false, "disable ANSI colors")
	rootCmd.Flags().IntVar(&options.minScore, "min-score", 0, "only show findings with score >= value")
	rootCmd.Flags().BoolVar(&options.exitCode, "exit-code", false, "exit with status 1 when findings are present")
	rootCmd.Flags().BoolVar(&options.showAll, "show-all", false, "show all findings instead of the top results")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func hasNoStdin() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	return stat.Mode()&os.ModeCharDevice != 0
}
