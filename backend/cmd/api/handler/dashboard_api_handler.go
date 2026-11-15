package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"backend/pkg/core/api"
	"backend/pkg/core/api/middleware"
	coremodels "backend/pkg/core/api/models"
	"backend/pkg/modules/dashboard"
	"backend/pkg/modules/dashboard/models"
)

type DashboardAPIHandler struct {
	router           fiber.Router
	apiHandler       *api.ApiHandler
	dashboardUseCase dashboard.DashboardUseCase
}

func NewDashboardAPIHandler(router fiber.Router, apiHandler *api.ApiHandler, dashboardUseCase dashboard.DashboardUseCase) *DashboardAPIHandler {
	return &DashboardAPIHandler{router: router, apiHandler: apiHandler, dashboardUseCase: dashboardUseCase}
}

// Init registers the dashboard route. It needs a session only: the use case leaves out each
// section the caller lacks the list permission for.
func (h *DashboardAPIHandler) Init(jwtAuth middleware.Auth) {
	h.router.Post("/dashboard/summary", jwtAuth.Auth, h.DashboardSummary)
}

// DashboardSummary godoc
// @Summary     Backoffice dashboard figures
// @Description Totals, chart buckets and breakdowns for the last `days` days (Thailand time), with the
// @Description previous period of the same length for deltas. Sections: tours (tour-program:list),
// @Description customer booking requests (booking:list) and LINE bookings (line-booking:list, limited
// @Description to the caller's own like the list). A section the caller may not see is null.
// @Tags        dashboard
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.DashboardSummaryRequest true "Period"
// @Success     200 {object} models.DashboardSummaryResponse
// @Failure     400 {object} map[string]any "VALIDATION"
// @Router      /dashboard/summary [post]
func (h *DashboardAPIHandler) DashboardSummary(c *fiber.Ctx) error {
	request := &models.DashboardSummaryRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.dashboardUseCase.Summary(ctx, request)
	})
}
