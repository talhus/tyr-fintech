package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iamtbay/tyr-fintech/internal/dto"
	"github.com/iamtbay/tyr-fintech/internal/models"
	"github.com/iamtbay/tyr-fintech/internal/services"
	"github.com/iamtbay/tyr-fintech/pkg/response"
	"github.com/iamtbay/tyr-fintech/pkg/utils"
)

type CardHandler struct {
	service *services.CardService
}

func NewCardHandler(service *services.CardService) *CardHandler {
	return &CardHandler{service: service}
}

//HANDLERS

// create card
func (h *CardHandler) CreateCard(c *gin.Context) {
	var req dto.CardCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	userID := c.GetString("userID")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if req.WalletID == "" {
		response.Error(c, http.StatusBadRequest, "Wallet ID is required")
		return
	}

	card, err := h.service.CreateCard(c.Request.Context(), userID, req.WalletID, req.LimitAmount)
	if err != nil {
		fmt.Println(err.Error())
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	resp := dto.CardResponse{
		ID:          card.ID,
		WalletID:    card.WalletID,
		CardNumber:  utils.MaskCardNumber(card.CardNumber),
		ExpiryMonth: card.ExpiryMonth,
		ExpiryYear:  card.ExpiryYear,
		LimitAmount: card.LimitAmount,
		SpentAmount: card.SpentAmount,
		Status:      card.Status,
		Currency:    card.Currency,
	}

	response.Success(c, http.StatusCreated, resp)
}

// get cards
func (h *CardHandler) GetCards(c *gin.Context) {
	userID := c.GetString("userID")
	cards, err := h.service.GetCardsByUserID(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get cards")
		return
	}
	cardList := []dto.CardResponse{}
	for _, card := range cards {
		cardList = append(cardList, dto.CardResponse{
			ID:          card.ID,
			WalletID:    card.WalletID,
			CardNumber:  utils.MaskCardNumber(card.CardNumber),
			ExpiryMonth: card.ExpiryMonth,
			ExpiryYear:  card.ExpiryYear,
			LimitAmount: card.LimitAmount,
			SpentAmount: card.SpentAmount,
			Status:      card.Status,
			Currency:    card.Currency,
		})
	}
	response.Success(c, http.StatusOK, cardList)
}

// get unmasked card details
func (h *CardHandler) GetCardDetails(c *gin.Context) {
	cardID := c.Param("cardID")
	userID := c.GetString("userID")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	card, err := h.service.GetCardDetails(c.Request.Context(), cardID, userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Card not found")
		return
	}

	resp := dto.CardDetailResponse{
		ID:          card.ID,
		WalletID:    card.WalletID,
		CardNumber:  card.CardNumber,
		CVV:         card.CVV,
		ExpiryMonth: card.ExpiryMonth,
		ExpiryYear:  card.ExpiryYear,
		LimitAmount: card.LimitAmount,
		SpentAmount: card.SpentAmount,
		Status:      card.Status,
		Currency:    card.Currency,
	}

	response.Success(c, http.StatusOK, resp)
}

// freeze card
func (h *CardHandler) FreezeCard(c *gin.Context) {
	userID := c.GetString("userID")
	cardID := c.Param("cardID")
	err := h.service.UpdateCardStatus(c.Request.Context(), cardID, userID, models.CardStatusFrozen)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to freeze card")
		return
	}
	response.Success(c, http.StatusOK, "Card frozen successfully")
}

// activate card
func (h *CardHandler) ActivateCard(c *gin.Context) {
	userID := c.GetString("userID")
	cardID := c.Param("cardID")
	err := h.service.UpdateCardStatus(c.Request.Context(), cardID, userID, models.CardStatusActive)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to activate card")
		return
	}
	response.Success(c, http.StatusOK, "Card activated successfully")
}

// close card
func (h *CardHandler) CloseCard(c *gin.Context) {
	userID := c.GetString("userID")
	cardID := c.Param("cardID")
	err := h.service.UpdateCardStatus(c.Request.Context(), cardID, userID, models.CardStatusClosed)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to close card")
		return
	}
	response.Success(c, http.StatusOK, "Card closed successfully")
}

// get card transactions / spendings
func (h *CardHandler) GetCardTransactions(c *gin.Context) {
	cardID := c.Param("cardID")
	userID := c.GetString("userID")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	txs, err := h.service.GetCardTransactions(c.Request.Context(), cardID, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get card transactions")
		return
	}
	response.Success(c, http.StatusOK, txs)
}

// process payment
func (h *CardHandler) ProcessPayment(c *gin.Context) {
	cardID := c.Param("cardID")
	var req dto.CardPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println(err.Error())
		response.Error(c, http.StatusBadRequest, "Invalid request payload")
		return
	}
	txID, err := h.service.ProcessPayment(c.Request.Context(), cardID, req.CVV, req.ExpiryMonth, req.ExpiryYear, req.Amount, req.MerchantName)
	if err != nil {
		log.Println("2", err.Error())

		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusOK, dto.CardPaymentResponse{
		TransactionID: txID,
		AmountCharged: req.Amount,
		Message:       "Payment processed successfully",
		Status:        dto.CardPaymentStatusApproved,
	})
}
