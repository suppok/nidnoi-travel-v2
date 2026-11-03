package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"backend/pkg/constants/apicode"
	"backend/pkg/core/api"
	"backend/pkg/core/api/middleware"
	coremodels "backend/pkg/core/api/models"
	"backend/pkg/modules/tourprogram"
	"backend/pkg/modules/tourprogram/models"
)

// publicPeriodsMaxAge bounds how stale seats and prices may be on the tour page (browser + edge cache).
const publicPeriodsMaxAge = 30 * time.Second

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

	// --- Public (customer site) — no auth. Registered BEFORE "/:id" routes. ---
	r.Get("/tour-programs/public", h.TourProgramPublicList)
	r.Get("/tour-programs/public/:code", h.TourProgramPublicGet)
	r.Get("/tour-programs/public/:code/periods", middleware.CacheControl(publicPeriodsMaxAge), h.TourProgramPublicPeriods)

	// --- Staff (backoffice) — fixed paths before "/:id". ---
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

// TourProgramPublicList godoc
// @Summary     List public tour programs (customer site build)
// @Description Active tours in a public state (selling, closed), ordered by code. Prices and dates
// @Description cover upcoming periods only (departing after today, Thailand time). No auth.
// @Tags        public
// @Produce     json
// @Param       country  query string false "Country slug"
// @Param       city     query string false "City slug"
// @Param       dateFrom query string false "Has an upcoming departure on/after (YYYY-MM-DD)"
// @Param       dateTo   query string false "Has an upcoming departure on/before (YYYY-MM-DD)"
// @Param       page     query int    false "Page (default 1)"
// @Param       pageSize query int    false "Page size (default 20, max 100)"
// @Success     200 {object} models.TourProgramPublicListResponse
// @Failure     400 {object} map[string]any "VALIDATION"
// @Router      /tour-programs/public [get]
func (h *TourProgramAPIHandler) TourProgramPublicList(c *fiber.Ctx) error {
	request := &models.TourProgramPublicListRequest{
		Country:  c.Query("country"),
		City:     c.Query("city"),
		DateFrom: c.Query("dateFrom"),
		DateTo:   c.Query("dateTo"),
		Page:     c.QueryInt("page"),
		PageSize: c.QueryInt("pageSize"),
	}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, _ *coremodels.RequestInfo) (any, error) {
		return h.tourProgramUseCase.PublicList(ctx, request)
	})
}

// TourProgramPublicGet godoc
// @Summary     Get a public tour program by code (customer site build)
// @Description Tour page data with gallery and upcoming periods. 404 when the code is unknown,
// @Description deleted or not public. No auth.
// @Tags        public
// @Produce     json
// @Param       code path string true "Tour code"
// @Success     200 {object} models.TourProgramPublicGetResponse
// @Failure     404 {object} map[string]any "NOT_FOUND"
// @Router      /tour-programs/public/{code} [get]
func (h *TourProgramAPIHandler) TourProgramPublicGet(c *fiber.Ctx) error {
	request := &models.TourProgramPublicGetRequest{Code: c.Params("code")}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, _ *coremodels.RequestInfo) (any, error) {
		return h.tourProgramUseCase.PublicGet(ctx, request)
	})
}

// TourProgramPublicPeriods godoc
// @Summary     Live periods of a public tour (browser)
// @Description Upcoming departures with prices and seats. Cache-Control: public, max-age=30. No auth.
// @Tags        public
// @Produce     json
// @Param       code path string true "Tour code"
// @Success     200 {object} models.TourProgramPublicPeriodsResponse
// @Failure     404 {object} map[string]any "NOT_FOUND"
// @Router      /tour-programs/public/{code}/periods [get]
func (h *TourProgramAPIHandler) TourProgramPublicPeriods(c *fiber.Ctx) error {
	request := &models.TourProgramPublicPeriodsRequest{Code: c.Params("code")}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, _ *coremodels.RequestInfo) (any, error) {
		return h.tourProgramUseCase.PublicPeriods(ctx, request)
	})
}
