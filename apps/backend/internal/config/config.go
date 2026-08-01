package config

import (
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	_ "github.com/joho/godotenv/autoload"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
)

type Config struct {
	Primary       Primary              `koanf:"primary" validate:"required"`
	Server        ServerConfig         `koanf:"server" validate:"required"`
	Database      DatabaseConfig       `koanf:"database" validate:"required"`
	Auth          AuthConfig           `koanf:"auth" validate:"required"`
	Redis         RedisConfig          `koanf:"redis" validate:"required"`
	Integration   IntegrationConfig    `koanf:"integration" validate:"required"`
	Service       ServiceConfig        `koanf:"service"`
	RateLimit     RateLimitConfig      `koanf:"rate_limit"`
	Observability *ObservabilityConfig `koanf:"observability"`
}

type Primary struct {
	Env string `koanf:"env" validate:"required"`
}

type ServerConfig struct {
	Port               string   `koanf:"port" validate:"required"`
	ReadTimeout        int      `koanf:"read_timeout" validate:"required"`
	WriteTimeout       int      `koanf:"write_timeout" validate:"required"`
	IdleTimeout        int      `koanf:"idle_timeout" validate:"required"`
	CORSAllowedOrigins []string `koanf:"cors_allowed_origins" validate:"required"`
}

type DatabaseConfig struct {
	Host            string `koanf:"host" validate:"required"`
	Port            int    `koanf:"port" validate:"required"`
	User            string `koanf:"user" validate:"required"`
	Password        string `koanf:"password"`
	Name            string `koanf:"name" validate:"required"`
	SSLMode         string `koanf:"ssl_mode" validate:"required"`
	MaxOpenConns    int    `koanf:"max_open_conns" validate:"required"`
	MaxIdleConns    int    `koanf:"max_idle_conns" validate:"required"`
	ConnMaxLifetime int    `koanf:"conn_max_lifetime" validate:"required"`
	ConnMaxIdleTime int    `koanf:"conn_max_idle_time" validate:"required"`
}

type RedisConfig struct {
	Address string `koanf:"address" validate:"required"`
}

type IntegrationConfig struct {
	ResendAPIKey string `koanf:"resend_api_key" validate:"required"`
}

type AuthConfig struct {
	SecretKey string `koanf:"secret_key" validate:"required"`
}

type ServiceConfig struct {
	Domain              string `koanf:"domain"`
	ShortCodeLength     int    `koanf:"short_code_length"`
	DefaultExpiryHours  int    `koanf:"default_expiry_hours"`
	GeoIPDBPath         string `koanf:"geoip_db_path"`
	LinkCacheTTLSeconds int    `koanf:"link_cache_ttl_seconds"`
}

type RateLimitConfig struct {
	LinkCreatePerMin int `koanf:"link_create_per_min"`
	RedirectPerMin   int `koanf:"redirect_per_min"`
}

func LoadConfig() (*Config, error) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	k := koanf.New(".")

	err := k.Load(
		env.Provider("BOILERPLATE_", ".", func(s string) string {
			return strings.ToLower(strings.TrimPrefix(s, "BOILERPLATE_"))
		}),
		nil,
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("could not load initial env variables")
	}

	mainConfig := &Config{}

	err = k.Unmarshal("", mainConfig)
	if err != nil {
		logger.Fatal().Err(err).Msg("could not unmarshal main config")
	}

	// Application defaults keep local development ergonomic while allowing every
	// value to be overridden using the BOILERPLATE_ environment convention.
	if mainConfig.Service.ShortCodeLength == 0 {
		mainConfig.Service.ShortCodeLength = 6
	}

	if mainConfig.Service.DefaultExpiryHours == 0 {
		mainConfig.Service.DefaultExpiryHours = 24
	}

	if mainConfig.Service.LinkCacheTTLSeconds == 0 {
		mainConfig.Service.LinkCacheTTLSeconds = 300
	}

	if mainConfig.RateLimit.LinkCreatePerMin == 0 {
		mainConfig.RateLimit.LinkCreatePerMin = 10
	}

	if mainConfig.RateLimit.RedirectPerMin == 0 {
		mainConfig.RateLimit.RedirectPerMin = 60
	}

	validate := validator.New()

	err = validate.Struct(mainConfig)
	if err != nil {
		logger.Fatal().Err(err).Msg("config validation failed")
	}

	// Set default observability config if not provided
	if mainConfig.Observability == nil {
		mainConfig.Observability = DefaultObservabilityConfig()
	}

	// Override service name and environment from primary config
	mainConfig.Observability.ServiceName = "boilerplate"
	mainConfig.Observability.Environment = mainConfig.Primary.Env

	// Validate observability config
	if err := mainConfig.Observability.Validate(); err != nil {
		logger.Fatal().Err(err).Msg("invalid observability config")
	}

	return mainConfig, nil
}
