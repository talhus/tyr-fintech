package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iamtbay/tyr-fintech/internal/dto"
	"github.com/iamtbay/tyr-fintech/internal/middleware"
	"github.com/iamtbay/tyr-fintech/internal/services"
)

type CheckoutHandler struct {
	checkoutService services.CheckoutService
}

func NewCheckoutHandler(checkoutService services.CheckoutService) *CheckoutHandler {
	return &CheckoutHandler{checkoutService: checkoutService}
}

// CreateSession handles POST /api/v1/checkout/sessions (Protected by Merchant API Key)
func (h *CheckoutHandler) CreateSession(c *gin.Context) {
	merchant, ok := middleware.GetAuthenticatedMerchant(c)
	if !ok || merchant == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: merchant context missing"})
		return
	}

	var req dto.CreateCheckoutSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	resp, err := h.checkoutService.CreateSession(c.Request.Context(), merchant, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create checkout session: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetSessionDetails handles GET /api/v1/checkout/sessions/:session_id (Public endpoint for checkout UI)
func (h *CheckoutHandler) GetSessionDetails(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id path parameter is required"})
		return
	}

	resp, err := h.checkoutService.GetSessionDetails(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// PaySession handles POST /api/v1/checkout/sessions/:session_id/pay (Protected by User JWT)
func (h *CheckoutHandler) PaySession(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: user authentication required"})
		return
	}

	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id path parameter is required"})
		return
	}

	var req dto.PayCheckoutSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	resp, err := h.checkoutService.PaySession(c.Request.Context(), sessionID, userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
