package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	GRPCAddress string
	LoggerLevel string
	DataBaseURL string
	JWTSecret   string
}

func LoadConfig() (*Config, error) {
	v := viper.New()

	// Дефолтные значения
	v.SetDefault("grpc_address", ":8080")
	v.SetDefault("logger_level", "INFO")

	// Переменные окружения
	v.AllowEmptyEnv(true)
	v.SetEnvPrefix("GOPHKEEPER")
	// Заменяем точки на подчеркивания для переменных окружения
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	config := &Config{
		GRPCAddress: v.GetString("grpc_address"),
		LoggerLevel: v.GetString("logger_level"),
		DataBaseURL: v.GetString("database_url"),
		JWTSecret:   v.GetString("jwt_secret"),
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// validate проверяет, что заданы обязательные секреты и параметры.
func (c *Config) validate() error {
	if c.DataBaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	return nil
}
