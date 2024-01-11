package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/setup-hw/config"
	"github.com/cybozu-go/setup-hw/lib"
	"github.com/cybozu-go/setup-hw/redfish"
	"github.com/cybozu-go/well"
	"github.com/spf13/cobra"
)

var opts struct {
	listenAddress string
	interval      int
	noResetFile   string
}

const (
	defaultAddress  = ":9105"
	defaultInterval = 60
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

		var client redfish.Client
		var ruleGetter redfish.RuleGetter
		switch vendor {
		case lib.QEMU:
			client = redfish.NewMockClient(redfish.DummyRedfishFile)
			ruleFile := "qemu.yml"
			rule, ok := redfish.Rules[ruleFile]
			if !ok {
				return errors.New("unknown rule file: " + ruleFile)
			}
			ruleGetter = func(context.Context) (*redfish.CollectRule, error) {
				return rule, nil
			}

		case lib.Dell:
			if err := initDell(cmd.Context()); err != nil {
				return err
			}
			cc := &redfish.ClientConfig{
				AddressConfig: ac,
				UserConfig:    uc,
				NoEscape:      true,
			}
			cl, err := redfish.NewRedfishClient(cc)
			if err != nil {
				return err
			}
			client = cl
			ruleGetter = func(ctx context.Context) (*redfish.CollectRule, error) {
				version, err := cl.GetVersion(ctx)
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

		default:
			return errors.New("unsupported vendor hardware")
		}

		if err := startExporter(ruleGetter, client); err != nil {
			return err
		}

		well.Stop()
		if err := well.Wait(); err != nil && !well.IsSignaled(err) {
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
	rootCmd.Flags().StringVar(&opts.listenAddress, "listen", defaultAddress, "listening address and port number")
	rootCmd.Flags().IntVar(&opts.interval, "interval", defaultInterval, "interval of collecting metrics in seconds")
}
