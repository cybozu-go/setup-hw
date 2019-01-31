package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var jsonOutput bool

var rootCmd = &cobra.Command{
	Use:   "idrac-passwd-hash",
	Short: "Generate hashed password for iDRAC",
	Long: `Generate hashed password and random salt for iDRAC.

This tool asks a password and outputs hashed password with
a generated random salt as described in Dell manual:
https://www.dell.com/support/manuals/us/en/04/poweredge-r940/idrac_3.15.15.15_ug/generating-hash-password-without-snmpv3-and-ipmi-authentication?guid=guid-e4486863-89bc-4b0c-9578-ff564fade424&lang=en-us`,

	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		passwd, err := askPassword()
		if err != nil {
			return err
		}

		salt, err := generateSalt()
		if err != nil {
			return err
		}

		hash, err := hashPassword(passwd, salt)
		if err != nil {
			return err
		}

		if jsonOutput {
			return outputJSON(hash, salt)
		}
		return outputPlain(hash, salt)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVar(&jsonOutput, "json", false, "output hash and salt in JSON")
}
