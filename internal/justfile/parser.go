package justfile

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
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
	dir := filepath.Dir(justfilePath)
	
	cmd := exec.Command("just", "--list", "--unsorted")
	cmd.Dir = dir
	
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var targets []Target
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	
	// Skip the first line (header)
	scanner.Scan()
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		
		// Parse target line: "target_name    # description"
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}
		
		target := Target{
			Name:         parts[0],
			JustfilePath: justfilePath,
		}
		
		// Look for description after #
		if idx := strings.Index(line, "#"); idx != -1 {
			target.Description = strings.TrimSpace(line[idx+1:])
		}
		
		targets = append(targets, target)
	}
	
	return targets, scanner.Err()
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
			targetName := strings.TrimSuffix(line, ":")
			targetName = strings.TrimSpace(targetName)
			
			// Skip if it contains spaces (probably not a simple target)
			if !strings.Contains(targetName, " ") {
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