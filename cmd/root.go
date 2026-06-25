package cmd

import (
	"fmt"
	"os"
	"sync"

	"github.com/kush/gitli/internal/config"
	"github.com/kush/gitli/internal/database"
	"github.com/spf13/cobra"
)

var (
	cfgFile    string
	version    = "dev"
	db         *database.DB
	dbOnce     sync.Once
	dbInitErr  error
)

var rootCmd = &cobra.Command{
	Use:   "gitli",
	Short: "Local-first developer memory system",
	Long:  `gitli continuously indexes Git repositories and provides fast search, timeline, and developer insights.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip DB init for the version command
		if cmd.Name() == "version" {
			return nil
		}
		return initDB()
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if db != nil {
			return db.Close()
		}
		return nil
	},
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

func initDB() error {
	dbOnce.Do(func() {
		cfg := config.Get()
		db, dbInitErr = database.New(cfg.Database.Path)
	})
	return dbInitErr
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("gitli version %s\n", version)
	},
}
