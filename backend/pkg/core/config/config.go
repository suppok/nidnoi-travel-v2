package config

import (
	"errors"
	"fmt"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"

	"backend/pkg/core/crypto"
)

const (
	EnvLocal = "local"
	EnvDev   = "dev"
	EnvPrd   = "prd"
)

// Config holds non-secret settings from config/config_api.yml and secrets from environment variables.
type Config struct {
	AppEnv             string     `mapstructure:"-"`
	Port               string     `mapstructure:"-"`
	DatabaseURL        string     `mapstructure:"-"`
	JWTSecret          string     `mapstructure:"-"`
	CORSAllowedOrigins []string   `mapstructure:"-"`
	R2                 R2Config   `mapstructure:"-"`
	LINE               LINEConfig `mapstructure:"-"`
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
	Mail         MailConfig         `mapstructure:"mail"`
	AI           AIConfig           `mapstructure:"ai"`
}

// AIConfig: the Claude API used to prefill tour programs from PDFs. The API key, model and on/off
// switch live in app config AI_SETTINGS (admin page); the key is encrypted there with
// AI_SETTINGS_ENCRYPTION_KEY. Without that variable the AI features stay off.
type AIConfig struct {
	// SettingsEncryptionKey is AI_SETTINGS_ENCRYPTION_KEY decoded (32 bytes, AES-256-GCM); nil = not set.
	SettingsEncryptionKey []byte `mapstructure:"-"`

	// RequestTimeout bounds one extraction (the Claude API call, retries included).
	RequestTimeout time.Duration `mapstructure:"requestTimeout"`
	// MaxConcurrent extractions at once; more wait for a slot (each holds the PDF in memory).
	MaxConcurrent int `mapstructure:"maxConcurrent"`
	// MaxRetries on 429 (rate limited) and 529 (overloaded) only.
	MaxRetries int `mapstructure:"maxRetries"`
	// MaxPdfMB is the hard cap for an uploaded PDF; the admin setting maxPdfSizeMb can only be lower.
	// Must stay below server.bodyLimitMB.
	MaxPdfMB int `mapstructure:"maxPdfMB"`
	// JobTTL: how long a finished extraction result is kept for the backoffice to fetch.
	JobTTL time.Duration `mapstructure:"jobTTL"`
}

// Enabled reports whether AI_SETTINGS_ENCRYPTION_KEY is set.
func (a AIConfig) Enabled() bool { return len(a.SettingsEncryptionKey) > 0 }

// MailConfig: outgoing email (Resend). The outbox settings come from config_api.yml; the API key, the
// From address and dry-run come from MAIL_* environment variables. Reply-To/Cc live in app_config.
type MailConfig struct {
	APIKey string `mapstructure:"-"` // MAIL_API_KEY (secret)
	From   string `mapstructure:"-"` // MAIL_FROM, e.g. "nidnoi travel <booking@nidnoitravelv2.com>"
	DryRun bool   `mapstructure:"-"` // MAIL_DRY_RUN: log emails instead of sending them

	HTTPTimeout time.Duration    `mapstructure:"httpTimeout"`
	Outbox      MailOutboxConfig `mapstructure:"outbox"`
}

type MailOutboxConfig struct {
	TickInterval time.Duration `mapstructure:"tickInterval"`
	BatchSize    int           `mapstructure:"batchSize"`
	MaxAttempts  int           `mapstructure:"maxAttempts"`
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

// LINEConfig comes from LINE_* environment variables. LINE integration is disabled until the
// required values are set.
type LINEConfig struct {
	ChannelSecret      string // Messaging API channel secret
	ChannelAccessToken string // Messaging API channel access token (long-lived)
	// LoginChannelID / LoginChannelSecret are the LINE Login channel's: backoffice "ล็อกอินด้วย LINE" and
	// "เชื่อมบัญชี LINE" are off without both. It must be under the same provider as the Messaging API
	// channel, or the LINE user IDs differ.
	LoginChannelID     string
	LoginChannelSecret string
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
	ConnMaxIdleTime time.Duration `mapstructure:"connMaxIdleTime"`
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
	Booking RateLimitRule `mapstructure:"booking"`
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
	cfg.LINE = LINEConfig{
		ChannelSecret:      strings.TrimSpace(os.Getenv("LINE_CHANNEL_SECRET")),
		ChannelAccessToken: strings.TrimSpace(os.Getenv("LINE_CHANNEL_ACCESS_TOKEN")),
		LoginChannelID:     strings.TrimSpace(os.Getenv("LINE_LOGIN_CHANNEL_ID")),
		LoginChannelSecret: strings.TrimSpace(os.Getenv("LINE_LOGIN_CHANNEL_SECRET")),
	}
	if s := strings.TrimSpace(os.Getenv("AI_SETTINGS_ENCRYPTION_KEY")); s != "" {
		key, err := crypto.ParseKey(s)
		if err != nil {
			return nil, fmt.Errorf("AI_SETTINGS_ENCRYPTION_KEY must be 32 random bytes in base64: %w", err)
		}
		cfg.AI.SettingsEncryptionKey = key
	}
	cfg.Mail.APIKey = strings.TrimSpace(os.Getenv("MAIL_API_KEY"))
	cfg.Mail.From = strings.TrimSpace(os.Getenv("MAIL_FROM"))
	if s := strings.TrimSpace(os.Getenv("MAIL_DRY_RUN")); s != "" {
		dryRun, err := strconv.ParseBool(s)
		if err != nil {
			return nil, fmt.Errorf("MAIL_DRY_RUN must be true or false (got %q)", s)
		}
		cfg.Mail.DryRun = dryRun
	}
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
	problems = append(problems, c.Mail.problems()...)
	if c.AI.RequestTimeout <= 0 || c.AI.JobTTL <= 0 {
		problems = append(problems, "ai.requestTimeout and ai.jobTTL must be positive")
	}
	if c.AI.MaxConcurrent < 1 || c.AI.MaxRetries < 0 {
		problems = append(problems, "ai.maxConcurrent must be at least 1 and ai.maxRetries at least 0")
	}
	if c.AI.MaxPdfMB < 1 || c.AI.MaxPdfMB >= c.Server.BodyLimitMB {
		problems = append(problems, "ai.maxPdfMB must be at least 1 and below server.bodyLimitMB")
	}
	if len(problems) > 0 {
		return errors.New("config: " + strings.Join(problems, "; "))
	}
	return nil
}

func (m MailConfig) problems() []string {
	var problems []string
	if !m.DryRun {
		if m.APIKey == "" {
			problems = append(problems, "MAIL_API_KEY is required (or set MAIL_DRY_RUN=true)")
		}
		if m.From == "" {
			problems = append(problems, "MAIL_FROM is required (or set MAIL_DRY_RUN=true)")
		}
	}
	if m.From != "" {
		if _, err := mail.ParseAddress(m.From); err != nil {
			problems = append(problems, "MAIL_FROM must be an email address, e.g. \"nidnoi travel <booking@example.com>\"")
		}
	}
	if m.HTTPTimeout <= 0 || m.Outbox.TickInterval <= 0 {
		problems = append(problems, "mail.httpTimeout and mail.outbox.tickInterval must be positive")
	}
	if m.Outbox.BatchSize < 1 || m.Outbox.BatchSize > 100 {
		problems = append(problems, "mail.outbox.batchSize must be 1-100")
	}
	if m.Outbox.MaxAttempts < 1 {
		problems = append(problems, "mail.outbox.maxAttempts must be at least 1")
	}
	return problems
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
