package main

import (
	"fmt"
	
	"github.com/websim/j/internal/completion"
)

func main() {
	// Test different completion strategies
	
	// Strategy 1: Show paths containing the target name
	completions1 := []string{
		"@headless/machines/developer",
		"@headless/machines/our-computer", 
		"@modal-sandbox-manager",
	}
	
	fmt.Println("Strategy 1 - paths only:")
	result1 := completion.FuzzyMatchStrings("deploy", completions1)
	for i, r := range result1 {
		fmt.Printf("  %d: %s\n", i, r)
	}
	
	// Strategy 2: Show target@path format
	completions2 := []string{
		"deploy@headless/machines/developer",
		"deploy@headless/machines/our-computer", 
		"deploy@modal-sandbox-manager",
	}
	
	fmt.Println("\nStrategy 2 - target@path:")
	result2 := completion.FuzzyMatchStrings("deploy", completions2)
	for i, r := range result2 {
		fmt.Printf("  %d: %s\n", i, r)
	}
	
	// Strategy 3: Show "target (path)" format
	completions3 := []string{
		"deploy (@headless/machines/developer)",
		"deploy (@headless/machines/our-computer)", 
		"deploy (@modal-sandbox-manager)",
	}
	
	fmt.Println("\nStrategy 3 - target (path):")
	result3 := completion.FuzzyMatchStrings("deploy", completions3)
	for i, r := range result3 {
		fmt.Printf("  %d: %s\n", i, r)
	}
	
	// Test partial matches too
	fmt.Println("\nTesting 'dev' match on paths:")
	result4 := completion.FuzzyMatchStrings("dev", completions1)
	for i, r := range result4 {
		fmt.Printf("  %d: %s\n", i, r)
	}
}