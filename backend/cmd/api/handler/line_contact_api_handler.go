package handler

import (
	"bytes"
	"context"

	"github.com/gofiber/fiber/v2"

	"backend/pkg/constants/apicode"
	"backend/pkg/core/api"
	"backend/pkg/core/api/middleware"
	coremodels "backend/pkg/core/api/models"
	"backend/pkg/modules/linecontact"
	"backend/pkg/modules/linecontact/models"
	"backend/pkg/modules/shared/client/line"
)

type LineContactAPIHandler struct {
	router             fiber.Router
	apiHandler         *api.ApiHandler
	lineContactUseCase linecontact.LineContactUseCase
}

func NewLineContactAPIHandler(router fiber.Router, apiHandler *api.ApiHandler, lineContactUseCase linecontact.LineContactUseCase) *LineContactAPIHandler {
	return &LineContactAPIHandler{router: router, apiHandler: apiHandler, lineContactUseCase: lineContactUseCase}
}

func (h *LineContactAPIHandler) Init(jwtAuth middleware.Auth, authPermissions middleware.AuthPermissions) {
	r := h.router

	// --- LINE Platform — no JWT; the X-Line-Signature header is verified in the use case. ---
	r.Post("/line/webhook", h.LineWebhook)

	// --- Staff (backoffice) ---
	r.Post("/line-contacts/list", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineContactList), h.LineContactList)
}

// LineWebhook godoc
// @Summary     Messaging API webhook
// @Description Called by the LINE Platform. Verifies X-Line-Signature, then records follow, unfollow
// @Description and message events from users as LINE contacts (profile only, never message text).
// @Tags        line-contacts
// @Accept      json
// @Produce     json
// @Param       X-Line-Signature header string true "Base64 HMAC-SHA256 of the body"
// @Success     200 {object} models.LineWebhookResponse
// @Failure     401 {object} map[string]any "INVALID_LINE_SIGNATURE"
// @Failure     503 {object} map[string]any "LINE_NOT_CONFIGURED"
// @Router      /line/webhook [post]
func (h *LineContactAPIHandler) LineWebhook(c *fiber.Ctx) error {
	request := &models.LineWebhookRequest{
		Body:      bytes.Clone(c.Body()),
		Signature: c.Get(line.SignatureHeader),
	}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, _ *coremodels.RequestInfo) (any, error) {
		return h.lineContactUseCase.HandleWebhook(ctx, request)
	})
}

// LineContactList godoc
// @Summary  List LINE OA contacts, most recent chat first
// @Tags     line-contacts
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body models.LineContactListRequest true "Filter"
// @Success  200 {object} models.LineContactListResponse
// @Failure  400 {object} map[string]any "VALIDATION"
// @Router   /line-contacts/list [post]
func (h *LineContactAPIHandler) LineContactList(c *fiber.Ctx) error {
	request := &models.LineContactListRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.lineContactUseCase.List(ctx, request)
	})
}
