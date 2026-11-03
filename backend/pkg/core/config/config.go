package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const (
	EnvLocal = "local"
	EnvDev   = "dev"
	EnvPrd   = "prd"
)

// Config holds non-secret settings from config/config_api.yml and secrets from environment variables.
type Config struct {
	AppEnv             string   `mapstructure:"-"`
	Port               string   `mapstructure:"-"`
	DatabaseURL        string   `mapstructure:"-"`
	JWTSecret          string   `mapstructure:"-"`
	CORSAllowedOrigins []string `mapstructure:"-"`
	R2                 R2Config `mapstructure:"-"`
	// DeployHookURL is the customer Worker's Cloudflare Deploy Hook (the URL is the credential).
	// Optional: without it, rebuilds are logged and skipped (local).
	DeployHookURL string `mapstructure:"-"`

	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Auth       AuthConfig       `mapstructure:"auth"`
	TourCode   TourCodeConfig   `mapstructure:"tourCode"`
	RateLimit  RateLimitConfig  `mapstructure:"rateLimit"`
	Permission PermissionConfig `mapstructure:"permission"`
	Media      MediaConfig      `mapstructure:"media"`

	Revalidation RevalidationConfig `mapstructure:"revalidation"`
}

// RevalidationConfig drives the debounced customer-site rebuild.
type RevalidationConfig struct {
	// Quiet: fire once no new change has arrived for this long.
	Quiet time.Duration `mapstructure:"quiet"`
	// MaxWait: fire at most this long after the first pending change, even while changes keep coming.
	MaxWait time.Duration `mapstructure:"maxWait"`
	// DailyAt ("HH:MM", Thailand time) rebuilds once a day because "upcoming" periods move with the
	// calendar. Empty disables it.
	DailyAt string `mapstructure:"dailyAt"`
}

// R2Config comes from R2_* environment variables. Uploads are disabled (STORAGE_UNAVAILABLE) until
// every value is set, so the API still starts without them.
type R2Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	PublicURL       string // no trailing slash, e.g. https://dev-media.nidnoitravelv2.com
}

func (r R2Config) Configured() bool {
	return r.AccountID != "" && r.AccessKeyID != "" && r.SecretAccessKey != "" && r.Bucket != "" && r.PublicURL != ""
}

type MediaConfig struct {
	MaxUploadMB   int `mapstructure:"maxUploadMB"`
	MaxMegapixels int `mapstructure:"maxMegapixels"`
}

type ServerConfig struct {
	ReadTimeout  time.Duration `mapstructure:"readTimeout"`
	WriteTimeout time.Duration `mapstructure:"writeTimeout"`
	BodyLimitMB  int           `mapstructure:"bodyLimitMB"`
	// Header carrying the real client IP (set by Cloudflare). Falls back to the socket address when absent.
	ClientIPHeader string `mapstructure:"clientIPHeader"`
}

type DatabaseConfig struct {
	MaxOpenConns    int           `mapstructure:"maxOpenConns"`
	MaxIdleConns    int           `mapstructure:"maxIdleConns"`
	ConnMaxLifetime time.Duration `mapstructure:"connMaxLifetime"`
}

type AuthConfig struct {
	AccessTokenTTL time.Duration `mapstructure:"accessTokenTTL"`
	// Refresh token lifetime with "remember me" (persistent cookie).
	RefreshTokenTTL time.Duration `mapstructure:"refreshTokenTTL"`
	// Refresh token lifetime without "remember me" (session cookie, server-side expiry).
	SessionRefreshTokenTTL time.Duration `mapstructure:"sessionRefreshTokenTTL"`
	// A rotated refresh token presented again within this window is treated as a benign race
	// (two tabs refreshing at once), not as token theft.
	RefreshReuseGrace time.Duration `mapstructure:"refreshReuseGrace"`
	RefreshCookieName string        `mapstructure:"refreshCookieName"`
	BcryptCost        int           `mapstructure:"bcryptCost"`
}

type TourCodeConfig struct {
	Prefix string `mapstructure:"prefix"`
}

type RateLimitConfig struct {
	Login   RateLimitRule `mapstructure:"login"`
	Inquiry RateLimitRule `mapstructure:"inquiry"`
}

type RateLimitRule struct {
	Max    int           `mapstructure:"max"`
	Window time.Duration `mapstructure:"window"`
}

type PermissionConfig struct {
	ModelPath  string `mapstructure:"modelPath"`
	PolicyPath string `mapstructure:"policyPath"`
}

