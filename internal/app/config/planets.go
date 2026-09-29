package config

import (
	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	ServiceHost string
	ServicePort int

	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	MinioUseSSL    bool
	MinioPublicURL string

	DefaultPhotoPath string
	DefaultVideoPath string
}

func NewConfig() (*Config, error) {
	_ = godotenv.Load()

	configName := "config"
	if name := os.Getenv("CONFIG_NAME"); name != "" {
		configName = name
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("toml")
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")
	viper.WatchConfig()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// Переменные окружения имеют приоритет над config.toml
	viper.BindEnv("MinioEndpoint", "MINIO_ENDPOINT")
	viper.BindEnv("MinioAccessKey", "MINIO_ACCESS_KEY")
	viper.BindEnv("MinioSecretKey", "MINIO_SECRET_KEY")
	viper.BindEnv("MinioBucket", "MINIO_BUCKET_NAME")
	viper.BindEnv("MinioUseSSL", "MINIO_USE_SSL")
	viper.BindEnv("MinioPublicURL", "MINIO_PUBLIC_URL")

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	log.Info("config parsed")

	return cfg, nil
}
