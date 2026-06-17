package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// findCmd searches for files matching a pattern under a root directory.
var findCmd = &cobra.Command{
	Use:   "find [root]",
	Short: "Find files matching a name pattern",
	Long: `Recursively find files whose names match a glob or regex pattern.

Example:
  devtool find . --name "*.go" --size-min 1024
  devtool find /tmp --regex "test_.*\.log"`,
	Args: cobra.MaximumNArgs(1),
	RunE: runFind,
}

// renameCmd batch-renames files matching a pattern.
var renameCmd = &cobra.Command{
	Use:   "rename [root]",
	Short: "Batch rename files using a regex substitution",
	Long: `Rename all files under root whose names match --pattern,
replacing the match with --replace. Use --dry-run to preview changes.

Example:
  devtool rename . --pattern "^test_" --replace "spec_" --dry-run`,
	Args: cobra.MaximumNArgs(1),
	RunE: runRename,
}

func init() {
	rootCmd.AddCommand(findCmd)
	rootCmd.AddCommand(renameCmd)

	findCmd.Flags().String("name", "*", "glob pattern for filename (e.g. '*.go')")
	findCmd.Flags().String("regex", "", "regex pattern for filename (overrides --name)")
	findCmd.Flags().Int64("size-min", 0, "minimum file size in bytes")
	findCmd.Flags().Int64("size-max", 0, "maximum file size in bytes (0 = no limit)")

	renameCmd.Flags().String("pattern", "", "regex pattern to match in filename (required)")
	renameCmd.Flags().String("replace", "", "replacement string (supports $1 capture groups)")
	renameCmd.Flags().Bool("dry-run", true, "preview renames without making changes")
	renameCmd.MarkFlagRequired("pattern")
	renameCmd.MarkFlagRequired("replace")
}

func runFind(cmd *cobra.Command, args []string) error {
	root := "."
	if len(args) > 0 {
		root = args[0]
	}

	nameGlob, _ := cmd.Flags().GetString("name")
	regexStr, _ := cmd.Flags().GetString("regex")
	sizeMin, _ := cmd.Flags().GetInt64("size-min")
	sizeMax, _ := cmd.Flags().GetInt64("size-max")

	var re *regexp.Regexp
	if regexStr != "" {
		var err error
		re, err = regexp.Compile(regexStr)
		if err != nil {
			return fmt.Errorf("invalid regex: %w", err)
		}
	}

	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		name := info.Name()

		// Name matching
		if re != nil {
			if !re.MatchString(name) {
				return nil
			}
		} else {
			matched, _ := filepath.Match(nameGlob, name)
			if !matched {
				return nil
			}
		}

		// Size filtering
		// TODO: implement size-min / size-max filtering
		_ = sizeMin
		_ = sizeMax

		fmt.Println(path)
		return nil
	})
}

func runRename(cmd *cobra.Command, args []string) error {
	root := "."
	if len(args) > 0 {
		root = args[0]
	}

	patternStr, _ := cmd.Flags().GetString("pattern")
	replace, _ := cmd.Flags().GetString("replace")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	re, err := regexp.Compile(patternStr)
	if err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}

	if dryRun {
		fmt.Println("[dry-run] no files will be changed")
	}

	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		oldName := info.Name()
		if !re.MatchString(oldName) {
			return nil
		}
		newName := re.ReplaceAllString(oldName, replace)
		if newName == oldName {
			return nil
		}

		newPath := strings.Replace(path, oldName, newName, 1)
		fmt.Printf("  %s -> %s\n", path, newPath)

		if !dryRun {
			// TODO: implement os.Rename(path, newPath) with conflict detection
			return fmt.Errorf("non-dry-run rename not yet implemented")
		}
		return nil
	})
}
