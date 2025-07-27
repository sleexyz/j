package completion

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/websim/j/internal/justfile"
	"github.com/websim/j/internal/repo"
)

// CompleteWebsimPaths provides completion for @path arguments
func CompleteWebsimPaths(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	repoRoot, err := repo.FindRepoRoot()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	// Remove @ prefix from toComplete if present
	searchPath := toComplete
	if strings.HasPrefix(searchPath, "@") {
		searchPath = strings.TrimPrefix(searchPath, "@")
	}

	// Find directories with justfiles
	var allPaths []string

	// Walk the directory tree looking for justfiles
	err = filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors, continue walking
		}

		// Only look at directories
		if !info.IsDir() {
			return nil
		}

		// Skip hidden directories and common build/cache directories
		baseName := filepath.Base(path)
		if strings.HasPrefix(baseName, ".") || 
		   baseName == "node_modules" || 
		   baseName == "target" || 
		   baseName == "dist" || 
		   baseName == "build" ||
		   baseName == ".cache" ||
		   baseName == ".tmp" ||
		   baseName == "tmp" ||
		   baseName == "vendor" ||
		   baseName == ".venv" ||
		   baseName == "venv" ||
		   baseName == "__pycache__" ||
		   baseName == ".pytest_cache" ||
		   baseName == "coverage" ||
		   baseName == ".nyc_output" ||
		   baseName == "logs" {
			return filepath.SkipDir
		}

		// Check if this directory has a justfile
		justfilePath := filepath.Join(path, "justfile")
		if _, err := os.Stat(justfilePath); err == nil {
			// Convert absolute path to relative path from repo root
			relPath, err := filepath.Rel(repoRoot, path)
			if err != nil {
				return nil
			}

			// Skip the root directory itself
			if relPath == "." {
				return nil
			}

			websimPath := "@" + relPath
			allPaths = append(allPaths, websimPath)
		}

		return nil
	})

	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	// Use fuzzy matching instead of prefix matching
	completions := FuzzyMatchStrings(toComplete, allPaths)

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// CompletePathsWithTarget provides completion for @path arguments filtered by target
func CompletePathsWithTarget(cmd *cobra.Command, args []string, toComplete string, target string) ([]string, cobra.ShellCompDirective) {
	repoRoot, err := repo.FindRepoRoot()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	// Remove @ prefix from toComplete if present
	searchPath := toComplete
	if strings.HasPrefix(searchPath, "@") {
		searchPath = strings.TrimPrefix(searchPath, "@")
	}

	// Find all justfiles and check which ones contain the target
	var allPaths []string
	
	justfiles, err := justfile.FindAllJustfiles(repoRoot)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	
	for _, justfilePath := range justfiles {
		// Check if the justfile contains the target
		hasTarget, err := justfileContainsTarget(justfilePath, target)
		if err != nil || !hasTarget {
			continue
		}
		
		// Convert absolute path to @path format
		dir := filepath.Dir(justfilePath)
		relPath, err := filepath.Rel(repoRoot, dir)
		if err != nil {
			continue
		}
		
		if relPath == "." {
			// Skip root directory since we use @path syntax
			continue
		}
		
		websimPath := "@" + relPath
		allPaths = append(allPaths, websimPath)
	}

	// Use fuzzy matching instead of prefix matching
	completions := FuzzyMatchStrings(toComplete, allPaths)
	
	return completions, cobra.ShellCompDirectiveNoFileComp
}

// justfileContainsTarget checks if a justfile contains the specified target
func justfileContainsTarget(justfilePath, target string) (bool, error) {
	targets, err := justfile.GetTargets(justfilePath)
	if err != nil {
		// Fallback to file parsing
		targets, err = justfile.GetTargetsFromFile(justfilePath)
		if err != nil {
			return false, err
		}
	}

	for _, t := range targets {
		if t.Name == target {
			return true, nil
		}
	}

	return false, nil
}