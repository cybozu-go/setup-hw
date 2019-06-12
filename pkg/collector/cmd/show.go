package cmd

import (
	"context"
	"encoding/json"
	"os"

	"github.com/cybozu-go/well"
	"github.com/spf13/cobra"
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "show collected Redfish data",
	Long:  `show collected Redfish data`,

	RunE: func(cmd *cobra.Command, args []string) error {
		well.Go(func(ctx context.Context) error {
			data, err := collectOrLoad(ctx, inputFile)
			if err != nil {
				return err
			}

			result := make(map[string]interface{})
			for k, v := range data {
				result[k] = v.Data()
			}

			out, err := json.MarshalIndent(result, "", "    ")
			if err != nil {
				return err
			}
			_, err = os.Stdout.Write(out)
			if err != nil {
				return err
			}
			return nil
		})

		well.Stop()
		err := well.Wait()
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}
