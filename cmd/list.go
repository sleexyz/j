package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/websim/j/internal/completion"
	"github.com/websim/j/internal/justfile"
	"github.com/websim/j/internal/repo"
)

var (
	outputFormat string
	recursive    bool
)

type TargetInfo struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Directory   string `json:"directory"`
	JustfilePath string `json:"justfile_path"`
}

var listCmd = &cobra.Command{
	Use:   "list [path]",
	Short: "List available justfile targets",
	Long: `List all available justfile targets in the repository or a specific directory.

Without arguments, lists targets from the current directory or repository root.
With a @path argument, lists targets from that specific directory.`,
	Example: `  j list                          # List targets in current directory or repo root
  j list @pages                  # List targets in pages directory
  j list --format json           # Output as JSON
  j list --recursive             # List targets from all justfiles in repo
  j -l                           # Short flag for list (just compatibility)`,
	Args: cobra.MaximumNArgs(1),
	RunE: listTargets,
}

func init() {
	listCmd.Flags().StringVarP(&outputFormat, "format", "f", "table", "output format (table, json)")
	listCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "include subdirectories")
	
	// Set up completion for path argument
	listCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completion.CompleteWebsimPaths(cmd, args, toComplete)
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

func listTargets(cmd *cobra.Command, args []string) error {
	repoRoot, err := repo.FindRepoRoot()
	if err != nil {
		return fmt.Errorf("failed to find repository root: %w", err)
	}

	var targets []TargetInfo

	if len(args) == 1 {
		// List targets from specific path
		websimPath := args[0]
		if !strings.HasPrefix(websimPath, "@") {
			return fmt.Errorf("path must start with @, got: %s", websimPath)
		}

		resolvedPath, err := repo.ResolveWebsimPath(websimPath, repoRoot)
		if err != nil {
			return fmt.Errorf("failed to resolve path %s: %w", websimPath, err)
		}

		targets, err = getTargetsFromDirectory(resolvedPath)
		if err != nil {
			return err
		}
	} else if recursive {
		// List targets from all justfiles in repo
		targets, err = getAllTargetsRecursive(repoRoot)
		if err != nil {
			return err
		}
	} else {
		// List targets from current directory or repo root
		justfilePath, err := justfile.FindBestJustfile(repoRoot)
		if err != nil {
			return err
		}

		dir := filepath.Dir(justfilePath)
		targets, err = getTargetsFromDirectory(dir)
		if err != nil {
			return err
		}
	}

	return outputTargets(targets)
}

func getTargetsFromDirectory(dir string) ([]TargetInfo, error) {
	justfilePath, err := justfile.FindJustfile(dir)
	if err != nil {
		return nil, err
	}

	targets, err := justfile.GetTargets(justfilePath)
	if err != nil {
		targets, err = justfile.GetTargetsFromFile(justfilePath)
		if err != nil {
			return nil, err
		}
	}

	var targetInfos []TargetInfo
	for _, target := range targets {
		targetInfos = append(targetInfos, TargetInfo{
			Name:        target.Name,
			Description: target.Description,
			Directory:   dir,
			JustfilePath: justfilePath,
		})
	}

	return targetInfos, nil
}

func getAllTargetsRecursive(repoRoot string) ([]TargetInfo, error) {
	var allTargets []TargetInfo

	err := filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors, continue walking
		}

		// Skip hidden directories and common build/cache directories
		if info.IsDir() {
			baseName := filepath.Base(path)
			if strings.HasPrefix(baseName, ".") ||
				baseName == "node_modules" ||
				baseName == "target" ||
				baseName == "dist" ||
				baseName == "build" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if this is a justfile
		if filepath.Base(path) == "justfile" {
			dir := filepath.Dir(path)
			targets, err := getTargetsFromDirectory(dir)
			if err != nil {
				// Skip directories with problematic justfiles
				return nil
			}
			allTargets = append(allTargets, targets...)
		}

		return nil
	})

	return allTargets, err
}

func outputTargets(targets []TargetInfo) error {
	switch outputFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(targets)
	case "table":
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "TARGET\tDESCRIPTION\tDIRECTORY")
		for _, target := range targets {
			fmt.Fprintf(w, "%s\t%s\t%s\n", target.Name, target.Description, target.Directory)
		}
		return w.Flush()
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}