package main

import (
	"fmt"
	
	"github.com/websim/j/internal/completion"
)

func main() {
	completions := []string{
		"@headless/machines/developer",
		"@headless/machines/our-computer", 
		"@modal-sandbox-manager",
		"chrome",
	}
	
	fmt.Println("Testing fuzzy match for 'deploy':")
	result := completion.FuzzyMatchStrings("deploy", completions)
	for i, r := range result {
		fmt.Printf("  %d: %s\n", i, r)
	}
}