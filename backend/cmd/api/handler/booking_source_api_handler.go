package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"backend/pkg/constants/apicode"
	"backend/pkg/core/api"
	"backend/pkg/core/api/middleware"
	coremodels "backend/pkg/core/api/models"
	"backend/pkg/modules/bookingsource"
	"backend/pkg/modules/bookingsource/models"
)

type BookingSourceAPIHandler struct {
	router               fiber.Router
	apiHandler           *api.ApiHandler
	bookingSourceUseCase bookingsource.BookingSourceUseCase
}

func NewBookingSourceAPIHandler(router fiber.Router, apiHandler *api.ApiHandler, bookingSourceUseCase bookingsource.BookingSourceUseCase) *BookingSourceAPIHandler {
	return &BookingSourceAPIHandler{router: router, apiHandler: apiHandler, bookingSourceUseCase: bookingSourceUseCase}
}

func (h *BookingSourceAPIHandler) Init(jwtAuth middleware.Auth, authPermissions middleware.AuthPermissions) {
	r := h.router

	// --- Staff (backoffice) — fixed paths before "/:id". ---
	r.Post("/booking-sources", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.BookingSourceCreate), h.BookingSourceCreate)
	r.Put("/booking-sources", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.BookingSourceUpdate), h.BookingSourceUpdate)
	r.Post("/booking-sources/list", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.BookingSourceList), h.BookingSourceList)
	r.Put("/booking-sources/status", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.BookingSourceStatus), h.BookingSourceUpdateStatus)
	r.Get("/booking-sources/:id", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.BookingSourceGet), h.BookingSourceGet)
}

// BookingSourceList godoc
// @Summary  List booking sources
// @Tags     booking-sources
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body models.BookingSourceListRequest true "Filter"
// @Success  200 {object} models.BookingSourceListResponse
// @Router   /booking-sources/list [post]
func (h *BookingSourceAPIHandler) BookingSourceList(c *fiber.Ctx) error {
	request := &models.BookingSourceListRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.bookingSourceUseCase.List(ctx, request)
	})
}

// BookingSourceGet godoc
// @Summary  Get a booking source
// @Tags     booking-sources
// @Produce  json
// @Security BearerAuth
// @Param    id path string true "Booking source ID"
// @Success  200 {object} models.BookingSourceGetResponse
// @Failure  404 {object} map[string]any "NOT_FOUND"
// @Router   /booking-sources/{id} [get]
func (h *BookingSourceAPIHandler) BookingSourceGet(c *fiber.Ctx) error {
	request := &models.BookingSourceGetRequest{ID: c.Params("id")}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.bookingSourceUseCase.Get(ctx, request)
	})
}

// BookingSourceCreate godoc
// @Summary     Create a booking source
// @Description Names are unique, case-insensitive.
// @Tags        booking-sources
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.BookingSourceCreateRequest true "Booking source"
// @Success     200 {object} models.BookingSourceCreateResponse
// @Failure     400 {object} map[string]any "VALIDATION"
// @Failure     409 {object} map[string]any "NAME_TAKEN"
// @Router      /booking-sources [post]
func (h *BookingSourceAPIHandler) BookingSourceCreate(c *fiber.Ctx) error {
	request := &models.BookingSourceCreateRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.bookingSourceUseCase.Create(ctx, request)
	})
}

// BookingSourceUpdate godoc
// @Summary     Update a booking source
// @Description Replaces every editable field.
// @Tags        booking-sources
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.BookingSourceUpdateRequest true "Booking source"
// @Success     200 {object} models.BookingSourceUpdateResponse
// @Failure     400 {object} map[string]any "VALIDATION"
// @Failure     404 {object} map[string]any "NOT_FOUND"
// @Failure     409 {object} map[string]any "NAME_TAKEN"
// @Router      /booking-sources [put]
func (h *BookingSourceAPIHandler) BookingSourceUpdate(c *fiber.Ctx) error {
	request := &models.BookingSourceUpdateRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.bookingSourceUseCase.Update(ctx, request)
	})
}

// BookingSourceUpdateStatus godoc
// @Summary  Soft-delete or restore a booking source
// @Tags     booking-sources
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body models.BookingSourceUpdateStatusRequest true "Status"
// @Success  200 {object} map[string]any "data: null"
// @Failure  404 {object} map[string]any "NOT_FOUND"
// @Router   /booking-sources/status [put]
func (h *BookingSourceAPIHandler) BookingSourceUpdateStatus(c *fiber.Ctx) error {
	request := &models.BookingSourceUpdateStatusRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.bookingSourceUseCase.UpdateStatus(ctx, request)
	})
}
