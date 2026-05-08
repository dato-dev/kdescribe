package cmd

import (
	"io"
	"os"

	"github.com/dato-dev/kdescribe/internal/analyzer"
	"github.com/dato-dev/kdescribe/internal/renderer"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kdescribe",
	Short: "Kubernetes describe highlighter",
	RunE: func(cmd *cobra.Command, args []string) error {
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}

		findings := analyzer.Analyze(string(input))

		renderer.Render(findings)

		return nil
	},
}

func Execute() {
	_ = rootCmd.Execute()
}
