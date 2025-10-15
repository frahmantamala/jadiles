package internal

import (
	"encoding/base64"
	"strings"
	"time"

	"log/slog"
)

type Config struct {
	Env          string             `mapstructure:"env"`
	Name         string             `mapstructure:"name"`
	Namespace    string             `mapstructure:"namespace"`
	InstanceID   string             `mapstructure:"instance_id"`
	HTTPServer   HTTPServerConfig   `mapstructure:"http_server"`
	Database     DatabaseConfig     `mapstructure:"database"`
	Logger       LoggerConfig       `mapstructure:"logger"`
	Notification NotificationConfig `mapstructure:"notification"`
	RateLimit    RateLimitConfig    `mapstructure:"rate_limit"`
	Redis        RedisConfig        `mapstructure:"redis"`
	Swagger      SwaggerConfig      `mapstructure:"swagger"`
}

type HTTPServerConfig struct {
	Port              int           `mapstructure:"port"`
	AllowedOrigins    string        `mapstructure:"allowed_origins"`
	ReadHeaderTimeout time.Duration `mapstructure:"read_header_timeout"`
	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	IdleTimeout       time.Duration `mapstructure:"idle_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
	AuthConfig        AuthConfig    `mapstructure:"auth"`
}

type AuthConfig struct {
	AccessTokenDuration       time.Duration `mapstructure:"access_token_duration"`
	RefreshTokenDuration      time.Duration `mapstructure:"refresh_token_duration"`
	JWTSecretEncoded          string        `mapstructure:"jwt_secret_encoded"`
	RefreshTokenSecretEncoded string        `mapstructure:"refresh_token_secret_encoded"`
	Issuer                    string        `mapstructure:"issuer"`
}

type DatabaseConfig struct {
	URL          string        `mapstructure:"url"`
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	DBName       string        `mapstructure:"db_name"`
	Username     string        `mapstructure:"username"`
	Password     string        `mapstructure:"password"`
	SSLMode      string        `mapstructure:"sslmode"`
	MaxOpenConns int           `mapstructure:"max_open_conns"`
	MaxIdleConns int           `mapstructure:"max_idle_conns"`
	MaxLifetime  time.Duration `mapstructure:"max_lifetime"`
}

type NotificationConfig struct {
	Provider       string         `mapstructure:"provider"`
	AppURL         string         `mapstructure:"app_url"`        // Web URL (fallback)
	MobileAppURL   string         `mapstructure:"mobile_app_url"` // Deep link for mobile app
	RequestTimeout time.Duration  `mapstructure:"request_timeout"`
	MaxRetries     int            `mapstructure:"max_retries"`
	SMTP           SMTPConfig     `mapstructure:"smtp"`
	SendGrid       SendGridConfig `mapstructure:"sendgrid"`
	Resend         ResendConfig   `mapstructure:"resend"`
}

type SMTPConfig struct {
	Host           string        `mapstructure:"host"`
	Port           int           `mapstructure:"port"`
	Username       string        `mapstructure:"username"`
	Password       string        `mapstructure:"password"`
	From           string        `mapstructure:"from"`
	UseTLS         bool          `mapstructure:"use_tls"`
	RequestTimeout time.Duration `mapstructure:"request_timeout"`
	MaxRetries     int           `mapstructure:"max_retries"`
}

type SendGridConfig struct {
	APIKey    string `mapstructure:"api_key"`
	FromEmail string `mapstructure:"from_email"`
	FromName  string `mapstructure:"from_name"`
}

type ResendConfig struct {
	APIKey         string        `mapstructure:"api_key"`
	FromEmail      string        `mapstructure:"from_email"`
	FromName       string        `mapstructure:"from_name"`
	RequestTimeout time.Duration `mapstructure:"request_timeout"`
	MaxRetries     int           `mapstructure:"max_retries"`
}

type FirebaseConfig struct {
	ProjectID            string        `mapstructure:"project_id"`
	ServiceAccountPath   string        `mapstructure:"service_account_path"`
	ServiceAccountBase64 string        `mapstructure:"service_account_base64"`
	RequestTimeout       time.Duration `mapstructure:"request_timeout"`
	MaxRetries           int           `mapstructure:"max_retries"`
}

type LoggerConfig struct {
	Level                string `mapstructure:"level"`
	Format               string `mapstructure:"format"`
	Output               string `mapstructure:"output"`
	EnableRequestLogging bool   `mapstructure:"enable_request_logging"`
	LogRequestBody       bool   `mapstructure:"log_request_body"`
	LogResponseBody      bool   `mapstructure:"log_response_body"`
}

type RedisConfig struct {
	Source              string        `mapstructure:"source"`
	PoolMaxActive       int           `mapstructure:"pool_max_active"`
	PoolMaxIdle         int           `mapstructure:"pool_max_idle"`
	PoolIdleTimeout     time.Duration `mapstructure:"pool_idle_timeout"`
	PoolMaxConnLifetime time.Duration `mapstructure:"pool_max_conn_lifetime"`
	AnalyticsRate       float64       `mapstructure:"analytics_rate"`
}

type RateLimitConfig struct {
	Global struct {
		RequestsPerMinute int `mapstructure:"requests_per_minute"`
		BurstSize         int `mapstructure:"burst_size"`
	} `mapstructure:"global"`

	AuthEndpoints struct {
		RequestsPerMinute int `mapstructure:"requests_per_minute"`
		BurstSize         int `mapstructure:"burst_size"`
	} `mapstructure:"auth_endpoints"`

	PublicEndpoints struct {
		RequestsPerMinute int `mapstructure:"requests_per_minute"`
		BurstSize         int `mapstructure:"burst_size"`
	} `mapstructure:"public_endpoints"`

	AuthenticatedEndpoints struct {
		RequestsPerMinute int `mapstructure:"requests_per_minute"`
		BurstSize         int `mapstructure:"burst_size"`
	} `mapstructure:"authenticated_endpoints"`
}

type SwaggerConfig struct {
	Enable bool `mapstructure:"enable"`
}

func (h HTTPServerConfig) GetJWTSecret() ([]byte, error) {
	return base64.StdEncoding.DecodeString(h.AuthConfig.JWTSecretEncoded)
}

func (h HTTPServerConfig) GetRefreshTokenSecret() ([]byte, error) {
	return base64.StdEncoding.DecodeString(h.AuthConfig.RefreshTokenSecretEncoded)
}

func (h *HTTPServerConfig) GetAllowedOrigins() []string {
	return strings.Split(h.AllowedOrigins, " ")
}

func (c LoggerConfig) ParseSlogLevel() slog.Level {
	switch strings.ToUpper(c.Level) {
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	case "DEBUG":
		return slog.LevelDebug
	default:
		return slog.LevelInfo
	}
}
