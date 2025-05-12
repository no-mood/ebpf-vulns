package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"xvtlas/report"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func ConfirmPrompt(msg string) bool {
	fmt.Print(msg + " (y/n): ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	return response == "y\n"
}

func CompileEBPF(path string, cfg interface{}, row *report.CSVRow) string {
	oFile := path[:len(path)-2] + ".o"
	cmd := exec.Command("clang", "-O2", "-target", "bpf", "-c", path, "-o", oFile)
	output, err := cmd.CombinedOutput()
	row.Compiled = err == nil
	row.LoadOutput += string(output)
	return oFile
}

func RunVerifier(oFile, cFile, prettyPath string, row *report.CSVRow) {
	cmdStr := fmt.Sprintf("bpftool prog load %s /sys/fs/bpf/%s 2>&1 | python3 %s -c %s", oFile, filepath.Base(oFile), prettyPath, cFile)
	cmd := exec.Command("bash", "-c", cmdStr)
	output, err := cmd.CombinedOutput()
	row.Verified = err == nil
	row.LoadOutput += string(output)
}

func LoadEBPF(oFile string, cfg *config.EBPFConfig, row *report.CSVRow) {
	// TODO add loads when I figure out how to compile out of three
	row.Loaded = true
}

