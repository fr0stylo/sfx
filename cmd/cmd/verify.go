package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/fr0stylo/sfx/config"
)

func init() {
	RegisterSubCommand(newVerifyCommand())
}

func newVerifyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify the sfx configuration",
		Long:  "Load the .sfx.yaml configuration and report validation issues.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("load configuration: %w", err)
			}

			if err := config.Validate(cfg); err != nil {
				return fmt.Errorf("configuration invalid: %w", err)
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "configuration is valid")
			return err
		},
	}

	return cmd
}
