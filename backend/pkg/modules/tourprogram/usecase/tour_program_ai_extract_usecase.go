package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"backend/pkg/constants/errorcode"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/utils"
	"backend/pkg/domain/repomodels"
	"backend/pkg/modules/shared/client/anthropic"
	"backend/pkg/modules/shared/services/aisettings"
	"backend/pkg/modules/shared/services/asyncjob"
	"backend/pkg/modules/tourprogram/commons"
	"backend/pkg/modules/tourprogram/models"
)

const (
	aiExtractJobKind = "tour-program-ai-extract"
	// aiExtractMaxTokens bounds the answer: a 40-departure table plus extras is ~8k tokens.
	aiExtractMaxTokens = 32000
	lookupPageSize     = 10000
)

// AIExtractConfig tells the create page whether the AI prefill is available.
func (u *tourProgramUseCase) AIExtractConfig(ctx context.Context, _ *models.TourProgramAIExtractConfigRequest) (*models.TourProgramAIExtractConfigResponse, error) {
	svc := u.shareRegistry.ServiceRegistry.AISettings
	s, err := svc.Load(ctx)
	if err != nil {
		return nil, err
	}
	return &models.TourProgramAIExtractConfigResponse{Ready: svc.Configured() && s.Ready(), MaxPdfSizeMb: s.MaxPdfSizeMb}, nil
}

// AIExtract checks the PDF and starts the extraction in the background. The AI never writes to
// the database: the result is a draft staff review and save through POST /tour-programs.
func (u *tourProgramUseCase) AIExtract(ctx context.Context, request *models.TourProgramAIExtractRequest) (*models.TourProgramAIExtractStartResponse, error) {
	svc := u.shareRegistry.ServiceRegistry.AISettings
	settings, err := svc.Load(ctx)
	if err != nil {
		return nil, err
	}
	if !svc.Configured() || !settings.Ready() {
		return nil, aisettings.NotConfigured("")
	}

	if request.File == nil {
		return nil, cerrors.FieldError("file", "a PDF file is required")
	}
	maxBytes := int64(settings.MaxPdfSizeMb) * 1024 * 1024
	if request.File.Size > maxBytes {
		e := cerrors.New(http.StatusRequestEntityTooLarge, errorcode.MediaTooLarge)
		e.Message += ": at most " + strconv.Itoa(settings.MaxPdfSizeMb) + " MB"
		return nil, e
	}
	if ct := request.File.Header.Get("Content-Type"); ct != "" && !strings.HasPrefix(ct, "application/pdf") && ct != "application/octet-stream" {
		return nil, cerrors.New(http.StatusBadRequest, errorcode.MediaTypeNotAllowed)
	}
	f, err := request.File.Open()
	if err != nil {
		return nil, cerrors.Wrap(err, "open upload")
	}
	pdf, err := io.ReadAll(io.LimitReader(f, maxBytes+1))
	_ = f.Close()
	if err != nil {
		return nil, cerrors.Wrap(err, "read upload")
	}
	if int64(len(pdf)) > maxBytes {
		return nil, cerrors.New(http.StatusRequestEntityTooLarge, errorcode.MediaTooLarge)
	}
	// By content, never the name: a PDF starts with %PDF- (within the first 1 KB per the spec).
	head := pdf[:min(len(pdf), 1024)]
	if !bytes.Contains(head, []byte("%PDF-")) {
		return nil, cerrors.New(http.StatusBadRequest, errorcode.MediaTypeNotAllowed)
	}

	userID := request.UserRequestInfo.UserID
	fileName := request.File.Filename
	jobID := u.shareRegistry.ServiceRegistry.AsyncJobs.Start(aiExtractJobKind, userID, func(ctx context.Context) (any, error) {
		return u.runAIExtract(ctx, settings, pdf, fileName, userID)
	})
	u.coreRegistry.Logger.Info("ai_extract_started", zap.String("jobId", jobID), zap.String("userId", userID),
		zap.Int("pdfBytes", len(pdf)), zap.String("model", settings.Model))
	return &models.TourProgramAIExtractStartResponse{JobID: jobID, State: string(asyncjob.StatePending)}, nil
}

// AIExtractResult returns the job state, and the draft once done. Only the user who started it may read it.
func (u *tourProgramUseCase) AIExtractResult(_ context.Context, request *models.TourProgramAIExtractResultRequest) (*models.TourProgramAIExtractResultResponse, error) {
	job, ok := u.shareRegistry.ServiceRegistry.AsyncJobs.Get(request.JobID, aiExtractJobKind, request.UserRequestInfo.UserID)
	if !ok {
		return nil, cerrors.NotFound()
	}
	res := &models.TourProgramAIExtractResultResponse{JobID: job.ID, State: string(job.State)}
	switch job.State {
	case asyncjob.StateDone:
		result, _ := job.Result.(*models.TourProgramAIExtractResult)
		res.Result = result
	case asyncjob.StateFailed:
		var appErr *cerrors.AppError
		if errors.As(job.Err, &appErr) {
			res.Error = &models.TourProgramAIExtractFailure{Code: appErr.Code, Message: appErr.Message}
		} else {
			res.Error = &models.TourProgramAIExtractFailure{Code: errorcode.Internal.Code, Message: errorcode.Internal.Message}
		}
	}
	return res, nil
}

