remember to `git add` before running any nix build / install commands.

Bump the patch number up every update so we can keep track of what version is installed at the current time.

DO NOT vendor dependencies in ./vendor

## Development & Testing

Use the `_sandbox` directory for throwaway test and debug code. This directory is gitignored and can be used for:
- Testing new features
- Writing debug scripts
- Temporary code experiments
- Prototype implementations

Example:
```bash
mkdir _sandbox
cd _sandbox
# Write test/debug code here
go run main.go
```
