package cmd

import (
	"fmt"
	"os"
	"xvtlas/ebpf"
	"xvtlas/logger"
	"github.com/spf13/cobra"
)

var (
	rootPath      string
	interactive   bool
	verbose       bool
	exportPath    string
	prettyPath    string
	kernelVersion string
)

var rootCmd = &cobra.Command{
	Use:   "xvtlas",
	Short: "XVTlas - XDP Verifier Launch Automation Suite ",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Init(verbose, exportPath)
		ebpf.RunPipeline(rootPath, prettyPath, kernelVersion, exportPath, interactive)
	},
}

func Execute() {
	rootCmd.Flags().StringVarP(&rootPath, "path", "p", ".", "Root path to eBPF program folders")
	rootCmd.Flags().BoolVar(&interactive, "interactive", false, "Stop on errors and prompt user to continue")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.Flags().StringVarP(&exportPath, "export", "e", "./output", "Export location for logs and CSV")
	rootCmd.Flags().StringVar(&prettyPath, "pretty", "", "Path to pretty verifier")
	rootCmd.Flags().StringVar(&kernelVersion, "kernel", "", "Target kernel version for accounting")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