func (u *tourProgramUseCase) runAIExtract(ctx context.Context, settings *aisettings.Settings, pdf []byte, fileName, userID string) (*models.TourProgramAIExtractResult, error) {
	log := u.coreRegistry.Logger.With(zap.String("userId", userID), zap.String("model", settings.Model),
		zap.String("promptVersion", commons.AIExtractPromptVersion))
	started := time.Now()

	out, err := u.shareRegistry.ClientRegistry.Anthropic.ExtractJSON(ctx, anthropic.ExtractRequest{
		APIKey:      settings.APIKey,
		Model:       settings.Model,
		System:      []anthropic.SystemBlock{{Text: commons.AIExtractSystemPrompt, Cache: true}},
		PDF:         pdf,
		Instruction: commons.AIExtractInstruction,
		Schema:      commons.AIExtractSchema,
		MaxTokens:   aiExtractMaxTokens,
	})
	if err != nil {
		log.Warn("ai_extract_failed", zap.Duration("duration", time.Since(started)), zap.Error(err))
		return nil, aisettings.MapClientError(err)
	}
	log = log.With(zap.String("answeredBy", out.Model), zap.String("messageId", out.RequestID),
		zap.Int64("inputTokens", out.Usage.InputTokens), zap.Int64("outputTokens", out.Usage.OutputTokens),
		zap.Int64("cacheReadInputTokens", out.Usage.CacheReadInputTokens),
		zap.Int64("cacheCreationInputTokens", out.Usage.CacheCreationInputTokens))

	var extraction commons.AIExtraction
	if err := json.Unmarshal(out.JSON, &extraction); err != nil {
		log.Error("ai_extract_bad_answer", zap.Error(err))
		e := cerrors.New(http.StatusBadGateway, errorcode.AIRequestFailed)
		e.Message += ": the answer did not match the schema"
		e.Err = err
		return nil, e
	}
	normalized, warnings := commons.NormalizeAIExtraction(&extraction, utils.TodayInThailand(time.Now()))
	lookup, err := u.aiLookup(ctx)
	if err != nil {
		return nil, err
	}
	draft, resolveWarnings := commons.ResolveAIDraft(normalized, lookup)
	warnings = append(warnings, resolveWarnings...)

	result := &models.TourProgramAIExtractResult{
		Draft:      *draft,
		Warnings:   warnings,
		Confidence: normalized.Confidence,
		Usage: models.TourProgramAIUsage{
			InputTokens:              out.Usage.InputTokens,
			OutputTokens:             out.Usage.OutputTokens,
			CacheReadInputTokens:     out.Usage.CacheReadInputTokens,
			CacheCreationInputTokens: out.Usage.CacheCreationInputTokens,
		},
		Model:      out.Model,
		FileName:   fileName,
		DurationMs: time.Since(started).Milliseconds(),
	}
	if result.Warnings == nil {
		result.Warnings = []commons.AIWarning{}
	}
	log.Info("ai_extract_finished", zap.Duration("duration", time.Since(started)),
		zap.Int("periods", len(draft.Periods)), zap.Int("warnings", len(result.Warnings)))
	return result, nil
}

// aiLookup loads the active countries, cities and airlines names are matched against.
func (u *tourProgramUseCase) aiLookup(ctx context.Context) (commons.AILookup, error) {
	active := true
	reg := u.repositoryRegistry
	var lookup commons.AILookup

	countries, _, err := reg.CountryRepository.List(ctx, repomodels.CountryListCriteria{Status: &active, Page: 1, PageSize: lookupPageSize})
	if err != nil {
		return lookup, cerrors.Wrap(err, "list countries")
	}
	for _, c := range countries {
		lookup.Countries = append(lookup.Countries, commons.LookupCountry{ID: c.ID, Slug: c.Slug, NameTh: c.NameTh, NameEn: c.NameEn})
	}
	cities, _, err := reg.CityRepository.List(ctx, repomodels.CityListCriteria{Status: &active, Page: 1, PageSize: lookupPageSize})
	if err != nil {
		return lookup, cerrors.Wrap(err, "list cities")
	}
	for _, c := range cities {
		lookup.Cities = append(lookup.Cities, commons.LookupCity{ID: c.ID, CountryID: c.CountryID, Slug: c.Slug, NameTh: c.NameTh, NameEn: c.NameEn})
	}
	airlines, _, err := reg.AirlineRepository.List(ctx, repomodels.AirlineListCriteria{Status: &active, Page: 1, PageSize: lookupPageSize})
	if err != nil {
		return lookup, cerrors.Wrap(err, "list airlines")
	}
	for _, a := range airlines {
		lookup.Airlines = append(lookup.Airlines, commons.LookupCode{ID: a.ID, Code: a.Code, Name: a.Name})
	}
	lookup.ExistingTourCode = func(code string) bool {
		found, _, err := reg.TourProgramRepository.List(ctx, repomodels.TourProgramListCriteria{Keyword: code, Page: 1, PageSize: 5})
		if err != nil {
			u.coreRegistry.Logger.Warn("ai_extract_code_lookup_failed", zap.Error(err))
			return false
		}
		for _, t := range found {
			if strings.EqualFold(t.Code, code) {
				return true
			}
		}
		return false
	}
	return lookup, nil
}
