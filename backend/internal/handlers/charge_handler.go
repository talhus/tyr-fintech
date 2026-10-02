package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iamtbay/tyr-fintech/internal/dto"
	"github.com/iamtbay/tyr-fintech/internal/middleware"
	"github.com/iamtbay/tyr-fintech/internal/services"
)

type ChargeHandler struct {
	chargeService services.ChargeService
}

func NewChargeHandler(chargeService services.ChargeService) *ChargeHandler {
	return &ChargeHandler{chargeService: chargeService}
}

// ProcessCharge handles POST /api/v1/charges
func (h *ChargeHandler) ProcessCharge(c *gin.Context) {
	merchant, ok := middleware.GetAuthenticatedMerchant(c)
	if !ok || merchant == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: merchant context missing"})
		return
	}

	var req dto.ChargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ChargeResponse{
			Success:       false,
			TransactionID: "",
			Status:        "FAILED",
			FailureReason: "Invalid request payload: " + err.Error(),
		})
		return
	}

	idempotencyKey := c.GetHeader("Idempotency-Key")

	resp, statusCode, err := h.chargeService.ProcessCharge(c.Request.Context(), merchant, idempotencyKey, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ChargeResponse{
			Success:       false,
			TransactionID: "",
			Status:        "FAILED",
			FailureReason: "Internal gateway processing error",
		})
		return
	}

	c.JSON(statusCode, resp)
}
