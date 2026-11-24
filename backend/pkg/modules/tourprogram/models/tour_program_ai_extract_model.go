package models

import (
	"mime/multipart"

	coremodels "backend/pkg/core/api/models"
	"backend/pkg/modules/tourprogram/commons"
)

// --- config ---

type TourProgramAIExtractConfigRequest struct {
	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

// TourProgramAIExtractConfigResponse tells the create page whether to offer "Upload PDF (AI prefill)".
type TourProgramAIExtractConfigResponse struct {
	Ready        bool `json:"ready"`        // AI on, a usable key, encryption configured
	MaxPdfSizeMb int  `json:"maxPdfSizeMb"` // the upload limit
}

// --- start ---

// TourProgramAIExtractRequest: multipart field "file", a PDF.
type TourProgramAIExtractRequest struct {
	File *multipart.FileHeader `json:"-"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

// TourProgramAIExtractStartResponse: poll GET /tour-programs/ai-extract/{jobId} until state is
// done or failed. Extraction takes about 30-90 seconds.
type TourProgramAIExtractStartResponse struct {
	JobID string `json:"jobId"`
	State string `json:"state" enums:"pending"`
}

// --- result ---

type TourProgramAIExtractResultRequest struct {
	JobID string `json:"-"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type TourProgramAIExtractResultResponse struct {
	JobID  string                       `json:"jobId"`
	State  string                       `json:"state" enums:"pending,done,failed"`
	Result *TourProgramAIExtractResult  `json:"result"` // set when done
	Error  *TourProgramAIExtractFailure `json:"error"`  // set when failed
}

// TourProgramAIExtractResult is a prefill for the create form. Nothing has been saved.
type TourProgramAIExtractResult struct {
	Draft      commons.AIDraft      `json:"draft"`
	Warnings   []commons.AIWarning  `json:"warnings"`
	Confidence commons.AIConfidence `json:"confidence"` // per section: high | medium | low
	Usage      TourProgramAIUsage   `json:"usage"`
	Model      string               `json:"model"` // the model that answered
	FileName   string               `json:"fileName"`
	DurationMs int64                `json:"durationMs"`
}

type TourProgramAIUsage struct {
	InputTokens              int64 `json:"inputTokens"`
	OutputTokens             int64 `json:"outputTokens"`
	CacheReadInputTokens     int64 `json:"cacheReadInputTokens"`
	CacheCreationInputTokens int64 `json:"cacheCreationInputTokens"`
}

// TourProgramAIExtractFailure is the error code and message the job failed with (the same codes as
// the API's errors: AI_BUSY, AI_REQUEST_FAILED, AI_INVALID_API_KEY, AI_NOT_CONFIGURED, INTERNAL).
type TourProgramAIExtractFailure struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
