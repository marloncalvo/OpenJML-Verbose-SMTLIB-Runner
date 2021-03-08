package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

var jmlPath string
var solverPath string
var methodTag string

// var solverType uint

func main() {
	flag.StringVar(&jmlPath, "verbose_file", "out.txt", "path to openjml verbose output")
	// flag.UintVar(&solverType, "solver_type", 0, "solver to use: \n\t0: z3\n\t1: cvc4")
	flag.StringVar(&solverPath, "solver_exe", "./cvc4.exe", "path to solver executable")
	flag.StringVar(&methodTag, "tag", "Test.smtlib_at()", "tag of current proof")
	flag.Parse()

	validPath(jmlPath, "verbose file could not be found")
	validPath(solverPath, "solver executable could not be found")

	if !isExecutable(solverPath) {
		exitError("solver executable must be executable")
	}

	smtlibInput, err := getSmtlibInput(jmlPath, methodTag)
	if err != nil {
		log.Fatal(err)
	}

	inputFile := temporaryInputFile(smtlibInput)
	_, err = inputFile.WriteString(smtlibInput)
	if err != nil {
		log.Fatal(err)
	}
	smtlibInputPath := inputFile.Name()
	err = inputFile.Close()
	if err != nil {
		log.Fatal(err)
	}

	run(smtlibInputPath, solverPath)
	defer os.Remove(inputFile.Name())
}

func run(inputPath string, solverPath string) {
	cmd := exec.Command(solverPath, "--lang=smt2.6", "--incremental", inputPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(out))
}

func temporaryInputFile(content string) *os.File {
	file, err := ioutil.TempFile("", "smtlib_input")
	if err != nil {
		exitError("unable to open temporary file")
	}

	return file
}

func getSmtlibInput(path string, tag string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	combinedTag := "SMT TRANSLATION OF " + tag

	contentString := string(content)
	contentString =
		contentString[strings.Index(contentString, combinedTag)+len(combinedTag):]

	contentString = strings.TrimSpace(contentString)[1:]
	contentString = contentString[:strings.Index(contentString, "(check-sat)")]
	contentString = contentString + "\n(check-sat)"

	return contentString, nil
}

func isExecutable(path string) bool {
	_, err := exec.LookPath(path)
	return err == nil
}

func validPath(path string, errorMessage string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}

	exitError(errorMessage)

	return false
}

func exitError(errorMessage string) {
	fmt.Println(errorMessage)
	os.Exit(3)
}
