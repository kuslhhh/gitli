package cmd

import (
	"fmt"
	"os"

	"github.com/kush/gitli/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	version = "dev"
)

var rootCmd = &cobra.Command{
	Use:   "gitli",
	Short: "Local-first developer memory system",
	Long:  `gitli continuously indexes Git repositories and provides fast search, timeline, and developer insights.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("gitli - Local-first developer memory system")
		fmt.Println("Run 'gitli --help' for available commands")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gitli.yaml)")
	rootCmd.AddCommand(versionCmd)
}

func initConfig() {
	if err := config.Load(cfgFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("gitli version %s\n", version)
	},
}