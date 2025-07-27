package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/sleexyz/j/internal/completion"
)

var (
	verbose bool
	version = "1.0.3"
)

var rootCmd = &cobra.Command{
	Use:   "j [target] [@path] [args...]",
	Short: "Modern justfile runner for monorepos",
	Long: `j is a powerful command-line tool for running justfile targets in monorepos.
It provides smart discovery of justfiles across your repository and supports
running targets from any directory using the @path syntax.`,
	Version: version,
	Args: cobra.MinimumNArgs(0),
	Example: `  j build                           # Run build target in current directory or repo root
  j dev @frontend                  # Run dev target in frontend directory
  j test @backend api              # Run test target in backend directory with 'api' argument
  j list                           # List all available targets
  j list @service                  # List targets in service directory`,
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
	
	// Add subcommands but hide them from completion
	runCmd.Hidden = true
	listCmd.Hidden = true
	completionCmd.Hidden = true
	
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(completionCmd)
	
	// Make run the default command when no subcommand is specified
	// This will be overridden in init() to handle the -l flag
	
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

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}