package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/sleexyz/j/internal/completion"
	"github.com/sleexyz/j/internal/justfile"
	"github.com/sleexyz/j/internal/repo"
)

var (
	verbose      bool
	quiet        bool
	directory    string
	outputFormat string
	recursive    bool
	version      = "1.0.3"
)

type TargetInfo struct {
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	Directory    string `json:"directory"`
	JustfilePath string `json:"justfile_path"`
}

var rootCmd = &cobra.Command{
	Use:   "j [target] [@path] [args...]",
	Short: "Modern justfile runner for monorepos",
	Long: `j is a powerful command-line tool for running justfile targets in monorepos.
It provides smart discovery of justfiles across your repository and supports
running targets from any directory using the @path syntax.`,
	Version: version,
	Args:    cobra.MinimumNArgs(0),
	Example: `  j build                           # Run build target in current directory or repo root
  j dev @frontend                  # Run dev target in frontend directory
  j test @backend api              # Run test target in backend directory with 'api' argument
  j list                           # List all available targets
  j list @service                  # List targets in service directory`,
}

var runCmd = &cobra.Command{
	Use:   "run [target] [@path] [args...]",
	Short: "Run a justfile target",
	Long: `Run a justfile target in the current directory, repo root, or specified path.

The target is the name of the justfile target to execute.
The optional @path argument specifies a subdirectory within the repository.
Additional arguments are passed through to the justfile target.`,
	Example: `  j run build                      # Run build target
  j run dev @frontend             # Run dev target in frontend directory
  j run test @backend api         # Run test target in backend directory with 'api' argument
  j build                         # Shorthand (run is default command)
  j dev @frontend                 # Shorthand syntax`,
	Args: cobra.MinimumNArgs(1),
	RunE: runTarget,
}

