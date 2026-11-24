// Package anthropic calls the Claude API (Messages) for AI features. The API key and model come from
// the caller on every call (AI_SETTINGS is read from the database each time), so no key is held here.
// The key and the documents sent are never logged.
package anthropic

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	sdk "github.com/anthropics/anthropic-sdk-go/v2"
	"github.com/anthropics/anthropic-sdk-go/v2/option"
	"go.uber.org/zap"

	"backend/pkg/core/config"
)

var (
	// ErrInvalidAPIKey: Anthropic rejected the key (401) or it lacks access (403).
	ErrInvalidAPIKey = errors.New("anthropic: invalid API key")
	// ErrModelNotFound: the model is unknown or not available to this key (404).
	ErrModelNotFound = errors.New("anthropic: model not found")
	// ErrBusy: still rate limited (429) or overloaded (529) after the retries.
	ErrBusy = errors.New("anthropic: rate limited or overloaded")
	// ErrRefused: the model declined the request (stop_reason refusal).
	ErrRefused = errors.New("anthropic: request refused")
	// ErrTruncated: the output hit max_tokens before the JSON was complete.
	ErrTruncated = errors.New("anthropic: output truncated")
	// ErrTimeout: the call ran past config ai.requestTimeout.
	ErrTimeout = errors.New("anthropic: timed out")
)

// RequestError is any other API error (400 too large/invalid PDF, 5xx, network). Message is
// Anthropic's error message; it never contains the key.
type RequestError struct {
	Status  int // 0 = no response (network)
	Message string
}

func (e *RequestError) Error() string {
	return fmt.Sprintf("anthropic: request failed (status %d): %s", e.Status, e.Message)
}

// SystemBlock is one block of the system prompt. Cache marks a prompt-cache breakpoint after it
// (put stable text first).
type SystemBlock struct {
	Text  string
	Cache bool
}

// ExtractRequest asks the model to read one PDF and answer with JSON matching Schema
// (structured outputs: the reply is constrained to the schema, never parsed out of free text).
type ExtractRequest struct {
	APIKey      string
	Model       string
	System      []SystemBlock
	PDF         []byte
	Instruction string          // user text after the document
	Schema      json.RawMessage // JSON schema of the answer
	MaxTokens   int64
}

type Usage struct {
	InputTokens              int64 `json:"inputTokens"`
	OutputTokens             int64 `json:"outputTokens"`
	CacheCreationInputTokens int64 `json:"cacheCreationInputTokens"`
	CacheReadInputTokens     int64 `json:"cacheReadInputTokens"`
}

type ExtractResult struct {
	JSON      []byte // the answer, valid against Schema
	Model     string // the model that answered (a server-side fallback may differ from the request)
	RequestID string
	Usage     Usage
}

type Client interface {
	// Ping checks that apiKey works and can use model (GET /v1/models/{model}, no tokens billed).
	Ping(ctx context.Context, apiKey, model string) error
	// ExtractJSON runs one extraction. At most config ai.maxConcurrent run at once; others wait.
	ExtractJSON(ctx context.Context, req ExtractRequest) (*ExtractResult, error)
}

type client struct {
	cfg     config.AIConfig
	log     *zap.Logger
	slots   chan struct{}
	backoff []time.Duration
	baseURL string // tests only
}

// New returns the client. It holds no credentials.
func New(cfg config.AIConfig, log *zap.Logger) Client {
	return &client{
		cfg:     cfg,
		log:     log,
		slots:   make(chan struct{}, cfg.MaxConcurrent),
		backoff: []time.Duration{2 * time.Second, 5 * time.Second, 15 * time.Second, 30 * time.Second},
	}
}

const pingTimeout = 15 * time.Second

func (c *client) sdkClient(apiKey string, timeout time.Duration) sdk.Client {
	opts := []option.RequestOption{
		// Nothing from ANTHROPIC_* environment variables or profiles: the key and the endpoint are
		// only ever the admin's key and api.anthropic.com.
		option.WithoutEnvironmentDefaults(),
		option.WithAPIKey(apiKey),
		option.WithMaxRetries(0), // retries are ours: 429/529 only
		option.WithRequestTimeout(timeout),
	}
	if c.baseURL != "" {
		opts = append(opts, option.WithBaseURL(c.baseURL))
	}
	return sdk.NewClient(opts...)
}

func (c *client) Ping(ctx context.Context, apiKey, model string) error {
	ctx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()
	api := c.sdkClient(apiKey, pingTimeout)
	_, err := api.Models.Get(ctx, model, sdk.ModelGetParams{})
	return mapError(ctx, err)
}

