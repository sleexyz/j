package completion

import (
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/websim/j/internal/justfile"
	"github.com/websim/j/internal/repo"
)

// CompleteTargets provides completion for justfile targets
func CompleteTargets(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// This function handles both first argument (target) and second argument (path) completion
	
	// If this is the second argument and first argument is a target, complete filtered paths
	if len(args) == 1 && !strings.HasPrefix(args[0], "@") {
		// First argument is a target name, complete paths that contain this target
		return CompletePathsWithTarget(cmd, args, toComplete, args[0])
	}

	// Check if there's a @path argument
	websimPath := ""
	for _, arg := range args {
		if strings.HasPrefix(arg, "@") {
			websimPath = arg
			break
		}
	}

	repoRoot, err := repo.FindRepoRoot()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var targets []justfile.Target

	if websimPath != "" {
		// If a specific websim path is provided, only get targets from that path
		resolvedPath, err := repo.ResolveWebsimPath(websimPath, repoRoot)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		justfilePath, err := justfile.FindJustfile(resolvedPath)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		targets, err = justfile.GetTargets(justfilePath)
		if err != nil {
			// Fallback to file parsing
			targets, err = justfile.GetTargetsFromFile(justfilePath)
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}
		}
	} else {
		// Get targets from all justfiles in the repository
		targets, err = justfile.GetTargetsFromAllJustfiles(repoRoot)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
	}

	// Group targets by name to detect duplicates
	targetMap := make(map[string][]justfile.Target)
	for _, target := range targets {
		targetMap[target.Name] = append(targetMap[target.Name], target)
	}

	var completions []string
	
	// For each unique target name
	for targetName, targetList := range targetMap {
		if len(targetList) == 1 {
			// Single target - just show the target name
			completions = append(completions, targetName)
		} else {
			// Multiple targets with same name - show "target (path)" format
			for _, target := range targetList {
				// Convert absolute path to @path format
				relPath, err := filepath.Rel(repoRoot, filepath.Dir(target.JustfilePath))
				if err != nil {
					continue
				}
				if relPath == "." {
					// Root directory - just show target name
					completions = append(completions, targetName)
					continue
				}
				websimPath := "@" + relPath
				// Show as "target (path)" so fuzzy matching can find both target name and path
				completion := targetName + " (" + websimPath + ")"
				completions = append(completions, completion)
			}
		}
	}

	// Use fuzzy matching instead of prefix matching
	filtered := FuzzyMatchStrings(toComplete, completions)

	return filtered, cobra.ShellCompDirectiveNoFileComp
}