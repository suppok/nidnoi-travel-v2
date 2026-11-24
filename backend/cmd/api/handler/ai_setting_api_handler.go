package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"backend/pkg/constants/apicode"
	"backend/pkg/core/api"
	"backend/pkg/core/api/middleware"
	coremodels "backend/pkg/core/api/models"
	"backend/pkg/modules/aisetting"
	"backend/pkg/modules/aisetting/models"
)

type AISettingAPIHandler struct {
	router           fiber.Router
	apiHandler       *api.ApiHandler
	aiSettingUseCase aisetting.AISettingUseCase
}

func NewAISettingAPIHandler(router fiber.Router, apiHandler *api.ApiHandler, aiSettingUseCase aisetting.AISettingUseCase) *AISettingAPIHandler {
	return &AISettingAPIHandler{router: router, apiHandler: apiHandler, aiSettingUseCase: aiSettingUseCase}
}

func (h *AISettingAPIHandler) Init(jwtAuth middleware.Auth, authPermissions middleware.AuthPermissions) {
	r := h.router
	r.Get("/ai-settings", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.AISettingGet), h.AISettingGet)
	r.Put("/ai-settings", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.AISettingUpdate), h.AISettingUpdate)
	r.Post("/ai-settings/test", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.AISettingUpdate), h.AISettingTest)
}

// AISettingGet godoc
// @Summary     Get the AI settings
// @Description The API key is never returned: hasApiKey and apiKeyLast4 only. ready = AI prefill is offered to staff.
// @Tags        ai-setting
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} models.AISettingResponse
// @Router      /ai-settings [get]
func (h *AISettingAPIHandler) AISettingGet(c *fiber.Ctx) error {
	request := &models.AISettingGetRequest{}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.aiSettingUseCase.Get(ctx, request)
	})
}

// AISettingUpdate godoc
// @Summary     Change the AI settings
// @Description apiKey omitted, null or "" keeps the stored key; clearApiKey removes it. The key is stored
// @Description encrypted (AES-256-GCM). enabled needs a key. AI_NOT_CONFIGURED when the server has no
// @Description AI_SETTINGS_ENCRYPTION_KEY.
// @Tags        ai-setting
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.AISettingUpdateRequest true "Settings"
// @Success     200 {object} models.AISettingResponse
// @Failure     400 {object} map[string]any "VALIDATION"
// @Failure     503 {object} map[string]any "AI_NOT_CONFIGURED"
// @Router      /ai-settings [put]
func (h *AISettingAPIHandler) AISettingUpdate(c *fiber.Ctx) error {
	request := &models.AISettingUpdateRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.aiSettingUseCase.Update(ctx, request)
	})
}

// AISettingTest godoc
// @Summary     Test an Anthropic API key
// @Description Checks the key typed on the page (apiKey) or the stored one (apiKey empty) against the model
// @Description with GET /v1/models/{model} — no tokens billed, nothing saved.
// @Tags        ai-setting
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.AISettingTestRequest true "Key and model to test"
// @Success     200 {object} models.AISettingTestResponse
// @Failure     400 {object} map[string]any "VALIDATION, AI_INVALID_API_KEY, AI_MODEL_UNAVAILABLE"
// @Failure     502 {object} map[string]any "AI_REQUEST_FAILED"
// @Failure     503 {object} map[string]any "AI_BUSY, AI_NOT_CONFIGURED"
// @Router      /ai-settings/test [post]
func (h *AISettingAPIHandler) AISettingTest(c *fiber.Ctx) error {
	request := &models.AISettingTestRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.aiSettingUseCase.Test(ctx, request)
	})
}
