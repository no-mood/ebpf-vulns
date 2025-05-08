/*
**XDP Verifier Test Launch Automation Suite**

This tool manages batch tests for XDP or EBPF programs, the program flows as follows:
- Data input : the test source directory containing the XDP/eBPF tests to load, with
	the corresponding metadata files, metadata files sintax is explained in the readme
- result of the tests are displayed in the CLI result and can be exported to LaTex or Markdown
*/

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

//Structs for YAML unmarshaller

type EBPFMetadataItem struct {
	EBPFProgram struct {
		Name          string `yaml:"name"`
		Type          string `yaml:"type"`
		AttachPoint   string `yaml:"attach_point"`
		KernelVersion string `yaml:"kernel_version"`
		VerifierOpts  struct {
			RelaxedChecking bool `yaml:"relaxed_checking"`
			AllowTailCalls  bool `yaml:"allow_tail_calls"`
			StackSizeLimit  int  `yaml:"stack_size_limit"`
		} `yaml:"verifier_options"`
		Maps []struct {
			Name       string `yaml:"name"`
			Type       string `yaml:"type"`
			KeySize    int    `yaml:"key_size"`
			ValueSize  int    `yaml:"value_size"`
			MaxEntries int    `yaml:"max_entries"`
		} `yaml:"maps"`
		Features     []string `yaml:"features"`
		Dependencies struct {
			ClangVersion string `yaml:"clang_version"`
			LLVMVersion  string `yaml:"llvm_version"`
			LibBpf       string `yaml:"libbpf"`
		} `yaml:"dependencies"`
		Loading struct {
			Pinning    bool `yaml:"pinning"`
			AutoAttach bool `yaml:"auto_attach"`
		} `yaml:"loading"`
		Testing struct {
			RuntimeTest      bool   `yaml:"runtime_test"`
			PacketReplay     string `yaml:"packet_replay"`
			ExpectedBehavior string `yaml:"expected_behavior"`
		} `yaml:"testing"`
	} `yaml:"ebpf_program"`
}


func customUsage() {

	fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\nRun `xvtlas -h` for details on each flag.")

}




func main() {

	//flags
	input := flag.String("input", "Dir", "Specify the target directory for input")
	kernel := flag.String("kernel", "6.8", "Linux Kernel Version to use for reference in the report")
	verbose := flag.Bool("verbose", false, "Enable verbose output")	
	export := flag.String("export", "false" , "Directory target for export of LaTex report")
	pretty := flag.String("pretty", "/usr/local", "Directory of Prettyverifier installation")
	showHelp := flag.Bool("help", false, "Display help")

	//Set custom usage functions
	flag.Usage = customUsage
	flag.Parse()

	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	
	//Main logic
	if *verbose {
		fmt.Println("Verbose mode ON")
	}
	fmt.Printf("XVTLAS starting, target Kernel: %s\n", *kernel)
	fmt.Printf("Input Directory: %s\n", *input)
	fmt.Printf("Export Directory: %s\n", *export)
	fmt.Printf("PrettyVerifier Directory: %s\n", *pretty)

	err := processInput(*input)

	fmt.Println(err)

}


func processInput(dirInput string) (error){

	var dicMetaData = make(map[string]*EBPFMetadataItem)
	//Function that reads the input directory lists the tuples that are going to be processed
	//For each corresponding tuple eBPF program and YAML config file ...
	// the YAML is passed to the parser and returned as object, the object is 
	// then passed to the EBPF launcher that will try to load it
// Lists to track C and YAML files
	cFiles := make(map[string]bool)
	yamlFiles := make(map[string]bool)

	// Read all files in the directory
	files, err := os.ReadDir(dirInput)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Collect C and YAML files
	for _, file := range files {
		ext := path.Ext(file.Name())
		baseName := strings.TrimSuffix(file.Name(), ext)

		if ext == ".c" {
			cFiles[baseName] = true
		} else if ext == ".yaml" {
			yamlFiles[baseName] = true
		}
	}

	// Check if each .c file has a corresponding .yaml file
	for cFile := range cFiles {
		if !yamlFiles[cFile] {
			return fmt.Errorf("missing YAML file for C program: %s.c", cFile)
		}
	}

	// Check if each .yaml file has a corresponding .c file
	for yamlFile := range yamlFiles {
		if !cFiles[yamlFile] {
			return fmt.Errorf("missing C file for YAML configuration: %s.yaml", yamlFile)
		}
	}

	// Process YAML files
	for yamlFile := range yamlFiles {
		yamlPath := path.Join(dirInput, yamlFile+".yaml")
		metaData, err := parseYaml(yamlPath)
		if err != nil {
			return fmt.Errorf("failed to parse YAML file %s: %w", yamlFile, err)
		}
		dicMetaData[yamlFile] = metaData
	}

	fmt.Println("Successfully processed all eBPF programs.")
	return nil
}

// parseYaml reads and parses a YAML file into an EBPFMetadataItem object
func parseYaml(inputYAML string) (*EBPFMetadataItem, error) {
	data, err := os.ReadFile(inputYAML)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	var metadata EBPFMetadataItem
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &metadata, nil
}
