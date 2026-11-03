package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"backend/pkg/constants/apicode"
	"backend/pkg/core/api"
	"backend/pkg/core/api/middleware"
	coremodels "backend/pkg/core/api/models"
	"backend/pkg/modules/airline"
	"backend/pkg/modules/airline/models"
)

type AirlineAPIHandler struct {
	router         fiber.Router
	apiHandler     *api.ApiHandler
	airlineUseCase airline.AirlineUseCase
}

func NewAirlineAPIHandler(router fiber.Router, apiHandler *api.ApiHandler, airlineUseCase airline.AirlineUseCase) *AirlineAPIHandler {
	return &AirlineAPIHandler{router: router, apiHandler: apiHandler, airlineUseCase: airlineUseCase}
}

func (h *AirlineAPIHandler) Init(jwtAuth middleware.Auth, authPermissions middleware.AuthPermissions) {
	r := h.router

	// --- Staff (backoffice) — fixed paths before "/:id". ---
	r.Post("/airlines", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.AirlineCreate), h.AirlineCreate)
	r.Put("/airlines", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.AirlineUpdate), h.AirlineUpdate)
	r.Post("/airlines/list", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.AirlineList), h.AirlineList)
	r.Put("/airlines/status", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.AirlineStatus), h.AirlineUpdateStatus)
	r.Get("/airlines/:id", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.AirlineGet), h.AirlineGet)
}

// AirlineList godoc
// @Summary  List airlines
// @Tags     airlines
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body models.AirlineListRequest true "Filter"
// @Success  200 {object} models.AirlineListResponse
// @Router   /airlines/list [post]
func (h *AirlineAPIHandler) AirlineList(c *fiber.Ctx) error {
	request := &models.AirlineListRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.airlineUseCase.List(ctx, request)
	})
}

// AirlineGet godoc
// @Summary  Get an airline
// @Tags     airlines
// @Produce  json
// @Security BearerAuth
// @Param    id path string true "Airline ID"
// @Success  200 {object} models.AirlineGetResponse
// @Failure  404 {object} map[string]any "NOT_FOUND"
// @Router   /airlines/{id} [get]
func (h *AirlineAPIHandler) AirlineGet(c *fiber.Ctx) error {
	request := &models.AirlineGetRequest{ID: c.Params("id")}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.airlineUseCase.Get(ctx, request)
	})
}

// AirlineCreate godoc
// @Summary     Create an airline
// @Description The code (IATA, or ICAO when there is none) is stored uppercase. The logo must be uploaded first through POST /media/upload (purpose airline-logo).
// @Tags        airlines
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.AirlineCreateRequest true "Airline"
// @Success     200 {object} models.AirlineCreateResponse
// @Failure     400 {object} map[string]any "VALIDATION"
// @Failure     409 {object} map[string]any "CODE_TAKEN, REF_ID_TAKEN"
// @Router      /airlines [post]
func (h *AirlineAPIHandler) AirlineCreate(c *fiber.Ctx) error {
	request := &models.AirlineCreateRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.airlineUseCase.Create(ctx, request)
	})
}

// AirlineUpdate godoc
// @Summary     Update an airline
// @Description Replaces every editable field.
// @Tags        airlines
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.AirlineUpdateRequest true "Airline"
// @Success     200 {object} models.AirlineUpdateResponse
// @Failure     400 {object} map[string]any "VALIDATION"
// @Failure     404 {object} map[string]any "NOT_FOUND"
// @Failure     409 {object} map[string]any "CODE_TAKEN, REF_ID_TAKEN"
// @Router      /airlines [put]
func (h *AirlineAPIHandler) AirlineUpdate(c *fiber.Ctx) error {
	request := &models.AirlineUpdateRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.airlineUseCase.Update(ctx, request)
	})
}

// AirlineUpdateStatus godoc
// @Summary  Soft-delete or restore an airline
// @Tags     airlines
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body models.AirlineUpdateStatusRequest true "Status"
// @Success  200 {object} map[string]any "data: null"
// @Failure  404 {object} map[string]any "NOT_FOUND"
// @Router   /airlines/status [put]
func (h *AirlineAPIHandler) AirlineUpdateStatus(c *fiber.Ctx) error {
	request := &models.AirlineUpdateStatusRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.airlineUseCase.UpdateStatus(ctx, request)
	})
}
