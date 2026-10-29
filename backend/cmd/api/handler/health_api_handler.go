package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"backend/pkg/core/api"
	coremodels "backend/pkg/core/api/models"
)

type HealthAPIHandler struct {
	router     fiber.Router
	apiHandler *api.ApiHandler
}

type HealthResponse struct {
	Status string `json:"status"`
}

func NewHealthAPIHandler(router fiber.Router, apiHandler *api.ApiHandler) *HealthAPIHandler {
	return &HealthAPIHandler{router: router, apiHandler: apiHandler}
}

func (h *HealthAPIHandler) Init() {
	h.router.Get("/health", h.Health)
}

// Health godoc
// @Summary Liveness check
// @Tags    health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router  /health [get]
func (h *HealthAPIHandler) Health(c *fiber.Ctx) error {
	return h.apiHandler.Do(c, nil, func(context.Context, *coremodels.RequestInfo) (any, error) {
		return &HealthResponse{Status: "ok"}, nil
	})
}
