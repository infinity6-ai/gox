package cmdversionz

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/infinity6-ai/gox/versionz/version"
	"github.com/spf13/cobra"
)

func Prepare(ctx context.Context) *cobra.Command {
	var rootCmd = &cobra.Command{
		Version: version.Version(),
		Use:     "versionz",
		Short:   "versionz",
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the current version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.Version())
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "inc",
		Short: "Increment the patch version number",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			if err := IncrementVersion(); err != nil {
				return fmt.Errorf("error incrementing version: %w", err)
			}
			fmt.Fprintln(os.Stderr, "Version incremented successfully.")
			return nil
		},
	})

	return rootCmd
}

func Execute(ctx context.Context, rootCmd *cobra.Command) {
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func IncrementVersion() error {
	filePath := "version/version.go"
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("could not read file %s: %w", filePath, err)
	}

	re := regexp.MustCompile(`return "v(\d+)\.(\d+)\.(\d+)"`)
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

	err = os.WriteFile(filePath, []byte(newContent), 0o644)
	if err != nil {
		return fmt.Errorf("could not write to file %s: %w", filePath, err)
	}

	return nil
}
