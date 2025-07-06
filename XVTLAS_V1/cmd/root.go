package cmd

import (
	"fmt"
	"os"
	"xvtlas/ebpf"
	"xvtlas/logger"

	"github.com/spf13/cobra"
)

var (
	rootPath       string
	interactive    bool
	verbose        bool
	exportPath     string
	prettyFilePath     string
	kernelVersion  string
	patchPath      string
	baseFile       string
	keepPatched    bool
)

var rootCmd = &cobra.Command{
	Use:   "xvtlas",
	Short: "XVTlas - XDP Verifier Launch Automation Suite",
	Long:  "XVTlas is a tool to compile, patch, load, and verify eBPF programs using bpftool and a pretty verifier.",
	SilenceUsage:  true,
	SilenceErrors: true,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(os.Args) == 1 {
			cmd.Help()
			return fmt.Errorf("no arguments provided")
		}
		if patchPath != "" && rootPath != "." {
			return fmt.Errorf("--patch-path and --path are mutually exclusive")
		}
		if patchPath != "" && baseFile == "" {
			return fmt.Errorf("--base-file must be specified when using --patch-path")
		}
		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		logger.Init(verbose, exportPath)

		if patchPath != "" {
			removePatched := !keepPatched
			ebpf.RunPipelinePatch(patchPath, baseFile, prettyFilePath, kernelVersion, exportPath, interactive, removePatched)
		} else {
			ebpf.RunPipeline(rootPath, prettyFilePath, kernelVersion, exportPath, interactive)
		}
	},
}

func Execute() {
	rootCmd.Flags().StringVarP(&rootPath, "path", "p", ".", "Root path to eBPF program folders")
	rootCmd.Flags().BoolVar(&interactive, "interactive", false, "Stop on errors and prompt user to continue")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.Flags().StringVarP(&exportPath, "export", "e", "./output", "Export location for logs and CSV")
	rootCmd.Flags().StringVar(&prettyFilePath, "pretty", "", "Pretty verifier file path")
	rootCmd.Flags().StringVar(&kernelVersion, "kernel", "", "Target kernel version for accounting")
	rootCmd.Flags().StringVar(&patchPath, "patch-path", "", "Path to folders containing patches and configs")
	rootCmd.Flags().StringVar(&baseFile, "base-file", "", "Absolute path to the master file to apply each patch to")
	rootCmd.Flags().BoolVar(&keepPatched, "keep-patched", false, "Keep the patched file after running (default: false)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

