package main

import (
	"fmt"
	"log"
	
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
	
	fmt.Printf("Found %d total targets:\n", len(targets))
	
	deployTargets := 0
	for _, target := range targets {
		if target.Name == "deploy" {
			deployTargets++
			fmt.Printf("  deploy at %s\n", target.JustfilePath)
		}
	}
	
	fmt.Printf("Found %d deploy targets\n", deployTargets)
}