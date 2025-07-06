package ebpf

import (
	"fmt"
	"os"
	"os/exec"
	"io"
	"path/filepath"
	"xvtlas/config"
	"xvtlas/logger"
	"xvtlas/report"
	"xvtlas/utils"
)
func RunPipeline(rootPath, prettyPath, kernelVersion, exportPath string, interactive bool) {
	var rows []report.CSVRow

	// Global compilation
	compilationLog, err := utils.RunMakeAll(rootPath)
	fmt.Println("Stuff is compiled :", compilationLog)
	if err != nil {
		logger.LogError("Makefile", compilationLog)
		if interactive && !utils.ConfirmPrompt("Compilation failed. Continue?") {
			os.Exit(0)
		}
	}

	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}

		ebpfFile := filepath.Join(path, "main.c")
		yamlFile := filepath.Join(path, "config.yaml")

		if !utils.FileExists(ebpfFile) || !utils.FileExists(yamlFile) {
			return nil
		}

		cfg, err := config.ParseConfig(yamlFile)
		if err != nil {
			logger.LogError(ebpfFile, err.Error())
			if interactive && !utils.ConfirmPrompt("Continue to next program?") {
				os.Exit(0)
			}
			return nil
		}

		row := report.CSVRow{
			Filename:        ebpfFile,
			LoadParameters:  fmt.Sprintf("%v", cfg.EBPFProgram),
			KernelVersion:   kernelVersion,
			Compiled:        true, // assume success if object exists ??
		}

		oFile := utils.FindObjectFile(path)
		if oFile == "" {
			row.Compiled = false
			row.LoadOutput += "No .o file found after make.\n"
		} else {
			utils.RunVerifier(oFile, ebpfFile, prettyPath, &row, cfg)
			//utils.LoadEBPF(oFile, cfg, &row)
		}

		logger.SaveLog(ebpfFile, row.LoadOutput)
		rows = append(rows, row)
		return nil
	})

	if err != nil {
		fmt.Println("Error scanning root path:", err)
	}

	report.ExportCSV(rows, exportPath)
}


func RunPipelinePatch(patchRoot, baseFile, prettyPath, kernelVersion, exportPath string, interactive, removePatched bool) {
	var rows []report.CSVRow

	submoduleRoot := filepath.Dir(baseFile)
	fmt.Println("Running pipeline")

	err := filepath.Walk(patchRoot, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}

		patchFile := filepath.Join(path, "patch.diff")
		yamlFile := filepath.Join(path, "config.yaml")
		if !utils.FileExists(patchFile) || !utils.FileExists(yamlFile) {
			return nil
		}

		// Ensure submodule base file is reset to clean state
		resetCmd := exec.Command("git", "-C", submoduleRoot, "reset", "--hard")
		if out, err := resetCmd.CombinedOutput(); err != nil {
			logger.LogError(baseFile, fmt.Sprintf("Failed to reset repo state: %s\nOutput: %s", err.Error(), out))
			return nil
		}

		// Convert patch file to absolute path to avoid path resolution issues
		absPatchFile, err := filepath.Abs(patchFile)
		if err != nil {
			logger.LogError(patchFile, fmt.Sprintf("Failed to get absolute path: %s", err.Error()))
			return nil
		}

		// Apply patch using git apply with absolute path from submodule root
		applyCmd := exec.Command("git", "-C", submoduleRoot, "apply", absPatchFile)
		if output, err := applyCmd.CombinedOutput(); err != nil {
			logger.LogError(patchFile, fmt.Sprintf("Failed to apply patch: %s\nOutput: %s", err.Error(), output))
			if interactive && !utils.ConfirmPrompt("Continue after failed patch?") {
				os.Exit(1)
			}
			return nil
		}

		// Copy patched source to test folder as main.c
		patchedDest := filepath.Join(path, "main.c")
		srcFile, err := os.Open(baseFile)
		if err != nil {
			logger.LogError(baseFile, "Failed to open patched base file")
			return nil
		}
		defer srcFile.Close()

		dstFile, err := os.Create(patchedDest)
		if err != nil {
			logger.LogError(patchedDest, "Failed to create target patched file")
			return nil
		}
		defer dstFile.Close()

		if _, err = io.Copy(dstFile, srcFile); err != nil {
			logger.LogError(patchedDest, "Failed to copy patched source")
			return nil
		}

		// Compile with make in target dir
		compilationLog, err := utils.RunMake(path)
		if err != nil {
			logger.LogError("Makefile", compilationLog)
			if interactive && !utils.ConfirmPrompt("Compilation failed. Continue?") {
				os.Exit(1)
			}
		}

		cfg, err := config.ParseConfig(yamlFile)
		if err != nil {
			logger.LogError(yamlFile, err.Error())
			return nil
		}

		row := report.CSVRow{
			Filename:        patchedDest,
			LoadParameters:  fmt.Sprintf("%v", cfg.EBPFProgram),
			KernelVersion:   kernelVersion,
			Compiled:        true,
		}

		oFile := utils.FindObjectFile(path)
		if oFile == "" {
			row.Compiled = false
			row.LoadOutput += "No .o file found after make.\n"
		} else {
			utils.RunVerifier(oFile, patchedDest, prettyPath, &row, cfg)
			row.Loaded = row.Verified
		}

		logger.SaveLog(patchedDest, row.LoadOutput)
		rows = append(rows, row)

		// Optionally remove main.c copy
		if removePatched {
			_ = os.Remove(patchedDest)
		}

		// Unpin loaded program if necessary
		progName := cfg.EBPFProgram.Name
		if progName == "" && oFile != "" {
			progName = filepath.Base(oFile)
		}
		_ = exec.Command("sudo", "rm", "-f", filepath.Join("/sys/fs/bpf", progName)).Run()

		return nil
	})

	if err != nil {
		fmt.Println("Error scanning patch path:", err)
	}

	report.ExportCSV(rows, exportPath)
}

