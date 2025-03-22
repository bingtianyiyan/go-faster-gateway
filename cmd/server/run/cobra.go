package run

import (
	"fmt"
	"go-faster-gateway/cmd/server/api"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "faster-gateway",
	Short:        "faster-gateway",
	SilenceUsage: true,
	Long:         `faster-gateway`,
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	PersistentPreRunE: func(*cobra.Command, []string) error { return nil },
	Run: func(cmd *cobra.Command, args []string) {
		tip()
	},
}

func tip() {
	usageStr := `welcome faster-gateway`
	fmt.Printf("%s\n", usageStr)
}

func init() {
	rootCmd.AddCommand(api.StartCmd)
}

// Execute : apply commands
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
