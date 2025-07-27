package main

import (
	"fmt"
	"log"
	"path/filepath"
	
	"github.com/websim/j/internal/justfile"
	"github.com/websim/j/internal/repo"
)

func main() {
	repoRoot, err := repo.FindRepoRoot()
	if err != nil {
		log.Fatal(err)
	}
	
	targets, err := justfile.GetTargetsFromAllJustfiles(repoRoot)
	if err != nil {
		log.Fatal(err)
	}
	
	// Group targets by name to detect duplicates
	targetMap := make(map[string][]justfile.Target)
	for _, target := range targets {
		targetMap[target.Name] = append(targetMap[target.Name], target)
	}

	var completions []string
	
	// For each unique target name
	for targetName, targetList := range targetMap {
		fmt.Printf("Target %s has %d instances:\n", targetName, len(targetList))
		if len(targetList) == 1 {
			// Single target - just show the target name
			completions = append(completions, targetName)
			fmt.Printf("  Added: %s\n", targetName)
		} else {
			// Multiple targets with same name - show paths where this target exists
			for _, target := range targetList {
				// Convert absolute path to @path format
				relPath, err := filepath.Rel(repoRoot, filepath.Dir(target.JustfilePath))
				if err != nil {
					fmt.Printf("  Error getting relative path: %v\n", err)
					continue
				}
				fmt.Printf("  RelPath: %s\n", relPath)
				if relPath == "." {
					// Skip root directory for now - could show as just target name
					fmt.Printf("  Skipping root directory\n")
					continue
				}
				websimPath := "@" + relPath
				completions = append(completions, websimPath)
				fmt.Printf("  Added: %s\n", websimPath)
			}
		}
	}
	
	fmt.Printf("\nFinal completions:\n")
	for _, comp := range completions {
		fmt.Printf("  %s\n", comp)
	}
}