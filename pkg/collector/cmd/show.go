package cmd

import (
	"context"
	"encoding/json"
	"os"

	"github.com/cybozu-go/setup-hw/config"
	"github.com/cybozu-go/setup-hw/gabs"
	"github.com/cybozu-go/setup-hw/redfish"
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
			var data map[string]*gabs.Container

			if len(inputFile) == 0 {
				ac, uc, err := config.LoadConfig()
				if err != nil {
					return err
				}

				cc := &redfish.ClientConfig{
					AddressConfig: ac,
					UserConfig:    uc,
					NoEscape:      true,
				}
				client, err := redfish.NewRedfishClient(cc)
				if err != nil {
					return err
				}

				rule := &redfish.CollectRule{
					TraverseRule: redfish.TraverseRule{
						Root: "/redfish/v1",
					},
				}

				collected := client.Traverse(ctx, rule)
				data = collected.Data()
			} else {
				var input map[string]interface{}
				f, err := os.Open(inputFile)
				if err != nil {
					return err
				}

				err = json.NewDecoder(f).Decode(&input)
				if err != nil {
					return err
				}

				data = make(map[string]*gabs.Container)
				for k, v := range input {
					c, err := gabs.Consume(v)
					if err != nil {
						return err
					}
					data[k] = c
				}
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
