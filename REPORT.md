# Nix Profile Installation Issue Report

## Problem Summary

The `j reinstall @nix/j` command was failing due to a mismatch between the expected package name and the actual installed profile name.

## Root Cause Analysis

### The Issue
1. **Profile Name Mismatch**: The `just uninstall` target was trying to remove a package named `j`, but the package was actually installed in the Nix profile under the name `websim` (the flake's directory name).

2. **Conflicting Versions**: When `nix build` was run multiple times during development, it didn't directly cause the profile issue, but it highlighted that version 1.0.1 and 1.0.3 were conflicting in the installation.

### Why This Happened
- When installing with `nix profile install ".#j"` from the root flake, Nix sometimes uses the flake name (directory name) instead of the package attribute name for the profile entry
- The `just uninstall` target was hardcoded to remove `j` but the actual profile name was `websim`
- This created a situation where uninstall would fail, leaving old versions installed and causing conflicts

## What `nix build` Does
- `nix build` creates a `result` symlink pointing to the built package in the Nix store
- It does **NOT** install anything in the user profile
- It does **NOT** have side effects on profile installations
- The `result` symlink is harmless and can be safely ignored or removed

## Solution Implemented
1. **Fixed the justfile**: Changed `nix profile remove j` to `nix profile remove websim` to match the actual profile name
2. **Removed conflicting package**: Manually removed the old `websim` profile entry
3. **Reinstalled cleanly**: The package now installs correctly as `j` in the profile

## Prevention Strategies

### 1. Use Explicit Profile Names
Consider using explicit profile names when installing:
```bash
nix profile install ".#j" --profile-name j
```
This ensures consistent naming regardless of flake directory name.

### 2. Robust Uninstall Script
Instead of hardcoding the package name, make the uninstall more robust:
```bash
# Remove j from global profile (handle both possible names)
uninstall:
	nix profile remove j || nix profile remove websim || echo "No j package found to remove"
```

### 3. Separate Flake for Standalone Packages
**Recommendation**: Consider creating a separate `flake.nix` inside `/nix/j/` for the `j` package specifically. This would:
- Avoid naming conflicts with the parent flake
- Make the package truly standalone
- Ensure consistent profile naming
- Simplify the installation process

Example `/nix/j/flake.nix`:
```nix
{
  description = "j - Modern justfile runner for monorepos";
  
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };
  
  outputs = { self, nixpkgs }:
    let
      forAllSystems = nixpkgs.lib.genAttrs [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
    in {
      packages = forAllSystems (system: {
        default = nixpkgs.legacyPackages.${system}.callPackage ./j-go.nix { };
      });
    };
}
```

### 4. Add Profile Verification
Add a verification step to check what's actually installed:
```bash
# Verify current installation
verify:
	@echo "Currently installed j packages:"
	@nix profile list | grep -E "(j|websim)" || echo "No j package found"
```

### 5. Development Workflow Improvements
- Always run `git add .` before nix operations (already implemented via `_check-clean`)
- Use `nix build` for testing builds without affecting the profile
- Use `just reinstall` for actual profile updates
- Consider using `nix develop` for development instead of global installation

## Conclusion

The issue was caused by Nix profile naming inconsistencies, not by `nix build` side effects. The fix involved correcting the expected profile name in the justfile. For future robustness, consider implementing a separate flake for the `j` package or using more defensive uninstall scripts.