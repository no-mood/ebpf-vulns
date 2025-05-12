package ebpf

import (
	"fmt"
	"os"
	"path/filepath"
	"xvtlas/config"
	"xvtlas/logger"
	"xvtlas/report"
	"xvtlas/utils"
)

func RunPipeline(rootPath, prettyPath, kernelVersion, exportPath string, interactive bool) {
	var rows []report.CSVRow

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
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
		}

		oFile := utils.CompileEBPF(ebpfFile, cfg, &row)
		if oFile != "" {
			utils.RunVerifier(oFile, ebpfFile, prettyPath, &row)
			utils.LoadEBPF(oFile, cfg, &row)
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

