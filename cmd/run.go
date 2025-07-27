package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/sleexyz/j/internal/completion"
	"github.com/sleexyz/j/internal/justfile"
	"github.com/sleexyz/j/internal/repo"
)

var (
	quiet     bool
	directory string
)

var runCmd = &cobra.Command{
	Use:   "run [target] [@path] [args...]",
	Short: "Run a justfile target",
	Long: `Run a justfile target in the current directory, repo root, or specified path.

The target is the name of the justfile target to execute.
The optional @path argument specifies a subdirectory within the repository.
Additional arguments are passed through to the justfile target.`,
	Example: `  j run build                      # Run build target
  j run dev @frontend             # Run dev target in frontend directory
  j run test @backend api         # Run test target in backend directory with 'api' argument
  j build                         # Shorthand (run is default command)
  j dev @frontend                 # Shorthand syntax`,
	Args: cobra.MinimumNArgs(1),
	RunE: runTarget,
}

func init() {
	runCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "suppress output")
	runCmd.Flags().StringVarP(&directory, "directory", "d", "", "run in specific directory")
	
	// Ignore unknown flags so they can be passed through to the inner just command
	runCmd.FParseErrWhitelist.UnknownFlags = true
	
	// Set up completion functions
	runCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			// Complete target names
			return completion.CompleteTargets(cmd, args, toComplete)
		} else if len(args) == 1 {
			// Complete @path filtered by the target in args[0]
			return completion.CompletePathsWithTarget(cmd, args, toComplete, args[0])
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

// findOriginalArgs reconstructs the original command arguments from os.Args
// to capture flags and arguments that Cobra may have consumed or reordered
func findOriginalArgs(parsedArgs []string) []string {
	// Start from os.Args[1:] (skip program name)
	rawArgs := os.Args[1:]
	
	// Find where the first non-flag argument (target) appears
	var targetIndex = -1
	for i, arg := range rawArgs {
		if !strings.HasPrefix(arg, "-") && arg != "run" {
			targetIndex = i
			break
		}
	}
	
	if targetIndex >= 0 {
		// Return everything starting from the target, but filter out known j flags
		result := rawArgs[targetIndex:]
		filtered := make([]string, 0, len(result))
		
		for i := 0; i < len(result); i++ {
			arg := result[i]
			// Skip known j flags and their values
			if arg == "--quiet" || arg == "-q" || 
			   arg == "--verbose" || arg == "-v" || strings.HasPrefix(arg, "--directory") || arg == "-d" {
				// Skip this flag and its value if it takes one
				if (arg == "--directory" || arg == "-d") && i+1 < len(result) && !strings.HasPrefix(result[i+1], "-") {
					i++ // Skip the directory value too
				}
				continue
			}
			filtered = append(filtered, arg)
		}
		return filtered
	}
	
	// Fallback to parsed args if we can't find the target in raw args
	return parsedArgs
}

func runTarget(cmd *cobra.Command, args []string) error {
	// Get original arguments from os.Args to capture all flags and arguments
	// Skip the program name and any global flags processed by Cobra
	originalArgs := findOriginalArgs(args)
	
	target := originalArgs[0]
	var repoPath string
	var extraArgs []string
	
	// Parse arguments to separate target, path, and extra args
	for i, arg := range originalArgs[1:] {
		if strings.HasPrefix(arg, "@") {
			repoPath = arg
			// Everything after the path becomes extra args
			if i+2 < len(originalArgs) {
				extraArgs = originalArgs[i+2:]
			}
			break
		} else if repoPath == "" {
			// If no @path found yet, treat remaining args as extra args
			extraArgs = originalArgs[1:]
			break
		}
	}
	
	// Find repository root
	repoRoot, err := repo.FindRepoRoot()
	if err != nil {
		return fmt.Errorf("failed to find repository root: %w", err)
	}
	
	var workingDir string
	var justfilePath string
	
	if repoPath != "" {
		// Handle @path syntax
		if !strings.HasPrefix(repoPath, "@") {
			return fmt.Errorf("path must start with @, got: %s", repoPath)
		}
		
		resolvedPath, err := repo.ResolveRepoPath(repoPath, repoRoot)
		if err != nil {
			return fmt.Errorf("failed to resolve path %s: %w", repoPath, err)
		}
		
		workingDir = resolvedPath
		justfilePath, err = justfile.FindJustfile(workingDir)
		if err != nil {
			return fmt.Errorf("no justfile found in %s", workingDir)
		}
	} else if directory != "" {
		// Handle -d/--directory flag
		workingDir = directory
		justfilePath, err = justfile.FindJustfile(workingDir)
		if err != nil {
			return fmt.Errorf("no justfile found in %s", workingDir)
		}
	} else {
		// Find the best justfile (current dir or repo root)
		justfilePath, err = justfile.FindBestJustfile(repoRoot)
		if err != nil {
			return err
		}
	}
	
	// Validate that the target exists
	if err := justfile.ValidateTarget(justfilePath, target); err != nil {
		return err
	}
	
	// Run the target with extra args
	return justfile.RunTarget(justfilePath, target, extraArgs, verbose && !quiet)
}