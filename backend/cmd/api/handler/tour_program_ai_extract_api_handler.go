package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"

	coremodels "backend/pkg/core/api/models"
	"backend/pkg/modules/tourprogram/models"
)

// TourProgramAIExtractConfig godoc
// @Summary     Is the AI prefill available?
// @Description ready = AI is on with a usable key; the create page offers "Upload PDF" only then.
// @Tags        tour-program
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} models.TourProgramAIExtractConfigResponse
// @Router      /tour-programs/ai-extract/config [get]
func (h *TourProgramAPIHandler) TourProgramAIExtractConfig(c *fiber.Ctx) error {
	request := &models.TourProgramAIExtractConfigRequest{}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.tourProgramUseCase.AIExtractConfig(ctx, request)
	})
}

// TourProgramAIExtract godoc
// @Summary     Read a tour program PDF with AI (start)
// @Description Checks the PDF (by content, at most AI setting maxPdfSizeMb) and starts reading it in the
// @Description background — it takes about 30-90 seconds. Poll GET /tour-programs/ai-extract/{jobId}.
// @Description Nothing is saved: the result prefills the create form.
// @Tags        tour-program
// @Accept      multipart/form-data
// @Produce     json
// @Security    BearerAuth
// @Param       file formData file true "Tour program PDF"
// @Success     200 {object} models.TourProgramAIExtractStartResponse
// @Failure     400 {object} map[string]any "VALIDATION, MEDIA_TYPE_NOT_ALLOWED"
// @Failure     413 {object} map[string]any "MEDIA_TOO_LARGE"
// @Failure     503 {object} map[string]any "AI_NOT_CONFIGURED"
// @Router      /tour-programs/ai-extract [post]
func (h *TourProgramAPIHandler) TourProgramAIExtract(c *fiber.Ctx) error {
	request := &models.TourProgramAIExtractRequest{}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		if fh, err := c.FormFile("file"); err == nil {
			request.File = fh
		}
		return h.tourProgramUseCase.AIExtract(ctx, request)
	})
}

// TourProgramAIExtractResult godoc
// @Summary     Result of a PDF read with AI
// @Description state pending → keep polling (every few seconds); done → result.draft prefills the create
// @Description form, with warnings and per-section confidence; failed → error {code, message}.
// @Description Only the user who started the job can read it; results are kept for config ai.jobTTL.
// @Tags        tour-program
// @Produce     json
// @Security    BearerAuth
// @Param       jobId path string true "Job id from POST /tour-programs/ai-extract"
// @Success     200 {object} models.TourProgramAIExtractResultResponse
// @Failure     404 {object} map[string]any "NOT_FOUND (unknown, expired or someone else's job)"
// @Router      /tour-programs/ai-extract/{jobId} [get]
func (h *TourProgramAPIHandler) TourProgramAIExtractResult(c *fiber.Ctx) error {
	request := &models.TourProgramAIExtractResultRequest{JobID: c.Params("jobId")}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.tourProgramUseCase.AIExtractResult(ctx, request)
	})
}
