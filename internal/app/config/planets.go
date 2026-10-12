package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// JWTConfig — настройки подписи и сроков жизни access-токена.
// Secret хранится в .env и никогда не попадает в swagger-описание.
type JWTConfig struct {
	Secret    string
	ExpiresIn time.Duration
	Issuer    string
}

// RedisConfig — подключение к хранилищу блеклиста токенов.
type RedisConfig struct {
	Host     string
	Port     int
	User     string
	Password string
}

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

	JWT   JWTConfig
	Redis RedisConfig
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

	viper.BindEnv("JWT.Secret", "JWT_SECRET")
	viper.BindEnv("JWT.ExpiresIn", "JWT_EXPIRES_IN")
	viper.BindEnv("JWT.Issuer", "JWT_ISSUER")

	viper.BindEnv("Redis.Host", "REDIS_HOST")
	viper.BindEnv("Redis.Port", "REDIS_PORT")
	viper.BindEnv("Redis.User", "REDIS_USER")
	viper.BindEnv("Redis.Password", "REDIS_PASSWORD")

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	log.Info("config parsed")

	return cfg, nil
}
