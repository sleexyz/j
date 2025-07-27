package justfile

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindJustfile searches for a justfile in the specified directory
func FindJustfile(dir string) (string, error) {
	justfilePath := filepath.Join(dir, "justfile")
	
	if _, err := os.Stat(justfilePath); err == nil {
		return justfilePath, nil
	}
	
	return "", fmt.Errorf("no justfile found in %s", dir)
}

// FindBestJustfile finds the most appropriate justfile for the current context
// It first checks the current directory, then the repo root
func FindBestJustfile(repoRoot string) (string, error) {
	// Try current directory first
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	if justfile, err := FindJustfile(cwd); err == nil {
		return justfile, nil
	}

	// Try repo root
	if justfile, err := FindJustfile(repoRoot); err == nil {
		return justfile, nil
	}

	return "", fmt.Errorf("no justfile found in current directory or repo root")
}

// FindAllJustfiles finds all justfiles in the repository, skipping common ignored directories
func FindAllJustfiles(repoRoot string) ([]string, error) {
	var justfiles []string
	
	// Common directories to skip for performance
	skipDirs := map[string]bool{
		".git":         true,
		"node_modules": true,
		".next":        true,
		"dist":         true,
		"build":        true,
		"target":       true,
		".cache":       true,
		".tmp":         true,
		"tmp":          true,
		"vendor":       true,
		".venv":        true,
		"venv":         true,
		"__pycache__":  true,
		".pytest_cache": true,
		"coverage":     true,
		".nyc_output":  true,
		"logs":         true,
		"*.log":        true,
	}
	
	err := filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking even if there are errors
		}
		
		// Skip common ignored directories
		if info.IsDir() && skipDirs[info.Name()] {
			return filepath.SkipDir
		}
		
		// Skip hidden directories (except .git which we already handle above)
		if info.IsDir() && len(info.Name()) > 0 && info.Name()[0] == '.' && info.Name() != ".git" {
			return filepath.SkipDir
		}
		
		if info.Name() == "justfile" && !info.IsDir() {
			justfiles = append(justfiles, path)
		}
		
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	return justfiles, nil
}