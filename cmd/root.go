package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
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
	help     bool
}

var rootCmd = &cobra.Command{
	Use:                "kdescribe [resource] [name] [kubectl describe flags]",
	Short:              "Kubernetes describe highlighter",
	Args:               cobra.ArbitraryArgs,
	DisableFlagParsing: true,
	SilenceUsage:       true,
	SilenceErrors:      true,
	Example:            "  kubectl kdescribe pod nginx\n  kubectl kdescribe pod nginx -n default\n  kubectl describe pod nginx | kdescribe --output json",
	RunE: func(cmd *cobra.Command, args []string) error {
		describeArgs, err := parseArgs(args)
		if err != nil {
			return err
		}

		if options.help {
			_ = cmd.Help()
			return nil
		}

		input, err := readInput(describeArgs)
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

func readInput(describeArgs []string) ([]byte, error) {
	if len(describeArgs) > 0 {
		args := append([]string{"describe"}, describeArgs...)
		command := exec.Command("kubectl", args...)
		command.Stderr = os.Stderr

		return command.Output()
	}

	if hasNoStdin() {
		return nil, nil
	}

	return io.ReadAll(os.Stdin)
}

func parseArgs(args []string) ([]string, error) {
	describeArgs := make([]string, 0, len(args))

	for index := 0; index < len(args); index++ {
		arg := args[index]

		switch {
		case arg == "--":
			describeArgs = append(describeArgs, args[index+1:]...)
			return describeArgs, nil
		case arg == "--help" || arg == "-h":
			options.help = true
		case arg == "--no-color":
			options.noColor = true
		case arg == "--exit-code":
			options.exitCode = true
		case arg == "--show-all":
			options.showAll = true
		case arg == "--output" || arg == "-o":
			value, ok := nextValue(args, &index)
			if !ok {
				return nil, fmt.Errorf("%s requires a value", arg)
			}
			options.output = value
		case strings.HasPrefix(arg, "--output="):
			options.output = strings.TrimPrefix(arg, "--output=")
		case arg == "--min-score":
			value, ok := nextValue(args, &index)
			if !ok {
				return nil, fmt.Errorf("%s requires a value", arg)
			}
			minScore, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid --min-score value %q", value)
			}
			options.minScore = minScore
		case strings.HasPrefix(arg, "--min-score="):
			value := strings.TrimPrefix(arg, "--min-score=")
			minScore, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid --min-score value %q", value)
			}
			options.minScore = minScore
		default:
			describeArgs = append(describeArgs, arg)
		}
	}

	return describeArgs, nil
}

func nextValue(args []string, index *int) (string, bool) {
	if *index+1 >= len(args) {
		return "", false
	}

	*index = *index + 1
	return args[*index], true
}
