package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"backend/pkg/constants/apicode"
	"backend/pkg/core/api"
	"backend/pkg/core/api/middleware"
	coremodels "backend/pkg/core/api/models"
	"backend/pkg/core/config"
	"backend/pkg/modules/booking"
	"backend/pkg/modules/booking/models"
)

type BookingAPIHandler struct {
	router         fiber.Router
	apiHandler     *api.ApiHandler
	config         *config.Config
	bookingUseCase booking.BookingUseCase
}

func NewBookingAPIHandler(router fiber.Router, apiHandler *api.ApiHandler, cfg *config.Config, bookingUseCase booking.BookingUseCase) *BookingAPIHandler {
	return &BookingAPIHandler{router: router, apiHandler: apiHandler, config: cfg, bookingUseCase: bookingUseCase}
}

func (h *BookingAPIHandler) Init(jwtAuth middleware.Auth, authPermissions middleware.AuthPermissions) {
	r := h.router

	// --- Public (customer site) — no auth, rate-limited per client IP. ---
	r.Post("/bookings/public", middleware.RateLimit(h.config.RateLimit.Booking, h.config.Server.ClientIPHeader), h.BookingPublicCreate)

	// --- Staff (backoffice) — fixed paths before "/:id". ---
	r.Put("/bookings", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.BookingUpdate), h.BookingUpdate)
	r.Post("/bookings/list", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.BookingList), h.BookingList)
	r.Put("/bookings/status", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.BookingStatus), h.BookingUpdateStatus)
	r.Get("/bookings/:id", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.BookingGet), h.BookingGet)
}

// BookingPublicCreate godoc
// @Summary     Send a booking request (customer site)
// @Description Records a sales lead for a departure. Reserves no seats and takes no payment: sales
// @Description contacts the customer to confirm. Prices are quoted from the period, never from the
// @Description request. Rate-limited per client IP.
// @Tags        bookings
// @Accept      json
// @Produce     json
// @Param       body body models.BookingPublicCreateRequest true "Booking request"
// @Success     200 {object} models.BookingPublicCreateResponse
// @Failure     400 {object} map[string]any "VALIDATION"
// @Failure     404 {object} map[string]any "NOT_FOUND (tour)"
// @Failure     409 {object} map[string]any "PERIOD_UNAVAILABLE"
// @Failure     429 {object} map[string]any "RATE_LIMITED"
// @Router      /bookings/public [post]
func (h *BookingAPIHandler) BookingPublicCreate(c *fiber.Ctx) error {
	request := &models.BookingPublicCreateRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, _ *coremodels.RequestInfo) (any, error) {
		return h.bookingUseCase.PublicCreate(ctx, request)
	})
}

// BookingList godoc
// @Summary  List booking requests, newest first
// @Tags     bookings
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body models.BookingListRequest true "Filter"
// @Success  200 {object} models.BookingListResponse
// @Failure  400 {object} map[string]any "VALIDATION"
// @Router   /bookings/list [post]
func (h *BookingAPIHandler) BookingList(c *fiber.Ctx) error {
	request := &models.BookingListRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.bookingUseCase.List(ctx, request)
	})
}

// BookingGet godoc
// @Summary  Get a booking request
// @Tags     bookings
// @Produce  json
// @Security BearerAuth
// @Param    id path string true "Booking ID"
// @Success  200 {object} models.BookingGetResponse
// @Failure  404 {object} map[string]any "NOT_FOUND"
// @Router   /bookings/{id} [get]
func (h *BookingAPIHandler) BookingGet(c *fiber.Ctx) error {
	request := &models.BookingGetRequest{ID: c.Params("id")}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.bookingUseCase.Get(ctx, request)
	})
}

// BookingUpdate godoc
// @Summary     Update a booking request's sales state and staff note
// @Description The first user to move the booking out of "new" becomes its assignee.
// @Tags        bookings
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.BookingUpdateRequest true "State and note"
// @Success     200 {object} models.BookingUpdateResponse
// @Failure     400 {object} map[string]any "VALIDATION"
// @Failure     404 {object} map[string]any "NOT_FOUND"
// @Router      /bookings [put]
func (h *BookingAPIHandler) BookingUpdate(c *fiber.Ctx) error {
	request := &models.BookingUpdateRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.bookingUseCase.Update(ctx, request)
	})
}

// BookingUpdateStatus godoc
// @Summary  Delete (spam, duplicates) or restore a booking request
// @Tags     bookings
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body models.BookingUpdateStatusRequest true "Status"
// @Success  200 {object} map[string]any "data: null"
// @Failure  404 {object} map[string]any "NOT_FOUND"
// @Router   /bookings/status [put]
func (h *BookingAPIHandler) BookingUpdateStatus(c *fiber.Ctx) error {
	request := &models.BookingUpdateStatusRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.bookingUseCase.UpdateStatus(ctx, request)
	})
}
