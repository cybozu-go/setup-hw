package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var keys []string

// generateRuleCmd represents the generateRule command
var generateRuleCmd = &cobra.Command{
	Use:   "generate-rule",
	Short: "output a collection rule to collect specified keys as metrics",
	Long:  `output a collection rule to collect specified keys as metrics.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("generateRule called")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateRuleCmd)

	generateRuleCmd.Flags().StringSliceVar(&keys, "key", nil, "Redfish data key to find")
}
