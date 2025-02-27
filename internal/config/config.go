// internal/config/config.go
package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	OpenAIAPIKey      string `mapstructure:"OPENAI_API_KEY"`
	OpenAIModel       string `mapstructure:"OPENAI_MODEL"`
	WeaviateHost      string `mapstructure:"WEAVIATE_HOST"`
	WeaviateAPIKey    string `mapstructure:"WEAVIATE_API_KEY"`
	WeaviateIndexName string `mapstructure:"WEAVIATE_INDEX_NAME"`
	Debug             string `mapstructure:"DEBUG"`
}

// LoadConfig loads environment variables into the Config struct
func LoadConfig() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()

	// Load the config file
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: No .env file found (%v), loading from environment variables only.", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
