package utils

import (
	"github.com/spf13/viper"
	"sync"
)

var AppVersion = "wip"

type AppConfig struct {
	Port int

	S3Region     string
	S3BucketName string

	// Configuration S3 / MinIO when hors AWS
	S3UseSSL    bool
	S3Endpoint  string
	S3AccessKey string
	S3SecretKey string
}

var (
	instance *AppConfig
	once     sync.Once
)

func GetConfig() *AppConfig {
	once.Do(func() {
		instance = loadConfig()
	})
	return instance
}

func loadConfig() *AppConfig {
	viper.AutomaticEnv()

	cfg := AppConfig{}

	viper.SetDefault("PORT", 9090)
	cfg.Port = viper.GetInt("PORT")

	viper.SetDefault("AWS_REGION", "eu-west-3")
	viper.SetDefault("BUCKET", "audio")
	cfg.S3Region = viper.GetString("AWS_REGION")
	cfg.S3BucketName = viper.GetString("S3_BUCKET_NAME")

	viper.SetDefault("AWS_S3_ENDPOINT", "")
	viper.SetDefault("AWS_S3_USE_SSL", false)
	viper.SetDefault("AWS_S3_ACCESS_KEY", "")
	viper.SetDefault("AWS_S3_SECRET_KEY", "")
	cfg.S3UseSSL = viper.GetBool("AWS_S3_USE_SSL")
	cfg.S3Endpoint = viper.GetString("AWS_S3_ENDPOINT")
	cfg.S3AccessKey = viper.GetString("AWS_S3_ACCESS_KEY")
	cfg.S3SecretKey = viper.GetString("AWS_S3_SECRET_KEY")

	return &cfg
}
