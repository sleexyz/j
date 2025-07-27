package justfile

import (
	"bufio"
	"os"
	"strings"
)

// Target represents a justfile target
type Target struct {
	Name         string
	Description  string
	JustfilePath string
}

// GetTargets extracts targets from a justfile using `just --list`
func GetTargets(justfilePath string) ([]Target, error) {
	// Always fall back to file parsing for simplicity and reliability
	return GetTargetsFromFile(justfilePath)
}

// GetTargetsFromFile parses a justfile directly (fallback method)
func GetTargetsFromFile(justfilePath string) ([]Target, error) {
	file, err := os.Open(justfilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var targets []Target
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Look for target definitions (lines ending with :)
		if strings.HasSuffix(line, ":") && !strings.Contains(line, "=") {
			// Handle targets with parameters like "target param1 param2:"
			targetLine := strings.TrimSuffix(line, ":")
			targetLine = strings.TrimSpace(targetLine)
			
			// Extract just the target name (first word)
			parts := strings.Fields(targetLine)
			if len(parts) > 0 {
				targetName := parts[0]
				
				// Skip internal/private targets that start with _
				if strings.HasPrefix(targetName, "_") {
					continue
				}
				
				targets = append(targets, Target{
					Name:         targetName,
					JustfilePath: justfilePath,
				})
			}
		}
	}
	
	return targets, scanner.Err()
}

// GetTargetsFromAllJustfiles gets targets from all justfiles in the repository
func GetTargetsFromAllJustfiles(repoRoot string) ([]Target, error) {
	justfiles, err := FindAllJustfiles(repoRoot)
	if err != nil {
		return nil, err
	}
	
	var allTargets []Target
	
	for _, justfilePath := range justfiles {
		targets, err := GetTargets(justfilePath)
		if err != nil {
			// Fallback to file parsing
			targets, err = GetTargetsFromFile(justfilePath)
			if err != nil {
				continue // Skip this justfile if we can't parse it
			}
		}
		
		// Add all targets without deduplication - for autocomplete we want all instances
		allTargets = append(allTargets, targets...)
	}
	
	return allTargets, nil
}