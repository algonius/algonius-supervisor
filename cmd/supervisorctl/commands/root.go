package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/algonius/algonius-supervisor/internal/cli/config"
)

var (
	cfgFile    string
	viperInstance *viper.Viper
	configManager config.IConfigManager
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "supervisorctl",
	Short: "Control and monitor supervised processes",
	Long: `supervisorctl is a command-line tool for controlling and monitoring
algonius supervisor daemon processes. It provides a traditional supervisor-like
interface for agent lifecycle management.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

// GetConfigManager returns the configuration manager instance
func GetConfigManager() config.IConfigManager {
	return configManager
}

// GetViperInstance returns the viper instance
func GetViperInstance() *viper.Viper {
	return viperInstance
}

func init() {
	// Initialize viper instance
	viperInstance = viper.New()

	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/supervisorctl/supervisorctl.yaml)")
	rootCmd.PersistentFlags().String("server-url", "", "supervisord server URL")
	rootCmd.PersistentFlags().String("token", "", "authentication token")
	rootCmd.PersistentFlags().String("format", "table", "output format (table, json, yaml)")
	rootCmd.PersistentFlags().Bool("no-colors", false, "disable colored output")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")

	// Bind flags to viper instance (not global viper)
	viperInstance.BindPFlag("server.url", rootCmd.PersistentFlags().Lookup("server-url"))
	viperInstance.BindPFlag("auth.token", rootCmd.PersistentFlags().Lookup("token"))
	viperInstance.BindPFlag("display.format", rootCmd.PersistentFlags().Lookup("format"))
	viperInstance.BindPFlag("display.colors", rootCmd.PersistentFlags().Lookup("no-colors"))
	viperInstance.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viperInstance.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".supervisorctl" (without extension).
		viperInstance.AddConfigPath(home + "/.config/supervisorctl")
		viperInstance.AddConfigPath(".")
		viperInstance.SetConfigType("yaml")
		viperInstance.SetConfigName("supervisorctl")
	}

	viperInstance.AutomaticEnv() // read in environment variables that match

	// Initialize config manager with injected viper instance
	configManager = config.NewConfigManager(viperInstance)

	// Load configuration
	if _, err := configManager.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load config: %v\n", err)
	}
}