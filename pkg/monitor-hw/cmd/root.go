package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/setup-hw/config"
	"github.com/cybozu-go/setup-hw/lib"
	"github.com/cybozu-go/setup-hw/redfish"
	"github.com/cybozu-go/well"
	"github.com/spf13/cobra"
)

var opts struct {
	listenAddress string
	redfishRoot   string
	interval      int
	resetInterval int
}

const (
	defaultPort          = ":9105"
	defaultInterval      = 60
	defaultResetInterval = 24
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "monitor-hw",
	Short: "monitor-hw daemon",
	Long:  `monitor-hw is a daemon to monitor server statuses.`,

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Validation of arguments has finished, so disable the usage message.
		cmd.SilenceUsage = true

		err := well.LogConfig{}.Apply()
		if err != nil {
			log.ErrorExit(err)
		}
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		ac, uc, err := config.LoadConfig()
		if err != nil {
			return err
		}

		vendor, err := lib.DetectVendor()
		if err != nil {
			return err
		}

		var monitor func(context.Context) error
		ruleGetter, client, err := func(vendor lib.Vendor) (redfish.RuleGetter, redfish.Client, error) {
			switch vendor {
			case lib.QEMU:
				monitor = monitorQEMU
				ruleFile := "qemu.yml"
				rule, ok := redfish.Rules[ruleFile]
				if !ok {
					return nil, nil, errors.New("unknown rule file: " + ruleFile)
				}

				client := redfish.NewMockClient(redfish.DummyRedfishFile)

				return func() (*redfish.CollectRule, error) {
					return rule, nil
				}, client, nil
			case lib.Dell:
				monitor = monitorDell
				endpoint, err := url.Parse("https://" + ac.IPv4.Address)
				if err != nil {
					return nil, nil, err
				}

				ruleGetter := func() (*redfish.CollectRule, error) {
					version, err := lib.DetectRedfishVersion(endpoint, uc)
					if err != nil {
						return nil, err
					}
					ruleFile := fmt.Sprintf("dell_redfish_%s.yml", version)
					rule, ok := redfish.Rules[ruleFile]
					if !ok {
						return nil, errors.New("unknown rule file: " + ruleFile)
					}
					return rule, nil
				}

				cc := &redfish.ClientConfig{
					AddressConfig: ac,
					UserConfig:    uc,
				}

				client, err := redfish.NewRedfishClient(cc)
				if err != nil {
					return nil, nil, err
				}

				return ruleGetter, client, nil
			default:
				return nil, nil, errors.New("unsupported vendor hardware")
			}
		}(vendor)
		if err != nil {
			return err
		}

		err = startExporter(ruleGetter, client)
		if err != nil {
			return err
		}

		well.Go(func(ctx context.Context) error {
			return monitor(ctx)
		})

		well.Stop()
		err = well.Wait()
		if err != nil && !well.IsSignaled(err) {
			return err
		}
		return nil
	},
}

// Execute executes monitor-hw
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.ErrorExit(err)
	}
}

func init() {
	rootCmd.Flags().StringVar(&opts.listenAddress, "listen", defaultPort, "listening address and port number")
	rootCmd.Flags().IntVar(&opts.interval, "interval", defaultInterval, "interval of collecting metrics in seconds")
	rootCmd.Flags().IntVar(&opts.resetInterval, "reset-interval", defaultResetInterval, "interval of resetting iDRAC in hours (dell servers only)")
}
