package repo

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FindRepoRoot finds the root of the git repository
func FindRepoRoot() (string, error) {
	// Try git rev-parse first
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output)), nil
	}

	// Fallback to current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return cwd, nil
}

// IsGitRepo checks if the current directory is inside a git repository
func IsGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

// ResolveWebsimPath resolves @path/to/dir to actual filesystem path
func ResolveWebsimPath(websimPath, repoRoot string) (string, error) {
	// Remove @ prefix
	if strings.HasPrefix(websimPath, "@") {
		websimPath = strings.TrimPrefix(websimPath, "@")
	}

	fullPath := filepath.Join(repoRoot, websimPath)
	
	// Check if directory exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", err
	}

	return fullPath, nil
}