package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"backend/pkg/constants/apicode"
	"backend/pkg/core/api"
	"backend/pkg/core/api/middleware"
	coremodels "backend/pkg/core/api/models"
	"backend/pkg/modules/tourprogramperiod"
	"backend/pkg/modules/tourprogramperiod/models"
)

type TourProgramPeriodAPIHandler struct {
	router        fiber.Router
	apiHandler    *api.ApiHandler
	periodUseCase tourprogramperiod.TourProgramPeriodUseCase
}

func NewTourProgramPeriodAPIHandler(router fiber.Router, apiHandler *api.ApiHandler, periodUseCase tourprogramperiod.TourProgramPeriodUseCase) *TourProgramPeriodAPIHandler {
	return &TourProgramPeriodAPIHandler{router: router, apiHandler: apiHandler, periodUseCase: periodUseCase}
}

func (h *TourProgramPeriodAPIHandler) Init(jwtAuth middleware.Auth, authPermissions middleware.AuthPermissions) {
	r := h.router

	// --- Staff (backoffice). The public periods endpoint (period island) is not built yet. ---
	r.Post("/tour-program-periods", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.PeriodCreate), h.PeriodCreate)
	r.Put("/tour-program-periods", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.PeriodUpdate), h.PeriodUpdate)
	r.Post("/tour-program-periods/list", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.PeriodList), h.PeriodList)
	r.Put("/tour-program-periods/status", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.PeriodStatus), h.PeriodUpdateStatus)
}

// PeriodList godoc
// @Summary  List the periods of a tour program by departure date
// @Tags     tour-program-periods
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body models.TourProgramPeriodListRequest true "Filter"
// @Success  200 {object} models.TourProgramPeriodListResponse
// @Router   /tour-program-periods/list [post]
func (h *TourProgramPeriodAPIHandler) PeriodList(c *fiber.Ctx) error {
	request := &models.TourProgramPeriodListRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.periodUseCase.List(ctx, request)
	})
}

// PeriodCreate godoc
// @Summary  Add a period to a tour program
// @Tags     tour-program-periods
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body models.TourProgramPeriodCreateRequest true "Period"
// @Success  200 {object} models.TourProgramPeriodCreateResponse
// @Failure  400 {object} map[string]any "VALIDATION"
// @Router   /tour-program-periods [post]
func (h *TourProgramPeriodAPIHandler) PeriodCreate(c *fiber.Ctx) error {
	request := &models.TourProgramPeriodCreateRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.periodUseCase.Create(ctx, request)
	})
}

// PeriodUpdate godoc
// @Summary  Update a period's dates, prices, seats and note
// @Tags     tour-program-periods
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body models.TourProgramPeriodUpdateRequest true "Period"
// @Success  200 {object} models.TourProgramPeriodUpdateResponse
// @Failure  400 {object} map[string]any "VALIDATION"
// @Failure  404 {object} map[string]any "NOT_FOUND"
// @Router   /tour-program-periods [put]
func (h *TourProgramPeriodAPIHandler) PeriodUpdate(c *fiber.Ctx) error {
	request := &models.TourProgramPeriodUpdateRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.periodUseCase.Update(ctx, request)
	})
}

// PeriodUpdateStatus godoc
// @Summary  Soft-delete or restore a period
// @Tags     tour-program-periods
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body models.TourProgramPeriodUpdateStatusRequest true "Status"
// @Success  200 {object} map[string]any "data: null"
// @Failure  404 {object} map[string]any "NOT_FOUND"
// @Router   /tour-program-periods/status [put]
func (h *TourProgramPeriodAPIHandler) PeriodUpdateStatus(c *fiber.Ctx) error {
	request := &models.TourProgramPeriodUpdateStatusRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.periodUseCase.UpdateStatus(ctx, request)
	})
}
