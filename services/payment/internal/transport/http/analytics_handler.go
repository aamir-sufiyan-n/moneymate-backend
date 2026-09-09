package http

import (
	"github.com/gofiber/fiber/v3"

	"github.com/moneymate-2026/moneymate-backend/services/payment/internal/usecases"
	response "github.com/moneymate-2026/moneymate-backend/shared/pkg/responses"
)

type AnalyticsHandler struct {
	analytics usecases.AnalyticsUsecase
}

func NewAnalyticsHandler(analytics usecases.AnalyticsUsecase) *AnalyticsHandler {
	return &AnalyticsHandler{analytics: analytics}
}

func (h *AnalyticsHandler) SpendByCategory(c fiber.Ctx) error {
	userID := userIDFromLocals(c)
	if userID == "" {
		return response.Unauthorized(c, "authentication required")
	}

	from := c.Query("from")
	to := c.Query("to")

	res, err := h.analytics.SpendByCategory(c.Context(), userID, from, to)
	if err != nil {
		return handleError(c, err)
	}

	return response.OK(c, "spend by category fetched", res)
}

func (h *AnalyticsHandler) SpendByPeriod(c fiber.Ctx) error {
	userID := userIDFromLocals(c)
	if userID == "" {
		return response.Unauthorized(c, "authentication required")
	}

	granularity := c.Query("granularity")
	from := c.Query("from")
	to := c.Query("to")

	res, err := h.analytics.SpendByPeriod(c.Context(), userID, granularity, from, to)
	if err != nil {
		return handleError(c, err)
	}

	return response.OK(c, "spend by period fetched", res)
}