var listCmd = &cobra.Command{
	Use:   "list [path]",
	Short: "List available justfile targets",
	Long: `List all available justfile targets in the repository or a specific directory.

Without arguments, lists targets from the current directory or repository root.
With a @path argument, lists targets from that specific directory.`,
	Example: `  j list                          # List targets in current directory or repo root
  j list @frontend               # List targets in frontend directory
  j list --format json           # Output as JSON
  j list --recursive             # List targets from all justfiles in repo
  j -l                           # Short flag for list (just compatibility)`,
	Args: cobra.MaximumNArgs(1),
	RunE: listTargets,
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `To load completions:

Bash:

  $ source <(j completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ j completion bash > /etc/bash_completion.d/j
  # macOS:
  $ j completion bash > $(brew --prefix)/etc/bash_completion.d/j

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ j completion zsh > "${fpath[1]}/_j"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ j completion fish | source

  # To load completions for each session, execute once:
  $ j completion fish > ~/.config/fish/completions/j.fish

PowerShell:

  PS> j completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> j completion powershell > j.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Add run command flags to root command
	rootCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "suppress output")
	rootCmd.Flags().StringVarP(&directory, "directory", "d", "", "run in specific directory")

	// Add -l flag for just compatibility (acts like "j list")
	var listFlag bool
	rootCmd.Flags().BoolVarP(&listFlag, "list", "l", false, "list available justfile targets (just compatibility)")

	// Disable flag parsing after the first non-flag argument to pass flags through to just command
	rootCmd.DisableFlagParsing = false // We need this false to allow our own flags
	rootCmd.FParseErrWhitelist.UnknownFlags = true

	// Set run logic to handle -l flag
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if listFlag {
			return listTargets(cmd, args)
		}
		// If no arguments provided, show usage
		if len(args) == 0 {
			return cmd.Help()
		}
		// Otherwise, run the target
		return runTarget(cmd, args)
	}

	// Configure run command
	runCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "suppress output")
	runCmd.Flags().StringVarP(&directory, "directory", "d", "", "run in specific directory")
	runCmd.FParseErrWhitelist.UnknownFlags = true
	runCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completion.CompleteTargets(cmd, args, toComplete)
		} else if len(args) == 1 {
			return completion.CompletePathsWithTarget(cmd, args, toComplete, args[0])
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Configure list command
	listCmd.Flags().StringVarP(&outputFormat, "format", "f", "table", "output format (table, json)")
	listCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "include subdirectories")
	listCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completion.CompleteWebsimPaths(cmd, args, toComplete)
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Add subcommands but hide them from completion
	runCmd.Hidden = true
	listCmd.Hidden = true
	completionCmd.Hidden = true

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(completionCmd)

	// Set up completion to only show justfile targets
	rootCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// For the first argument, complete targets only (no subcommands)
		if len(args) == 0 {
			targets, directive := completion.CompleteTargets(cmd, args, toComplete)
			// Force no space and no file completion to prevent fallback to subcommands
			return targets, directive | cobra.ShellCompDirectiveNoSpace
		}
		// Otherwise use the run command's completion logic
		return runCmd.ValidArgsFunction(cmd, args, toComplete)
	}

	// Disable built-in help and completion subcommands from appearing in completions
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
}

// findOriginalArgs reconstructs the original command arguments from os.Args
// to capture flags and arguments that Cobra may have consumed or reordered
func findOriginalArgs(parsedArgs []string) []string {
	// Start from os.Args[1:] (skip program name)
	rawArgs := os.Args[1:]

	// Find where the first non-flag argument (target) appears
	var targetIndex = -1
	for i, arg := range rawArgs {
		if !strings.HasPrefix(arg, "-") && arg != "run" {
			targetIndex = i
			break
		}
	}

	if targetIndex >= 0 {
		// Return everything starting from the target, but filter out known j flags
		result := rawArgs[targetIndex:]
		filtered := make([]string, 0, len(result))

		for i := 0; i < len(result); i++ {
			arg := result[i]
			// Skip known j flags and their values
			if arg == "--quiet" || arg == "-q" ||
				arg == "--verbose" || arg == "-v" || strings.HasPrefix(arg, "--directory") || arg == "-d" {
				// Skip this flag and its value if it takes one
				if (arg == "--directory" || arg == "-d") && i+1 < len(result) && !strings.HasPrefix(result[i+1], "-") {
					i++ // Skip the directory value too
				}
				continue
			}
			filtered = append(filtered, arg)
		}
		return filtered
	}

	// Fallback to parsed args if we can't find the target in raw args
	return parsedArgs
}

func runTarget(cmd *cobra.Command, args []string) error {
	// Get original arguments from os.Args to capture all flags and arguments
	// Skip the program name and any global flags processed by Cobra
	originalArgs := findOriginalArgs(args)

	target := originalArgs[0]
	var websimPath string
	var extraArgs []string

	// Parse arguments to separate target, path, and extra args
	for i, arg := range originalArgs[1:] {
		if strings.HasPrefix(arg, "@") {
			websimPath = arg
			// Everything after the path becomes extra args
			if i+2 < len(originalArgs) {
				extraArgs = originalArgs[i+2:]
			}
			break
		} else if websimPath == "" {
			// If no @path found yet, treat remaining args as extra args
			extraArgs = originalArgs[1:]
			break
		}
	}

	// Find repository root
	repoRoot, err := repo.FindRepoRoot()
	if err != nil {
		return fmt.Errorf("failed to find repository root: %w", err)
	}

	var workingDir string
	var justfilePath string

	if websimPath != "" {
		// Handle @path syntax
		if !strings.HasPrefix(websimPath, "@") {
			return fmt.Errorf("path must start with @, got: %s", websimPath)
		}

		resolvedPath, err := repo.ResolveWebsimPath(websimPath, repoRoot)
		if err != nil {
			return fmt.Errorf("failed to resolve path %s: %w", websimPath, err)
		}

		workingDir = resolvedPath
		justfilePath, err = justfile.FindJustfile(workingDir)
		if err != nil {
			return fmt.Errorf("no justfile found in %s", workingDir)
		}
	} else if directory != "" {
		// Handle -d/--directory flag
		workingDir = directory
		justfilePath, err = justfile.FindJustfile(workingDir)
		if err != nil {
			return fmt.Errorf("no justfile found in %s", workingDir)
		}
	} else {
		// Find the best justfile (current dir or repo root)
		justfilePath, err = justfile.FindBestJustfile(repoRoot)
		if err != nil {
			return err
		}
	}

	// Validate that the target exists
	if err := justfile.ValidateTarget(justfilePath, target); err != nil {
		return err
	}

	// Run the target with extra args
	return justfile.RunTarget(justfilePath, target, extraArgs, verbose && !quiet)
}

func listTargets(cmd *cobra.Command, args []string) error {
	repoRoot, err := repo.FindRepoRoot()
	if err != nil {
		return fmt.Errorf("failed to find repository root: %w", err)
	}

	var targets []TargetInfo

	if len(args) == 1 {
		// List targets from specific path
		websimPath := args[0]
		if !strings.HasPrefix(websimPath, "@") {
			return fmt.Errorf("path must start with @, got: %s", websimPath)
		}

		resolvedPath, err := repo.ResolveWebsimPath(websimPath, repoRoot)
		if err != nil {
			return fmt.Errorf("failed to resolve path %s: %w", websimPath, err)
		}

		targets, err = getTargetsFromDirectory(resolvedPath)
		if err != nil {
			return err
		}
	} else if recursive {
		// List targets from all justfiles in repo
		targets, err = getAllTargetsRecursive(repoRoot)
		if err != nil {
			return err
		}
	} else {
		// List targets from current directory or repo root
		justfilePath, err := justfile.FindBestJustfile(repoRoot)
		if err != nil {
			return err
		}

		dir := filepath.Dir(justfilePath)
		targets, err = getTargetsFromDirectory(dir)
		if err != nil {
			return err
		}
	}

	return outputTargets(targets)
}

func getTargetsFromDirectory(dir string) ([]TargetInfo, error) {
	justfilePath, err := justfile.FindJustfile(dir)
	if err != nil {
		return nil, err
	}

	targets, err := justfile.GetTargets(justfilePath)
	if err != nil {
		targets, err = justfile.GetTargetsFromFile(justfilePath)
		if err != nil {
			return nil, err
		}
	}

	var targetInfos []TargetInfo
	for _, target := range targets {
		targetInfos = append(targetInfos, TargetInfo{
			Name:         target.Name,
			Description:  target.Description,
			Directory:    dir,
			JustfilePath: justfilePath,
		})
	}

	return targetInfos, nil
}

func getAllTargetsRecursive(repoRoot string) ([]TargetInfo, error) {
	var allTargets []TargetInfo

	err := filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors, continue walking
		}

		// Skip hidden directories and common build/cache directories
		if info.IsDir() {
			baseName := filepath.Base(path)
			if strings.HasPrefix(baseName, ".") ||
				baseName == "node_modules" ||
				baseName == "target" ||
				baseName == "dist" ||
				baseName == "build" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if this is a justfile
		if filepath.Base(path) == "justfile" {
			dir := filepath.Dir(path)
			targets, err := getTargetsFromDirectory(dir)
			if err != nil {
				// Skip directories with problematic justfiles
				return nil
			}
			allTargets = append(allTargets, targets...)
		}

		return nil
	})

	return allTargets, err
}

func outputTargets(targets []TargetInfo) error {
	switch outputFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(targets)
	case "table":
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "TARGET\tDESCRIPTION\tDIRECTORY")
		for _, target := range targets {
			fmt.Fprintf(w, "%s\t%s\t%s\n", target.Name, target.Description, target.Directory)
		}
		return w.Flush()
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}