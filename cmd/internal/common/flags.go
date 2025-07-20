package common

import "github.com/spf13/cobra"

// AddCommonFlags adds verbose and debug flags that are shared by all commands
func AddCommonFlags(cmd *cobra.Command, verbose, debug *bool) {
	cmd.PersistentFlags().BoolVarP(verbose, "verbose", "v", false, "Enable verbose output")
	cmd.PersistentFlags().BoolVar(debug, "debug", false, "Enable debug output (includes technical details)")
}
