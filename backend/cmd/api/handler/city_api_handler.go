package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"backend/pkg/constants/apicode"
	"backend/pkg/core/api"
	"backend/pkg/core/api/middleware"
	coremodels "backend/pkg/core/api/models"
	"backend/pkg/modules/city"
	"backend/pkg/modules/city/models"
)

type CityAPIHandler struct {
	router      fiber.Router
	apiHandler  *api.ApiHandler
	cityUseCase city.CityUseCase
}

func NewCityAPIHandler(router fiber.Router, apiHandler *api.ApiHandler, cityUseCase city.CityUseCase) *CityAPIHandler {
	return &CityAPIHandler{router: router, apiHandler: apiHandler, cityUseCase: cityUseCase}
}

func (h *CityAPIHandler) Init(jwtAuth middleware.Auth, authPermissions middleware.AuthPermissions) {
	r := h.router

	// --- Public (customer site) — no auth. Registered BEFORE "/:id" routes. ---
	r.Get("/cities/public", h.CityPublicList)

	// --- Staff (backoffice) — fixed paths before "/:id". ---
	r.Post("/cities", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.CityCreate), h.CityCreate)
	r.Put("/cities", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.CityUpdate), h.CityUpdate)
	r.Post("/cities/list", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.CityList), h.CityList)
	r.Put("/cities/status", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.CityStatus), h.CityUpdateStatus)
	r.Get("/cities/:id", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.CityGet), h.CityGet)
}

// CityList godoc
// @Summary  List cities
// @Tags     cities
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body models.CityListRequest true "Filter"
// @Success  200 {object} models.CityListResponse
// @Router   /cities/list [post]
func (h *CityAPIHandler) CityList(c *fiber.Ctx) error {
	request := &models.CityListRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.cityUseCase.List(ctx, request)
	})
}

// CityGet godoc
// @Summary  Get a city
// @Tags     cities
// @Produce  json
// @Security BearerAuth
// @Param    id path string true "City ID"
// @Success  200 {object} models.CityGetResponse
// @Failure  404 {object} map[string]any "NOT_FOUND"
// @Router   /cities/{id} [get]
func (h *CityAPIHandler) CityGet(c *fiber.Ctx) error {
	request := &models.CityGetRequest{ID: c.Params("id")}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.cityUseCase.Get(ctx, request)
	})
}

// CityCreate godoc
// @Summary     Create a city
// @Description The slug is locked after create. The country must be active.
// @Tags        cities
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.CityCreateRequest true "City"
// @Success     200 {object} models.CityCreateResponse
// @Failure     400 {object} map[string]any "VALIDATION"
// @Failure     409 {object} map[string]any "SLUG_TAKEN, REF_ID_TAKEN"
// @Router      /cities [post]
func (h *CityAPIHandler) CityCreate(c *fiber.Ctx) error {
	request := &models.CityCreateRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.cityUseCase.Create(ctx, request)
	})
}

// CityUpdate godoc
// @Summary  Update a city's country, display fields and ref ID (not the slug)
// @Tags     cities
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body models.CityUpdateRequest true "City"
// @Success  200 {object} models.CityUpdateResponse
// @Failure  400 {object} map[string]any "VALIDATION"
// @Failure  404 {object} map[string]any "NOT_FOUND"
// @Failure  409 {object} map[string]any "REF_ID_TAKEN"
// @Router   /cities [put]
func (h *CityAPIHandler) CityUpdate(c *fiber.Ctx) error {
	request := &models.CityUpdateRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.cityUseCase.Update(ctx, request)
	})
}

// CityUpdateStatus godoc
// @Summary     Soft-delete or restore a city
// @Description Restoring requires the city's country to be active.
// @Tags        cities
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.CityUpdateStatusRequest true "Status"
// @Success     200 {object} map[string]any "data: null"
// @Failure     400 {object} map[string]any "VALIDATION (countryId: country deleted)"
// @Failure     404 {object} map[string]any "NOT_FOUND"
// @Router      /cities/status [put]
func (h *CityAPIHandler) CityUpdateStatus(c *fiber.Ctx) error {
	request := &models.CityUpdateStatusRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.cityUseCase.UpdateStatus(ctx, request)
	})
}

// CityPublicList godoc
// @Summary     List public cities with tour stats (customer site build)
// @Description Every active city in display order with its public tour count and lowest upcoming
// @Description price. The site skips landing pages whose tourCount is 0. No auth.
// @Tags        public
// @Produce     json
// @Success     200 {object} models.CityPublicListResponse
// @Router      /cities/public [get]
func (h *CityAPIHandler) CityPublicList(c *fiber.Ctx) error {
	return h.apiHandler.Do(c, nil, func(ctx context.Context, _ *coremodels.RequestInfo) (any, error) {
		return h.cityUseCase.PublicList(ctx)
	})
}
