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
	patchPath     string
)

var rootCmd = &cobra.Command{
	Use:   "xvtlas",
	Short: "XVTlas - XDP Verifier Launch Automation Suite ",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Init(verbose, exportPath)
		CastStatus(rootPath, prettyPath, kernelVersion, exportPath, interactive)
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
	rootCmd.Flags().StringVar(&patchPath, "patch-path", "", "Path to folders containing patches and configs")

	rootCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if patchPath != "" && rootPath != "." {
			return fmt.Errorf("--patch-path and --path are mutually exclusive")
		}
		return nil
	}

	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		logger.Init(verbose, exportPath)
		if patchPath != "" {
			ebpf.RunPatchPipeline(patchPath, prettyPath, kernelVersion, exportPath, interactive)
		} else {
			ebpf.RunPipeline(rootPath, prettyPath, kernelVersion, exportPath, interactive)
		}
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func CastStatus( rootPath string, prettyPath string, kernelVersion string, exportPath string, interactive bool) {
	fmt.Println("Root path : ", rootPath)
	fmt.Println("Pretty : ", prettyPath)
	fmt.Println("Kernel : ", kernelVersion)
	fmt.Println("Export : ", exportPath)
	fmt.Println("Interactive : ", interactive)
}
