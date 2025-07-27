remember to `git add` before running any nix build / install commands.

Bump the patch number up every update so we can keep track of what version is installed at the current time.

DO NOT vendor dependencies in ./vendor

Do not memorize codebase structure, since that documentation will easily skew; only memorize critical things e.g. how to build, test, etc.

## Build & Test

- `go build -o j ./cmd` - Build binary
- `./j` - Test locally built binary  
- `just build` - Build via justfile
- `just test` - Run tests
- `just install` - Install via nix profile


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
