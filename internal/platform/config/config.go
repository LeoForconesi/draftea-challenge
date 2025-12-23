package config

import (
	"errors"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"
)

// Config contains all application configuration.
type Config struct {
	App     AppConfig     `mapstructure:"app"`
	DB      DBConfig      `mapstructure:"db"`
	Rabbit  RabbitConfig  `mapstructure:"rabbit"`
	Gateway GatewayConfig `mapstructure:"gateway"`
	Logger  LoggerConfig  `mapstructure:"logger"`
}

// AppConfig defines HTTP server settings.
type AppConfig struct {
	HTTPAddr        string        `mapstructure:"http_addr"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	Env             string        `mapstructure:"env"`
	RequestTimeout  time.Duration `mapstructure:"request_timeout"`
	APIKey          string        `mapstructure:"api_key"`
}

// DBConfig defines Postgres connection settings.
type DBConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`
}

// RabbitConfig defines RabbitMQ connection settings.
type RabbitConfig struct {
	URL string `mapstructure:"url"`
}

// GatewayConfig defines external gateway settings.
type GatewayConfig struct {
	URL     string        `mapstructure:"url"`
	Timeout time.Duration `mapstructure:"timeout"`
}

// LoggerConfig defines logging settings.
type LoggerConfig struct {
	Level       string `mapstructure:"level"`
	Development bool   `mapstructure:"development"`
}

// Load reads configuration from an optional YAML file and environment variables.
func Load() (Config, error) {
	cfg := Config{}

	v := viper.New()
	v.SetConfigType("yaml")
	v.SetDefault("app.http_addr", ":8080")
	v.SetDefault("app.shutdown_timeout", 10*time.Second)
	v.SetDefault("app.env", "local")
	v.SetDefault("app.request_timeout", 5*time.Second)
	v.SetDefault("app.api_key", "")
	v.SetDefault("db.host", "localhost")
	v.SetDefault("db.port", 5432)
	v.SetDefault("db.user", "user")
	v.SetDefault("db.password", "password")
	v.SetDefault("db.name", "draftea")
	v.SetDefault("db.sslmode", "disable")
	v.SetDefault("rabbit.url", "amqp://guest:guest@localhost:5672/")
	v.SetDefault("gateway.url", "http://localhost:8081")
	v.SetDefault("gateway.timeout", 5*time.Second)
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.development", true)

	if configFile := os.Getenv("CONFIG_FILE"); configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		v.SetConfigName("config")
		v.AddConfigPath(".")
	}

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return cfg, err
		}
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, err
	}

	envCfg := envConfig{}
	if err := envconfig.Process("", &envCfg); err != nil {
		return cfg, err
	}

	applyEnvOverrides(&cfg, envCfg)
	return cfg, nil
}

type envConfig struct {
	App struct {
		HTTPAddr        *string        `envconfig:"HTTP_ADDR"`
		ShutdownTimeout *time.Duration `envconfig:"HTTP_SHUTDOWN_TIMEOUT"`
		Env             *string        `envconfig:"APP_ENV"`
		RequestTimeout  *time.Duration `envconfig:"REQUEST_TIMEOUT"`
		APIKey          *string        `envconfig:"API_KEY"`
	}
	DB struct {
		Host     *string `envconfig:"DB_HOST"`
		Port     *int    `envconfig:"DB_PORT"`
		User     *string `envconfig:"DB_USER"`
		Password *string `envconfig:"DB_PASSWORD"`
		Name     *string `envconfig:"DB_NAME"`
		SSLMode  *string `envconfig:"DB_SSLMODE"`
	}
	Rabbit struct {
		URL *string `envconfig:"RABBITMQ_URL"`
	}
	Gateway struct {
		URL     *string        `envconfig:"GATEWAY_URL"`
		Timeout *time.Duration `envconfig:"GATEWAY_TIMEOUT"`
	}
	Logger struct {
		Level       *string `envconfig:"LOG_LEVEL"`
		Development *bool   `envconfig:"LOG_DEVELOPMENT"`
	}
}

func applyEnvOverrides(cfg *Config, env envConfig) {
	if env.App.HTTPAddr != nil {
		cfg.App.HTTPAddr = *env.App.HTTPAddr
	}
	if env.App.ShutdownTimeout != nil {
		cfg.App.ShutdownTimeout = *env.App.ShutdownTimeout
	}
	if env.App.Env != nil {
		cfg.App.Env = *env.App.Env
	}
	if env.App.RequestTimeout != nil {
		cfg.App.RequestTimeout = *env.App.RequestTimeout
	}
	if env.App.APIKey != nil {
		cfg.App.APIKey = *env.App.APIKey
	}

	if env.DB.Host != nil {
		cfg.DB.Host = *env.DB.Host
	}
	if env.DB.Port != nil {
		cfg.DB.Port = *env.DB.Port
	}
	if env.DB.User != nil {
		cfg.DB.User = *env.DB.User
	}
	if env.DB.Password != nil {
		cfg.DB.Password = *env.DB.Password
	}
	if env.DB.Name != nil {
		cfg.DB.Name = *env.DB.Name
	}
	if env.DB.SSLMode != nil {
		cfg.DB.SSLMode = *env.DB.SSLMode
	}

	if env.Rabbit.URL != nil {
		cfg.Rabbit.URL = *env.Rabbit.URL
	}

	if env.Gateway.URL != nil {
		cfg.Gateway.URL = *env.Gateway.URL
	}
	if env.Gateway.Timeout != nil {
		cfg.Gateway.Timeout = *env.Gateway.Timeout
	}

	if env.Logger.Level != nil {
		cfg.Logger.Level = *env.Logger.Level
	}
	if env.Logger.Development != nil {
		cfg.Logger.Development = *env.Logger.Development
	}
}