func (c *client) ExtractJSON(ctx context.Context, req ExtractRequest) (*ExtractResult, error) {
	ctx, cancel := context.WithTimeout(ctx, c.cfg.RequestTimeout)
	defer cancel()

	select {
	case c.slots <- struct{}{}:
		defer func() { <-c.slots }()
	case <-ctx.Done():
		return nil, mapError(ctx, ctx.Err())
	}

	params := buildParams(req)
	api := c.sdkClient(req.APIKey, c.cfg.RequestTimeout)
	for attempt := 0; ; attempt++ {
		result, err := c.stream(ctx, api, params)
		if err == nil {
			return result, nil
		}
		var apiErr *sdk.Error
		retryable := errors.As(err, &apiErr) && (apiErr.StatusCode == http.StatusTooManyRequests || apiErr.StatusCode == 529)
		if !retryable || attempt >= c.cfg.MaxRetries {
			if retryable {
				return nil, ErrBusy
			}
			return nil, mapError(ctx, err)
		}
		wait := c.backoff[min(attempt, len(c.backoff)-1)]
		if ra := retryAfter(apiErr); ra > 0 {
			wait = min(ra, 60*time.Second)
		}
		c.log.Warn("anthropic busy, retrying", zap.Int("status", apiErr.StatusCode), zap.Int("attempt", attempt+1), zap.Duration("wait", wait))
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return nil, mapError(ctx, ctx.Err())
		}
	}
}

// stream sends the request as a stream (long outputs never hit an HTTP timeout) and returns the
// accumulated message.
func (c *client) stream(ctx context.Context, api sdk.Client, params sdk.BetaMessageNewParams) (*ExtractResult, error) {
	s := api.Beta.Messages.NewStreaming(ctx, params)
	defer s.Close()
	var msg sdk.BetaMessage
	for s.Next() {
		if err := msg.Accumulate(s.Current()); err != nil {
			return nil, err
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	switch msg.StopReason {
	case sdk.BetaStopReasonRefusal:
		return nil, ErrRefused
	case sdk.BetaStopReasonMaxTokens, sdk.BetaStopReasonModelContextWindowExceeded:
		return nil, ErrTruncated
	}
	var text strings.Builder
	for _, block := range msg.Content {
		if block.Type == "text" {
			text.WriteString(block.Text)
		}
	}
	out := strings.TrimSpace(text.String())
	if out == "" || !json.Valid([]byte(out)) {
		return nil, &RequestError{Status: http.StatusOK, Message: "the answer was not JSON"}
	}
	return &ExtractResult{
		JSON:      []byte(out),
		Model:     string(msg.Model),
		RequestID: msg.ID,
		Usage: Usage{
			InputTokens:              msg.Usage.InputTokens,
			OutputTokens:             msg.Usage.OutputTokens,
			CacheCreationInputTokens: msg.Usage.CacheCreationInputTokens,
			CacheReadInputTokens:     msg.Usage.CacheReadInputTokens,
		},
	}, nil
}

func buildParams(req ExtractRequest) sdk.BetaMessageNewParams {
	system := make([]sdk.BetaTextBlockParam, 0, len(req.System))
	for _, b := range req.System {
		block := sdk.BetaTextBlockParam{Text: b.Text}
		if b.Cache {
			block.CacheControl = sdk.NewBetaCacheControlEphemeralParam()
		}
		system = append(system, block)
	}
	return sdk.BetaMessageNewParams{
		Model:     sdk.Model(req.Model),
		MaxTokens: req.MaxTokens,
		System:    system,
		Messages: []sdk.BetaMessageParam{
			sdk.NewBetaUserMessage(
				sdk.NewBetaDocumentBlock(sdk.BetaBase64PDFSourceParam{Data: base64.StdEncoding.EncodeToString(req.PDF)}),
				sdk.NewBetaTextBlock(req.Instruction),
			),
		},
		OutputConfig: sdk.BetaOutputConfigParam{
			Effort: sdk.BetaOutputConfigEffortHigh,
			Format: sdk.BetaJSONOutputFormatParam{Schema: req.Schema},
		},
		// A request a safety classifier declines is retried by the API on a fallback model.
		Fallbacks: sdk.BetaFallbacksParamOfDefault(),
		Betas:     []sdk.AnthropicBeta{sdk.AnthropicBetaServerSideFallback2026_07_01},
	}
}

func retryAfter(e *sdk.Error) time.Duration {
	if e == nil || e.Response == nil {
		return 0
	}
	if s, err := strconv.Atoi(strings.TrimSpace(e.Response.Header.Get("retry-after"))); err == nil && s > 0 {
		return time.Duration(s) * time.Second
	}
	return 0
}

// mapError turns SDK errors into this package's errors. API error messages are kept; the request
// (which carries the key header) is never included.
func mapError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	var own *RequestError
	if errors.Is(err, ErrRefused) || errors.Is(err, ErrTruncated) || errors.As(err, &own) {
		return err
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return ErrTimeout
	}
	var apiErr *sdk.Error
	if errors.As(err, &apiErr) {
		switch apiErr.StatusCode {
		case http.StatusUnauthorized, http.StatusForbidden:
			return ErrInvalidAPIKey
		case http.StatusNotFound:
			return ErrModelNotFound
		case http.StatusTooManyRequests, 529:
			return ErrBusy
		}
		return &RequestError{Status: apiErr.StatusCode, Message: apiErr.Message}
	}
	if errors.Is(err, context.Canceled) {
		return err
	}
	return &RequestError{Message: "network error"}
}
