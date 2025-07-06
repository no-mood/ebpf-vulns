package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"xvtlas/report"
	"xvtlas/config"
	"path/filepath"
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
//Still need checks for some edge cases
func RunMakeAll(rootPath string) (string, error) {
	cmd := exec.Command("make", "-C", rootPath)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

//the .o file should be found in each folder
func FindObjectFile(folderPath string) string { 
	files, _ := os.ReadDir(folderPath)
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".o" {
			return filepath.Join(folderPath, file.Name())
		}
	}
	return ""
}

func RunVerifier(oFile, cFile, prettyPath string, row *report.CSVRow, cfg *config.EBPFConfig) {
	progName := cfg.EBPFProgram.Name
	fmt.Println(progName)
	if progName == "" {
		progName = filepath.Base(oFile)
	}

	pinPath := progName

	cmdStr := fmt.Sprintf("sudo bpftool prog load %s /sys/fs/bpf/%s 2>&1 | python3 %s -c %s", oFile, pinPath, prettyPath, cFile)
	fmt.Println("Comnmand : ",cmdStr)
	cmd := exec.Command("bash", "-c", cmdStr)
	output, err := cmd.CombinedOutput()
	//fmt.Println("Verifier output: ",output)
	//fmt.Println("Verifier errors: ", err)
	row.Verified = err == nil
	row.LoadOutput += string(output)
}

func LoadEBPF(oFile string, cfg *config.EBPFConfig, row *report.CSVRow) {
	//1) take as input the compile_object
	//2) use target from config_to try attach if configured ()
	//3) NO need for this now

	row.Loaded = true
}



//Function to run make on the single folder in order to compile only the actively patched file
func RunMake(rootPath string) (string, error) {
	cmd := exec.Command("make", "-C", rootPath)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
 
