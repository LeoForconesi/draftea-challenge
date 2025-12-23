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
	URL                   string        `mapstructure:"url"`
	Exchange              string        `mapstructure:"exchange"`
	MetricsQueue          string        `mapstructure:"metrics_queue"`
	AuditQueue            string        `mapstructure:"audit_queue"`
	PublishConfirmTimeout time.Duration `mapstructure:"publish_confirm_timeout"`
	RelayBatchSize        int           `mapstructure:"relay_batch_size"`
	RelayMaxInFlight      int           `mapstructure:"relay_max_in_flight"`
	RelayMaxRetries       int           `mapstructure:"relay_max_retries"`
	RelayInitialBackoff   time.Duration `mapstructure:"relay_initial_backoff"`
	RelayMaxBackoff       time.Duration `mapstructure:"relay_max_backoff"`
}

// GatewayConfig defines external gateway settings.
type GatewayConfig struct {
	URL                    string        `mapstructure:"url"`
	Timeout                time.Duration `mapstructure:"timeout"`
	MaxRetries             int           `mapstructure:"max_retries"`
	RetryInitialBackoff    time.Duration `mapstructure:"retry_initial_backoff"`
	RetryMaxBackoff        time.Duration `mapstructure:"retry_max_backoff"`
	CircuitBreakerFailures int           `mapstructure:"circuit_breaker_failures"`
	CircuitBreakerCooldown time.Duration `mapstructure:"circuit_breaker_cooldown"`
	MaxInFlight            int           `mapstructure:"max_in_flight"`
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
	v.SetDefault("rabbit.exchange", "payments.events")
	v.SetDefault("rabbit.metrics_queue", "metrics.queue")
	v.SetDefault("rabbit.audit_queue", "audit.queue")
	v.SetDefault("rabbit.publish_confirm_timeout", 2*time.Second)
	v.SetDefault("rabbit.relay_batch_size", 100)
	v.SetDefault("rabbit.relay_max_in_flight", 10)
	v.SetDefault("rabbit.relay_max_retries", 3)
	v.SetDefault("rabbit.relay_initial_backoff", 200*time.Millisecond)
	v.SetDefault("rabbit.relay_max_backoff", 2*time.Second)
	v.SetDefault("gateway.url", "http://localhost:8081")
	v.SetDefault("gateway.timeout", 5*time.Second)
	v.SetDefault("gateway.max_retries", 2)
	v.SetDefault("gateway.retry_initial_backoff", 200*time.Millisecond)
	v.SetDefault("gateway.retry_max_backoff", 2*time.Second)
	v.SetDefault("gateway.circuit_breaker_failures", 5)
	v.SetDefault("gateway.circuit_breaker_cooldown", 10*time.Second)
	v.SetDefault("gateway.max_in_flight", 20)
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
		URL                   *string        `envconfig:"RABBITMQ_URL"`
		Exchange              *string        `envconfig:"RABBITMQ_EXCHANGE"`
		MetricsQueue          *string        `envconfig:"RABBITMQ_METRICS_QUEUE"`
		AuditQueue            *string        `envconfig:"RABBITMQ_AUDIT_QUEUE"`
		PublishConfirmTimeout *time.Duration `envconfig:"RABBITMQ_PUBLISH_CONFIRM_TIMEOUT"`
		RelayBatchSize        *int           `envconfig:"RABBITMQ_RELAY_BATCH_SIZE"`
		RelayMaxInFlight      *int           `envconfig:"RABBITMQ_RELAY_MAX_IN_FLIGHT"`
		RelayMaxRetries       *int           `envconfig:"RABBITMQ_RELAY_MAX_RETRIES"`
		RelayInitialBackoff   *time.Duration `envconfig:"RABBITMQ_RELAY_INITIAL_BACKOFF"`
		RelayMaxBackoff       *time.Duration `envconfig:"RABBITMQ_RELAY_MAX_BACKOFF"`
	}
	Gateway struct {
		URL                    *string        `envconfig:"GATEWAY_URL"`
		Timeout                *time.Duration `envconfig:"GATEWAY_TIMEOUT"`
		MaxRetries             *int           `envconfig:"GATEWAY_MAX_RETRIES"`
		RetryInitialBackoff    *time.Duration `envconfig:"GATEWAY_RETRY_INITIAL_BACKOFF"`
		RetryMaxBackoff        *time.Duration `envconfig:"GATEWAY_RETRY_MAX_BACKOFF"`
		CircuitBreakerFailures *int           `envconfig:"GATEWAY_CIRCUIT_BREAKER_FAILURES"`
		CircuitBreakerCooldown *time.Duration `envconfig:"GATEWAY_CIRCUIT_BREAKER_COOLDOWN"`
		MaxInFlight            *int           `envconfig:"GATEWAY_MAX_IN_FLIGHT"`
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
	if env.Rabbit.Exchange != nil {
		cfg.Rabbit.Exchange = *env.Rabbit.Exchange
	}
	if env.Rabbit.MetricsQueue != nil {
		cfg.Rabbit.MetricsQueue = *env.Rabbit.MetricsQueue
	}
	if env.Rabbit.AuditQueue != nil {
		cfg.Rabbit.AuditQueue = *env.Rabbit.AuditQueue
	}
	if env.Rabbit.PublishConfirmTimeout != nil {
		cfg.Rabbit.PublishConfirmTimeout = *env.Rabbit.PublishConfirmTimeout
	}
	if env.Rabbit.RelayBatchSize != nil {
		cfg.Rabbit.RelayBatchSize = *env.Rabbit.RelayBatchSize
	}
	if env.Rabbit.RelayMaxInFlight != nil {
		cfg.Rabbit.RelayMaxInFlight = *env.Rabbit.RelayMaxInFlight
	}
	if env.Rabbit.RelayMaxRetries != nil {
		cfg.Rabbit.RelayMaxRetries = *env.Rabbit.RelayMaxRetries
	}
	if env.Rabbit.RelayInitialBackoff != nil {
		cfg.Rabbit.RelayInitialBackoff = *env.Rabbit.RelayInitialBackoff
	}
	if env.Rabbit.RelayMaxBackoff != nil {
		cfg.Rabbit.RelayMaxBackoff = *env.Rabbit.RelayMaxBackoff
	}

	if env.Gateway.URL != nil {
		cfg.Gateway.URL = *env.Gateway.URL
	}
	if env.Gateway.Timeout != nil {
		cfg.Gateway.Timeout = *env.Gateway.Timeout
	}
	if env.Gateway.MaxRetries != nil {
		cfg.Gateway.MaxRetries = *env.Gateway.MaxRetries
	}
	if env.Gateway.RetryInitialBackoff != nil {
		cfg.Gateway.RetryInitialBackoff = *env.Gateway.RetryInitialBackoff
	}
	if env.Gateway.RetryMaxBackoff != nil {
		cfg.Gateway.RetryMaxBackoff = *env.Gateway.RetryMaxBackoff
	}
	if env.Gateway.CircuitBreakerFailures != nil {
		cfg.Gateway.CircuitBreakerFailures = *env.Gateway.CircuitBreakerFailures
	}
	if env.Gateway.CircuitBreakerCooldown != nil {
		cfg.Gateway.CircuitBreakerCooldown = *env.Gateway.CircuitBreakerCooldown
	}
	if env.Gateway.MaxInFlight != nil {
		cfg.Gateway.MaxInFlight = *env.Gateway.MaxInFlight
	}

	if env.Logger.Level != nil {
		cfg.Logger.Level = *env.Logger.Level
	}
	if env.Logger.Development != nil {
		cfg.Logger.Development = *env.Logger.Development
	}
}
