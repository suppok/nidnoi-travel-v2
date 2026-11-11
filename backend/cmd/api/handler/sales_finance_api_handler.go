package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"backend/pkg/constants/apicode"
	"backend/pkg/core/api"
	"backend/pkg/core/api/middleware"
	coremodels "backend/pkg/core/api/models"
	"backend/pkg/modules/salesfinance"
	"backend/pkg/modules/salesfinance/models"
)

type SalesFinanceAPIHandler struct {
	router              fiber.Router
	apiHandler          *api.ApiHandler
	salesFinanceUseCase salesfinance.SalesFinanceUseCase
}

func NewSalesFinanceAPIHandler(router fiber.Router, apiHandler *api.ApiHandler, salesFinanceUseCase salesfinance.SalesFinanceUseCase) *SalesFinanceAPIHandler {
	return &SalesFinanceAPIHandler{router: router, apiHandler: apiHandler, salesFinanceUseCase: salesFinanceUseCase}
}

func (h *SalesFinanceAPIHandler) Init(jwtAuth middleware.Auth, authPermissions middleware.AuthPermissions) {
	r := h.router

	// --- Staff (backoffice) — always the caller's own settings. ---
	r.Get("/sales-finances", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.SalesFinanceGet), h.SalesFinanceGet)
	r.Put("/sales-finances", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.SalesFinanceUpdate), h.SalesFinanceUpdate)
}

// SalesFinanceGet godoc
// @Summary     Your finance users for confirmed LINE bookings
// @Description The finance users you can send a confirmed LINE booking to, your default, and every
// @Description active finance user as options.
// @Tags        sales-finances
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} models.SalesFinanceGetResponse
// @Router      /sales-finances [get]
func (h *SalesFinanceAPIHandler) SalesFinanceGet(c *fiber.Ctx) error {
	request := &models.SalesFinanceGetRequest{}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.salesFinanceUseCase.Get(ctx, request)
	})
}

// SalesFinanceUpdate godoc
// @Summary     Choose your finance users for confirmed LINE bookings
// @Description Replaces your list. defaultFinanceUserId must be in financeUserIds ("" = the only one
// @Description when there is one).
// @Tags        sales-finances
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.SalesFinanceUpdateRequest true "Finance users"
// @Success     200 {object} models.SalesFinanceUpdateResponse
// @Failure     400 {object} map[string]any "VALIDATION"
// @Router      /sales-finances [put]
func (h *SalesFinanceAPIHandler) SalesFinanceUpdate(c *fiber.Ctx) error {
	request := &models.SalesFinanceUpdateRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.salesFinanceUseCase.Update(ctx, request)
	})
}
