package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stuckinforloop/ticker/cmd/executor"
	"github.com/stuckinforloop/ticker/cmd/notifier"
	"github.com/stuckinforloop/ticker/cmd/scheduler"
	"github.com/stuckinforloop/ticker/cmd/server"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ticker",
	Short: "Ticker is an opensource distributed task scheduler",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(viper.GetString("DSN"))
		if err := cmd.Help(); err != nil {
			log.Fatal(err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ticker.yaml)")
	rootCmd.PersistentFlags().StringP("environment", "e", "dev", "execution environment")
	viper.BindPFlag("environment", rootCmd.PersistentFlags().Lookup("environment"))

	// add subcommands
	addSubCommands()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".ticker" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("toml")
		viper.SetConfigName(".ticker")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func addSubCommands() {
	rootCmd.AddCommand(executor.ExecutorCmd)
	rootCmd.AddCommand(scheduler.SchedulerCmd)
	rootCmd.AddCommand(notifier.NotifierCmd)
	rootCmd.AddCommand(server.ServerCmd)
}
