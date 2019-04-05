package cmd

import (
	"context"
	"errors"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/setup-hw/config"
	"github.com/cybozu-go/setup-hw/lib"
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
		var ruleFile string
		switch vendor {
		case lib.QEMU:
			monitor = monitorQEMU
			ruleFile = "/qemu.yml"
		case lib.Dell:
			monitor = monitorDell
			ruleFile = "/dell_redfish_1.0.2.yml"
		default:
			return errors.New("unsupported vendor hardware")
		}

		err = startExporter(ac, uc, ruleFile)
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
