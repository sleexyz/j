package main

import (
	"fmt"
	"log"
	"path/filepath"
	
	"github.com/websim/j/internal/completion"
	"github.com/websim/j/internal/justfile"
	"github.com/websim/j/internal/repo"
)

func main() {
	repoRoot, err := repo.FindRepoRoot()
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Repo root: %s\n", repoRoot)
	
	targets, err := justfile.GetTargetsFromAllJustfiles(repoRoot)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Found %d total targets\n", len(targets))
	
	// Group targets by name to detect duplicates
	targetMap := make(map[string][]justfile.Target)
	for _, target := range targets {
		targetMap[target.Name] = append(targetMap[target.Name], target)
	}

	var completions []string
	
	// For each unique target name
	for targetName, targetList := range targetMap {
		fmt.Printf("Processing target %s with %d instances\n", targetName, len(targetList))
		if len(targetList) == 1 {
			// Single target - just show the target name
			completions = append(completions, targetName)
			fmt.Printf("  Added single: %s\n", targetName)
		} else {
			// Multiple targets with same name - show "target (path)" format
			for _, target := range targetList {
				// Convert absolute path to @path format
				relPath, err := filepath.Rel(repoRoot, filepath.Dir(target.JustfilePath))
				if err != nil {
					fmt.Printf("  Error: %v\n", err)
					continue
				}
				if relPath == "." {
					// Root directory - just show target name
					completions = append(completions, targetName)
					fmt.Printf("  Added root: %s\n", targetName)
					continue
				}
				websimPath := "@" + relPath
				// Show as "target (path)" so fuzzy matching can find both target name and path
				completion := targetName + " (" + websimPath + ")"
				completions = append(completions, completion)
				fmt.Printf("  Added path: %s\n", completion)
			}
		}
	}
	
	fmt.Printf("\nAll completions before fuzzy match:\n")
	for i, comp := range completions {
		fmt.Printf("  %d: %s\n", i, comp)
	}
	
	// Use fuzzy matching instead of prefix matching
	toComplete := "deploy"
	filtered := completion.FuzzyMatchStrings(toComplete, completions)
	
	fmt.Printf("\nAfter fuzzy matching '%s':\n", toComplete)
	for i, comp := range filtered {
		fmt.Printf("  %d: %s\n", i, comp)
	}
}