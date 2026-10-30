package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"backend/pkg/constants/apicode"
	"backend/pkg/core/api"
	"backend/pkg/core/api/middleware"
	coremodels "backend/pkg/core/api/models"
	"backend/pkg/modules/country"
	"backend/pkg/modules/country/models"
)

type CountryAPIHandler struct {
	router         fiber.Router
	apiHandler     *api.ApiHandler
	countryUseCase country.CountryUseCase
}

func NewCountryAPIHandler(router fiber.Router, apiHandler *api.ApiHandler, countryUseCase country.CountryUseCase) *CountryAPIHandler {
	return &CountryAPIHandler{router: router, apiHandler: apiHandler, countryUseCase: countryUseCase}
}

func (h *CountryAPIHandler) Init(jwtAuth middleware.Auth, authPermissions middleware.AuthPermissions) {
	r := h.router

	// --- Staff (backoffice) — fixed paths before "/:id". ---
	r.Post("/countries", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.CountryCreate), h.CountryCreate)
	r.Put("/countries", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.CountryUpdate), h.CountryUpdate)
	r.Post("/countries/list", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.CountryList), h.CountryList)
	r.Put("/countries/status", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.CountryStatus), h.CountryUpdateStatus)
	r.Get("/countries/:id", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.CountryGet), h.CountryGet)
}

// CountryList godoc
// @Summary  List countries
// @Tags     countries
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body models.CountryListRequest true "Filter"
// @Success  200 {object} models.CountryListResponse
// @Router   /countries/list [post]
func (h *CountryAPIHandler) CountryList(c *fiber.Ctx) error {
	request := &models.CountryListRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.countryUseCase.List(ctx, request)
	})
}

// CountryGet godoc
// @Summary  Get a country
// @Tags     countries
// @Produce  json
// @Security BearerAuth
// @Param    id path string true "Country ID"
// @Success  200 {object} models.CountryGetResponse
// @Failure  404 {object} map[string]any "NOT_FOUND"
// @Router   /countries/{id} [get]
func (h *CountryAPIHandler) CountryGet(c *fiber.Ctx) error {
	request := &models.CountryGetRequest{ID: c.Params("id")}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.countryUseCase.Get(ctx, request)
	})
}

// CountryCreate godoc
// @Summary     Create a country
// @Description The slug is locked after create.
// @Tags        countries
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.CountryCreateRequest true "Country"
// @Success     200 {object} models.CountryCreateResponse
// @Failure     400 {object} map[string]any "VALIDATION"
// @Failure     409 {object} map[string]any "SLUG_TAKEN, REF_ID_TAKEN"
// @Router      /countries [post]
func (h *CountryAPIHandler) CountryCreate(c *fiber.Ctx) error {
	request := &models.CountryCreateRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.countryUseCase.Create(ctx, request)
	})
}

// CountryUpdate godoc
// @Summary  Update a country's display fields and ref ID (not the slug)
// @Tags     countries
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body models.CountryUpdateRequest true "Country"
// @Success  200 {object} models.CountryUpdateResponse
// @Failure  400 {object} map[string]any "VALIDATION"
// @Failure  404 {object} map[string]any "NOT_FOUND"
// @Failure  409 {object} map[string]any "REF_ID_TAKEN"
// @Router   /countries [put]
func (h *CountryAPIHandler) CountryUpdate(c *fiber.Ctx) error {
	request := &models.CountryUpdateRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.countryUseCase.Update(ctx, request)
	})
}

// CountryUpdateStatus godoc
// @Summary  Soft-delete or restore a country
// @Tags     countries
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body models.CountryUpdateStatusRequest true "Status"
// @Success  200 {object} map[string]any "data: null"
// @Failure  404 {object} map[string]any "NOT_FOUND"
// @Router   /countries/status [put]
func (h *CountryAPIHandler) CountryUpdateStatus(c *fiber.Ctx) error {
	request := &models.CountryUpdateStatusRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.countryUseCase.UpdateStatus(ctx, request)
	})
}
