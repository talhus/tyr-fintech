package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/iamtbay/tyr-fintech/internal/dto"
	"github.com/iamtbay/tyr-fintech/internal/models"
	"github.com/iamtbay/tyr-fintech/internal/repos"
)

type ChargeService interface {
	ProcessCharge(ctx context.Context, merchant *models.Merchant, idempotencyKey string, req *dto.ChargeRequest) (*dto.ChargeResponse, int, error)
}

type chargeService struct {
	chargeRepo repos.ChargeRepository
}

func NewChargeService(chargeRepo repos.ChargeRepository) ChargeService {
	return &chargeService{chargeRepo: chargeRepo}
}

func (s *chargeService) ProcessCharge(ctx context.Context, merchant *models.Merchant, idempotencyKey string, req *dto.ChargeRequest) (*dto.ChargeResponse, int, error) {
	// 1. Idempotency Check: Return cached response if request was previously processed
	var requestHash string
	if idempotencyKey != "" {
		reqBytes, _ := json.Marshal(req)
		hasher := sha256.New()
		hasher.Write(reqBytes)
		requestHash = hex.EncodeToString(hasher.Sum(nil))

		cachedRecord, err := s.chargeRepo.GetIdempotency(ctx, merchant.ID, idempotencyKey)
		if err == nil && cachedRecord != nil {
			var cachedResp dto.ChargeResponse
			if err := json.Unmarshal([]byte(cachedRecord.ResponseBody), &cachedResp); err == nil {
				return &cachedResp, cachedRecord.ResponseStatus, nil
			}
		}
	}

	// 2. Validate Card (Luhn + Expiry + Deterministic Test Card Simulator)
	cardValidation := ValidateAndSimulateCard(req.CardNumber, req.ExpireMonth, req.ExpireYear, req.CVV)

	cleanCardNumber := strings.ReplaceAll(req.CardNumber, " ", "")
	cardLast4 := cleanCardNumber[len(cleanCardNumber)-4:]

	// 3. Handle Card Decline / Failure
	if !cardValidation.IsValid {
		failedResp := &dto.ChargeResponse{
			Success:       false,
			TransactionID: "", // Wire contract requires empty string for declines
			Status:        "FAILED",
			FailureReason: cardValidation.FailureReason,
		}

		statusCode := http.StatusBadRequest
		if cardValidation.FailureReason == "Insufficient funds" {
			statusCode = http.StatusPaymentRequired // 402
		}

		// Log failed charge for merchant audit trail
		failedCharge := &models.Charge{
			ID:             fmt.Sprintf("trx_%s", strings.ReplaceAll(uuid.New().String(), "-", "")[:10]),
			MerchantID:     merchant.ID,
			Amount:         req.Amount,
			Currency:       req.Currency,
			FeeAmount:      0,
			NetAmount:      0,
			Status:         models.ChargeStatusFailed,
			FailureReason:  cardValidation.FailureReason,
			CardLast4:      cardLast4,
			CardHolderName: req.CardHolderName,
		}
		_ = s.chargeRepo.SaveFailedCharge(ctx, failedCharge)

		// Cache idempotent failed response
		if idempotencyKey != "" {
			respBytes, _ := json.Marshal(failedResp)
			_ = s.chargeRepo.SaveIdempotency(ctx, &models.ChargeIdempotencyRecord{
				IdempotencyKey: idempotencyKey,
				MerchantID:     merchant.ID,
				RequestHash:    requestHash,
				ResponseStatus: statusCode,
				ResponseBody:   string(respBytes),
			})
		}

		return failedResp, statusCode, nil
	}

	// 4. Handle Succeeded Payment & Double-Entry Ledger Calculation
	transactionID := fmt.Sprintf("trx_%s", strings.ReplaceAll(uuid.New().String(), "-", "")[:10])

	// Calculate Gateway Fee & Net Merchant Amount strictly in minor units (int64)
	feeBasisPoints := int64(merchant.FeePercentageBasisPoints)
	feeAmount := (req.Amount * feeBasisPoints) / 10000 // e.g. 10000 * 200 / 10000 = 200 (2.00 TRY)
	netAmount := req.Amount - feeAmount

	// Retrieve System Clearing & Fee Account IDs
	acquiringAccountID, err := s.chargeRepo.GetSystemAccountID(ctx, models.AccountTypeAcquiringClearing)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("system clearing account missing: %w", err)
	}

	platformFeeAccountID, err := s.chargeRepo.GetSystemAccountID(ctx, models.AccountTypePlatformFeeRevenue)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("system fee revenue account missing: %w", err)
	}

	// Construct Balanced Double-Entry Ledger Records
	ledgerEntries := []*models.LedgerEntry{
		// Leg 1: DEBIT External Acquiring Account (+ Gross Amount)
		{
			ChargeID:    transactionID,
			AccountType: models.AccountTypeAcquiringClearing,
			AccountID:   acquiringAccountID,
			EntryType:   models.EntryTypeDebit,
			Amount:      req.Amount,
			Currency:    req.Currency,
		},
		// Leg 2: CREDIT Merchant Payable / Settlement Wallet (+ Net Amount)
		{
			ChargeID:    transactionID,
			AccountType: models.AccountTypeMerchantPayable,
			AccountID:   merchant.SettlementWalletID,
			EntryType:   models.EntryTypeCredit,
			Amount:      netAmount,
			Currency:    req.Currency,
		},
	}

	// Leg 3: CREDIT Platform Fee Account (+ Fee Amount, if > 0)
	if feeAmount > 0 {
		ledgerEntries = append(ledgerEntries, &models.LedgerEntry{
			ChargeID:    transactionID,
			AccountType: models.AccountTypePlatformFeeRevenue,
			AccountID:   platformFeeAccountID,
			EntryType:   models.EntryTypeCredit,
			Amount:      feeAmount,
			Currency:    req.Currency,
		})
	}

	charge := &models.Charge{
		ID:             transactionID,
		MerchantID:     merchant.ID,
		Amount:         req.Amount,
		Currency:       req.Currency,
		FeeAmount:      feeAmount,
		NetAmount:      netAmount,
		Status:         models.ChargeStatusSucceeded,
		FailureReason:  "",
		CardLast4:      cardLast4,
		CardHolderName: req.CardHolderName,
	}

	// 5. Execute Atomic Settlement
	if err := s.chargeRepo.ProcessSettlement(ctx, charge, ledgerEntries, merchant.SettlementWalletID); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("ledger settlement failed: %w", err)
	}

	successResp := &dto.ChargeResponse{
		Success:       true,
		TransactionID: transactionID,
		Status:        "SUCCEEDED",
		FailureReason: "",
	}

	// 6. Cache Idempotent Succeeded Response
	if idempotencyKey != "" {
		respBytes, _ := json.Marshal(successResp)
		_ = s.chargeRepo.SaveIdempotency(ctx, &models.ChargeIdempotencyRecord{
			IdempotencyKey: idempotencyKey,
			MerchantID:     merchant.ID,
			RequestHash:    requestHash,
			ResponseStatus: http.StatusOK,
			ResponseBody:   string(respBytes),
		})
	}

	return successResp, http.StatusOK, nil
}