func (c *Config) IsLocal() bool { return c.AppEnv == EnvLocal }
func (c *Config) IsPrd() bool   { return c.AppEnv == EnvPrd }

// Load reads the YAML file and environment variables for the API, then validates the result. It fails fast.
func Load(path string) (*Config, error) {
	cfg, err := read(path)
	if err != nil {
		return nil, err
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// LoadForMigrate is for cmd/migrate, which needs only the database settings.
func LoadForMigrate(path string) (*Config, error) {
	cfg, err := read(path)
	if err != nil {
		return nil, err
	}
	if cfg.DatabaseURL == "" {
		return nil, errors.New("config: DATABASE_URL is required")
	}
	return cfg, nil
}

func read(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	cfg.AppEnv = os.Getenv("APP_ENV")
	cfg.Port = os.Getenv("PORT")
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	cfg.CORSAllowedOrigins = splitList(os.Getenv("CORS_ALLOWED_ORIGINS"))
	cfg.R2 = R2Config{
		AccountID:       os.Getenv("R2_ACCOUNT_ID"),
		AccessKeyID:     os.Getenv("R2_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
		Bucket:          os.Getenv("R2_BUCKET"),
		PublicURL:       strings.TrimRight(os.Getenv("R2_PUBLIC_URL"), "/"),
	}
	cfg.DeployHookURL = strings.TrimSpace(os.Getenv("CLOUDFLARE_DEPLOY_HOOK_URL"))
	return cfg, nil
}

func (c *Config) validate() error {
	var problems []string
	switch c.AppEnv {
	case EnvLocal, EnvDev, EnvPrd:
	default:
		problems = append(problems, fmt.Sprintf("APP_ENV must be local, dev or prd (got %q)", c.AppEnv))
	}
	if c.DatabaseURL == "" {
		problems = append(problems, "DATABASE_URL is required")
	}
	if len(c.JWTSecret) < 32 {
		problems = append(problems, "JWT_SECRET must be at least 32 characters")
	}
	if len(c.CORSAllowedOrigins) == 0 {
		problems = append(problems, "CORS_ALLOWED_ORIGINS is required")
	}
	for _, o := range c.CORSAllowedOrigins {
		if o == "*" {
			problems = append(problems, "CORS_ALLOWED_ORIGINS must list explicit origins (credentials are allowed)")
		}
	}
	if c.Auth.AccessTokenTTL <= 0 || c.Auth.RefreshTokenTTL <= 0 || c.Auth.SessionRefreshTokenTTL <= 0 {
		problems = append(problems, "auth token TTLs must be positive")
	}
	if c.Auth.RefreshCookieName == "" {
		problems = append(problems, "auth.refreshCookieName is required")
	}
	if c.Auth.BcryptCost < 10 {
		problems = append(problems, "auth.bcryptCost must be at least 10")
	}
	if c.Permission.ModelPath == "" || c.Permission.PolicyPath == "" {
		problems = append(problems, "permission.modelPath and permission.policyPath are required")
	}
	if c.Media.MaxUploadMB < 1 || c.Media.MaxUploadMB >= c.Server.BodyLimitMB {
		problems = append(problems, "media.maxUploadMB must be at least 1 and below server.bodyLimitMB")
	}
	if c.Media.MaxMegapixels < 1 {
		problems = append(problems, "media.maxMegapixels must be at least 1")
	}
	if c.R2.PublicURL != "" && !strings.HasPrefix(c.R2.PublicURL, "https://") && !c.IsLocal() {
		problems = append(problems, "R2_PUBLIC_URL must start with https://")
	}
	if c.DeployHookURL != "" && !strings.HasPrefix(c.DeployHookURL, "https://") && !c.IsLocal() {
		problems = append(problems, "CLOUDFLARE_DEPLOY_HOOK_URL must start with https://")
	}
	if c.Revalidation.Quiet <= 0 || c.Revalidation.MaxWait < c.Revalidation.Quiet {
		problems = append(problems, "revalidation.quiet must be positive and revalidation.maxWait at least quiet")
	}
	if c.Revalidation.DailyAt != "" {
		if _, err := time.Parse("15:04", c.Revalidation.DailyAt); err != nil {
			problems = append(problems, "revalidation.dailyAt must be HH:MM")
		}
	}
	if len(problems) > 0 {
		return errors.New("config: " + strings.Join(problems, "; "))
	}
	return nil
}

func splitList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}
