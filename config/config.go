package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config struct to hold configuration values from config.yaml
type Config struct {
	PairingPhone  string   `yaml:"pairing_phone"`
	BotName       string   `yaml:"bot_name"`
	Owners        []string `yaml:"owners"`
	CommandPrefix string   `yaml:"command_prefix"`
	LogLevel      string   `yaml:"log_level"`
}

// globalCfg stores the loaded configuration
var globalCfg Config

// Fungsi untuk mendapatkan path absolut folder tempat executable bot berada
func getExecutableDir() (string, error) {
	 execPath, err := os.Executable()
	 if err != nil {
		 return "", fmt.Errorf("gagal mendapatkan path executable: %w", err)
	 }
	 return filepath.Dir(execPath), nil
}

// LoadConfig reads the configuration from config.yaml into the globalCfg variable
func LoadConfig() error {
	 execDir, err := getExecutableDir()
	 if err != nil {
		 return fmt.Errorf("gagal mendapatkan direktori executable untuk config: %w", err)
	 }
	 configPath := filepath.Join(execDir, "config.yaml")

	 yamlFile, err := ioutil.ReadFile(configPath)
	 if err != nil {
		 return fmt.Errorf("gagal membaca file config %s: %w", configPath, err)
	 }

	 err = yaml.Unmarshal(yamlFile, &globalCfg)
	 if err != nil {
		 return fmt.Errorf("gagal unmarshal config YAML: %w", err)
	 }

	 // Set default log level if not specified
	 if globalCfg.LogLevel == "" {
		 globalCfg.LogLevel = "INFO"
	 }
	 // Set default command prefix if not specified
	 if globalCfg.CommandPrefix == "" {
		 globalCfg.CommandPrefix = "!"
	 }
	 // Set default bot name if not specified
	 if globalCfg.BotName == "" {
		 globalCfg.BotName = "MyBot"
	 }

	 return nil
}

// GetConfig returns the loaded configuration
func GetConfig() Config {
	 return globalCfg
}

