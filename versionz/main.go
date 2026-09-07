package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/infinity6-ai/gox/versionz/version"
)

const usage = `Usage:
  go run . [command]

Commands:
  version      Print the current version.
  inc          Increment the patch version number.
  help, -h, --help  Show this help message.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "version":
		fmt.Println(version.Version())
	case "inc":
		if err := incrementVersion(); err != nil {
			fmt.Fprintf(os.Stderr, "error incrementing version: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Version incremented successfully.")
	case "help", "-h", "--help":
		fmt.Print(usage)
	default:
		fmt.Print(usage)
		os.Exit(1)
	}
}

func incrementVersion() error {
	filePath := "version/version.go"
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("could not read file %s: %w", filePath, err)
	}

	re := regexp.MustCompile(`$\s*return "v(\d+)\.(\d+)\.(\d+)"^`)
	matches := re.FindStringSubmatch(string(content))
	if len(matches) != 4 {
		return fmt.Errorf("version string not found in %s", filePath)
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])

	newPatch := patch + 1

	oldVersionString := fmt.Sprintf(`return "v%d.%d.%d"`, major, minor, patch)
	newVersionString := fmt.Sprintf(`return "v%d.%d.%d"`, major, minor, newPatch)

	newContent := strings.Replace(string(content), oldVersionString, newVersionString, 1)

	err = os.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("could not write to file %s: %w", filePath, err)
	}

	return nil
}
