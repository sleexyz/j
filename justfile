# Check for unstaged changes in package root (git added files are ok)
_check-clean:
	#!/usr/bin/env bash
	if git status --porcelain . | grep -qE "^.M|^\?\?"; then
		echo "Error: There are unstaged changes. Please git add them first."
		git status --porcelain .
		exit 1
	fi

# Install j globally via nix profile
install: _check-clean
	nix profile install .
	echo "j installed! Start a new zsh session for new completions." 

# Enter development shell
dev:
	nix develop

# Remove j from global profile
uninstall:
	nix profile remove "j"

# Reinstall j globally (uninstall + install)
reinstall: _check-clean uninstall install

# Update version and reinstall with latest changes
upgrade: _update-and-install

# Update version and reinstall (full development workflow)
_update-and-install: _update-version uninstall install

# Build the project
build:
	go build -o j ./cmd

# Update version and prepare for nix build (bump patch version, sync vendor, stage changes)
_update-version:
	#!/usr/bin/env bash
	set -euo pipefail
	
	# Get current version from j.nix
	current_version=$(grep 'version = ' j.nix | head -1 | sed 's/.*"\(.*\)".*/\1/')
	echo "Current version: $current_version"
	
	# Bump patch version
	IFS='.' read -ra VERSION_PARTS <<< "$current_version"
	major=${VERSION_PARTS[0]}
	minor=${VERSION_PARTS[1]}
	patch=${VERSION_PARTS[2]}
	new_patch=$((patch + 1))
	new_version="$major.$minor.$new_patch"
	echo "New version: $new_version"
	
	# Update version in j.nix (both places)
	sed -i.bak "s/version = \"$current_version\"/version = \"$new_version\"/g" j.nix
	sed -i.bak "s/-X main.version=$current_version/-X main.version=$new_version/g" j.nix
	rm j.nix.bak
	
	# Sync dependencies
	echo "Syncing Go dependencies..."
	go mod tidy
	go mod vendor
	
	# Stage changes for nix build
	echo "Staging changes..."
	git add .
	
	echo "Version updated to $new_version and changes staged. Ready for nix build/install."

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -f j result
