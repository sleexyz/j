package justfile

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// RunTarget executes a just target in the specified directory with optional arguments
func RunTarget(justfilePath, target string, args []string, verbose bool) error {
	dir := filepath.Dir(justfilePath)
	
	cmdArgs := append([]string{target}, args...)
	
	if verbose {
		if len(args) > 0 {
			fmt.Printf("Running: cd %s && just %s %v\n", dir, target, args)
		} else {
			fmt.Printf("Running: cd %s && just %s\n", dir, target)
		}
	}
	
	cmd := exec.Command("just", cmdArgs...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	return cmd.Run()
}


// ValidateTarget checks if a target exists in the justfile
func ValidateTarget(justfilePath, target string) error {
	targets, err := GetTargets(justfilePath)
	if err != nil {
		// Fallback to file parsing if just --list fails
		targets, err = GetTargetsFromFile(justfilePath)
		if err != nil {
			return fmt.Errorf("failed to parse justfile: %w", err)
		}
	}
	
	for _, t := range targets {
		if t.Name == target {
			return nil
		}
	}
	
	var targetNames []string
	for _, t := range targets {
		targetNames = append(targetNames, t.Name)
	}
	
	return fmt.Errorf("target '%s' not found. Available targets: %v", target, targetNames)
}