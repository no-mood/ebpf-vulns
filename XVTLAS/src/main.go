/*
**XDP Verifier Test Launch Automation Suite**

This tool manages batch tests for XDP or EBPF programs, the program flows as follows:
- Data input : the test source directory containing the XDP/eBPF tests to load, with
	the corresponding metadata files, metadata files sintax is explained in the readme
- result of the tests are displayed in the CLI result and can be exported to LaTex or Markdown
*/



package main

import (
	"flag"
	"fmt"
	"os"
)


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
}

