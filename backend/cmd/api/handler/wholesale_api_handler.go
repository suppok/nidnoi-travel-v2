package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"backend/pkg/constants/apicode"
	"backend/pkg/core/api"
	"backend/pkg/core/api/middleware"
	coremodels "backend/pkg/core/api/models"
	"backend/pkg/modules/wholesale"
	"backend/pkg/modules/wholesale/models"
)

type WholesaleAPIHandler struct {
	router           fiber.Router
	apiHandler       *api.ApiHandler
	wholesaleUseCase wholesale.WholesaleUseCase
}

func NewWholesaleAPIHandler(router fiber.Router, apiHandler *api.ApiHandler, wholesaleUseCase wholesale.WholesaleUseCase) *WholesaleAPIHandler {
	return &WholesaleAPIHandler{router: router, apiHandler: apiHandler, wholesaleUseCase: wholesaleUseCase}
}

func (h *WholesaleAPIHandler) Init(jwtAuth middleware.Auth, authPermissions middleware.AuthPermissions) {
	r := h.router

	// --- Staff (backoffice) — fixed paths before "/:id". ---
	r.Post("/wholesales", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.WholesaleCreate), h.WholesaleCreate)
	r.Put("/wholesales", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.WholesaleUpdate), h.WholesaleUpdate)
	r.Post("/wholesales/list", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.WholesaleList), h.WholesaleList)
	r.Put("/wholesales/status", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.WholesaleStatus), h.WholesaleUpdateStatus)
	r.Get("/wholesales/:id", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.WholesaleGet), h.WholesaleGet)
}

// WholesaleList godoc
// @Summary  List wholesales with their contacts
// @Tags     wholesales
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body models.WholesaleListRequest true "Filter"
// @Success  200 {object} models.WholesaleListResponse
// @Router   /wholesales/list [post]
func (h *WholesaleAPIHandler) WholesaleList(c *fiber.Ctx) error {
	request := &models.WholesaleListRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.wholesaleUseCase.List(ctx, request)
	})
}

// WholesaleGet godoc
// @Summary  Get a wholesale with its contacts
// @Tags     wholesales
// @Produce  json
// @Security BearerAuth
// @Param    id path string true "Wholesale ID"
// @Success  200 {object} models.WholesaleGetResponse
// @Failure  404 {object} map[string]any "NOT_FOUND"
// @Router   /wholesales/{id} [get]
func (h *WholesaleAPIHandler) WholesaleGet(c *fiber.Ctx) error {
	request := &models.WholesaleGetRequest{ID: c.Params("id")}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.wholesaleUseCase.Get(ctx, request)
	})
}

// WholesaleCreate godoc
// @Summary     Create a wholesale
// @Description The code is stored uppercase. State defaults to 10 (on request).
// @Tags        wholesales
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.WholesaleCreateRequest true "Wholesale"
// @Success     200 {object} models.WholesaleCreateResponse
// @Failure     400 {object} map[string]any "VALIDATION"
// @Failure     409 {object} map[string]any "CODE_TAKEN, REF_ID_TAKEN"
// @Router      /wholesales [post]
func (h *WholesaleAPIHandler) WholesaleCreate(c *fiber.Ctx) error {
	request := &models.WholesaleCreateRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.wholesaleUseCase.Create(ctx, request)
	})
}

// WholesaleUpdate godoc
// @Summary     Update a wholesale
// @Description Replaces every field, including the whole contact list.
// @Tags        wholesales
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.WholesaleUpdateRequest true "Wholesale"
// @Success     200 {object} models.WholesaleUpdateResponse
// @Failure     400 {object} map[string]any "VALIDATION"
// @Failure     404 {object} map[string]any "NOT_FOUND"
// @Failure     409 {object} map[string]any "CODE_TAKEN, REF_ID_TAKEN"
// @Router      /wholesales [put]
func (h *WholesaleAPIHandler) WholesaleUpdate(c *fiber.Ctx) error {
	request := &models.WholesaleUpdateRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.wholesaleUseCase.Update(ctx, request)
	})
}

// WholesaleUpdateStatus godoc
// @Summary  Soft-delete or restore a wholesale
// @Tags     wholesales
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body models.WholesaleUpdateStatusRequest true "Status"
// @Success  200 {object} map[string]any "data: null"
// @Failure  404 {object} map[string]any "NOT_FOUND"
// @Router   /wholesales/status [put]
func (h *WholesaleAPIHandler) WholesaleUpdateStatus(c *fiber.Ctx) error {
	request := &models.WholesaleUpdateStatusRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.wholesaleUseCase.UpdateStatus(ctx, request)
	})
}
