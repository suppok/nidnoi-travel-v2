// Package aisettings reads the AI settings (app config AI_SETTINGS) and seals/opens the Anthropic API
// key stored in them. Settings are read from the database on every call, never cached, so a change
// on the admin page applies to the next request.
package aisettings

import (
	"context"
	"errors"
	"time"

	"backend/pkg/constants/aimodel"
	"backend/pkg/constants/appconfigcode"
	"backend/pkg/core/config"
	"backend/pkg/core/crypto"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/repository"
)

// DefaultMaxPdfSizeMb is the PDF size limit before an admin sets one (capped by config ai.maxPdfMB).
const DefaultMaxPdfSizeMb = 50

// Settings are the decoded AI settings with the API key decrypted. Never log or return APIKey.
type Settings struct {
	Enabled      bool
	Model        string
	MaxPdfSizeMb int // effective: never above config ai.maxPdfMB
	APIKey       string
	APIKeyLast4  string
	// KeyUnreadable: a key is stored but can't be decrypted (AI_SETTINGS_ENCRYPTION_KEY changed).
	KeyUnreadable bool
	// Saved is false while no AI_SETTINGS row exists (defaults are returned).
	Saved      bool
	UpdateDate time.Time
	UpdateBy   string // user id
}

// HasAPIKey reports whether a usable key is stored.
func (s *Settings) HasAPIKey() bool { return s.APIKey != "" }

// Ready reports whether AI calls may run: switched on and a usable key.
func (s *Settings) Ready() bool { return s.Enabled && s.HasAPIKey() }

type Service interface {
	// Configured reports whether AI_SETTINGS_ENCRYPTION_KEY is set; without it AI stays off.
	Configured() bool
	// Load reads AI_SETTINGS (defaults when there is no row) and decrypts the key.
	Load(ctx context.Context) (*Settings, error)
	// SealAPIKey encrypts a key for AI_SETTINGS.apiKeyCiphertext.
	SealAPIKey(apiKey string) (string, error)
	// MaxPdfSizeMbLimit is the highest value the admin may set (config ai.maxPdfMB).
	MaxPdfSizeMbLimit() int
}

type service struct {
	repo   repository.AppConfigRepository
	sealer *crypto.Sealer // nil when not configured
	maxPdf int
}

// New returns the service. Without cfg.SettingsEncryptionKey it still loads settings but reports
// Configured() == false and never yields a key.
func New(repo repository.AppConfigRepository, cfg config.AIConfig) (Service, error) {
	s := &service{repo: repo, maxPdf: cfg.MaxPdfMB}
	if cfg.Enabled() {
		sealer, err := crypto.NewSealer(cfg.SettingsEncryptionKey)
		if err != nil {
			return nil, err
		}
		s.sealer = sealer
	}
	return s, nil
}

func (s *service) Configured() bool       { return s.sealer != nil }
func (s *service) MaxPdfSizeMbLimit() int { return s.maxPdf }

func additionalData() []byte { return []byte(appconfigcode.AppConfigAISettings.Code()) }

func (s *service) SealAPIKey(apiKey string) (string, error) {
	if s.sealer == nil {
		return "", errors.New("aisettings: AI_SETTINGS_ENCRYPTION_KEY is not set")
	}
	return s.sealer.Seal([]byte(apiKey), additionalData())
}

func (s *service) Load(ctx context.Context) (*Settings, error) {
	out := &Settings{Model: aimodel.Default, MaxPdfSizeMb: min(DefaultMaxPdfSizeMb, s.maxPdf)}
	row, err := s.repo.FindByCode(ctx, appconfigcode.AppConfigAISettings.Code())
	if err != nil {
		return nil, cerrors.Wrap(err, "find AI settings")
	}
	if row == nil || !row.Status {
		return out, nil
	}
	v, err := appconfigcode.AppConfigAISettings.ToValue(row.Value)
	if err != nil {
		return nil, cerrors.Wrap(err, "decode AI settings")
	}
	out.Saved = true
	out.Enabled = v.Enabled
	out.UpdateDate = row.UpdateDate
	out.UpdateBy = row.UpdateBy
	if aimodel.Valid(v.Model) {
		out.Model = v.Model
	}
	if v.MaxPdfSizeMb > 0 {
		out.MaxPdfSizeMb = min(v.MaxPdfSizeMb, s.maxPdf)
	}
	if v.APIKeyCiphertext != "" {
		out.APIKeyLast4 = v.APIKeyLast4
		if s.sealer == nil {
			out.KeyUnreadable = true
		} else if key, err := s.sealer.Open(v.APIKeyCiphertext, additionalData()); err != nil {
			out.KeyUnreadable = true
		} else {
			out.APIKey = string(key)
		}
	}
	return out, nil
}
