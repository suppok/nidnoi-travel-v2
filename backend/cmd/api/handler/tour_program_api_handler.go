package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"backend/pkg/constants/apicode"
	"backend/pkg/core/api"
	"backend/pkg/core/api/middleware"
	coremodels "backend/pkg/core/api/models"
	"backend/pkg/modules/tourprogram"
	"backend/pkg/modules/tourprogram/models"
)

type TourProgramAPIHandler struct {
	router             fiber.Router
	apiHandler         *api.ApiHandler
	tourProgramUseCase tourprogram.TourProgramUseCase
}

func NewTourProgramAPIHandler(router fiber.Router, apiHandler *api.ApiHandler, tourProgramUseCase tourprogram.TourProgramUseCase) *TourProgramAPIHandler {
	return &TourProgramAPIHandler{router: router, apiHandler: apiHandler, tourProgramUseCase: tourProgramUseCase}
}

func (h *TourProgramAPIHandler) Init(jwtAuth middleware.Auth, authPermissions middleware.AuthPermissions) {
	r := h.router

	// --- Staff (backoffice) — fixed paths before "/:id". Public endpoints are not built yet. ---
	r.Post("/tour-programs", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.TourProgramCreate), h.TourProgramCreate)
	r.Put("/tour-programs", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.TourProgramUpdate), h.TourProgramUpdate)
	r.Post("/tour-programs/list", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.TourProgramList), h.TourProgramList)
	r.Put("/tour-programs/status", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.TourProgramStatus), h.TourProgramUpdateStatus)
	r.Get("/tour-programs/:id", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.TourProgramGet), h.TourProgramGet)
}

// TourProgramList godoc
// @Summary  List tour programs
// @Tags     tour-programs
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body models.TourProgramListRequest true "Filter"
// @Success  200 {object} models.TourProgramListResponse
// @Router   /tour-programs/list [post]
func (h *TourProgramAPIHandler) TourProgramList(c *fiber.Ctx) error {
	request := &models.TourProgramListRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.tourProgramUseCase.List(ctx, request)
	})
}

// TourProgramGet godoc
// @Summary  Get a tour program with its countries, cities and gallery
// @Tags     tour-programs
// @Produce  json
// @Security BearerAuth
// @Param    id path string true "Tour program ID"
// @Success  200 {object} models.TourProgramGetResponse
// @Failure  404 {object} map[string]any "NOT_FOUND"
// @Router   /tour-programs/{id} [get]
func (h *TourProgramAPIHandler) TourProgramGet(c *fiber.Ctx) error {
	request := &models.TourProgramGetRequest{ID: c.Params("id")}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.tourProgramUseCase.Get(ctx, request)
	})
}

// TourProgramCreate godoc
// @Summary     Create a tour program
// @Description The public code is generated (tour_code_seq) and never changes. State defaults to 10 (draft).
// @Description Media fields take URLs returned by POST /media/upload.
// @Tags        tour-programs
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.TourProgramCreateRequest true "Tour program"
// @Success     200 {object} models.TourProgramCreateResponse
// @Failure     400 {object} map[string]any "VALIDATION"
// @Failure     409 {object} map[string]any "REF_ID_TAKEN"
// @Router      /tour-programs [post]
func (h *TourProgramAPIHandler) TourProgramCreate(c *fiber.Ctx) error {
	request := &models.TourProgramCreateRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.tourProgramUseCase.Create(ctx, request)
	})
}

// TourProgramUpdate godoc
// @Summary     Update a tour program
// @Description Replaces every editable field, including the country, city and gallery lists. The code never changes.
// @Tags        tour-programs
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.TourProgramUpdateRequest true "Tour program"
// @Success     200 {object} models.TourProgramUpdateResponse
// @Failure     400 {object} map[string]any "VALIDATION"
// @Failure     404 {object} map[string]any "NOT_FOUND"
// @Failure     409 {object} map[string]any "REF_ID_TAKEN"
// @Router      /tour-programs [put]
func (h *TourProgramAPIHandler) TourProgramUpdate(c *fiber.Ctx) error {
	request := &models.TourProgramUpdateRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.tourProgramUseCase.Update(ctx, request)
	})
}

// TourProgramUpdateStatus godoc
// @Summary  Soft-delete or restore a tour program
// @Tags     tour-programs
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body models.TourProgramUpdateStatusRequest true "Status"
// @Success  200 {object} map[string]any "data: null"
// @Failure  404 {object} map[string]any "NOT_FOUND"
// @Router   /tour-programs/status [put]
func (h *TourProgramAPIHandler) TourProgramUpdateStatus(c *fiber.Ctx) error {
	request := &models.TourProgramUpdateStatusRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.tourProgramUseCase.UpdateStatus(ctx, request)
	})
}
