package services_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/iamtbay/tyr-fintech/internal/dto"
	"github.com/iamtbay/tyr-fintech/internal/models"
	"github.com/iamtbay/tyr-fintech/internal/services"
)

// mockChargeRepository implements repos.ChargeRepository in-memory for testing
type mockChargeRepository struct {
	idempotencyStore map[string]*models.ChargeIdempotencyRecord
	settledCharges   []*models.Charge
	failedCharges    []*models.Charge
	settledEntries   [][]*models.LedgerEntry
}

func newMockChargeRepository() *mockChargeRepository {
	return &mockChargeRepository{
		idempotencyStore: make(map[string]*models.ChargeIdempotencyRecord),
		settledCharges:   make([]*models.Charge, 0),
		failedCharges:    make([]*models.Charge, 0),
		settledEntries:   make([][]*models.LedgerEntry, 0),
	}
}

func (m *mockChargeRepository) GetIdempotency(ctx context.Context, merchantID, key string) (*models.ChargeIdempotencyRecord, error) {
	compositeKey := merchantID + ":" + key
	return m.idempotencyStore[compositeKey], nil
}

func (m *mockChargeRepository) SaveIdempotency(ctx context.Context, record *models.ChargeIdempotencyRecord) error {
	compositeKey := record.MerchantID + ":" + record.IdempotencyKey
	m.idempotencyStore[compositeKey] = record
	return nil
}

func (m *mockChargeRepository) SaveFailedCharge(ctx context.Context, charge *models.Charge) error {
	m.failedCharges = append(m.failedCharges, charge)
	return nil
}

func (m *mockChargeRepository) GetSystemAccountID(ctx context.Context, accountType models.AccountType) (string, error) {
	switch accountType {
	case models.AccountTypeAcquiringClearing:
		return "sys-acquiring-clearing-id", nil
	case models.AccountTypePlatformFeeRevenue:
		return "sys-platform-fee-id", nil
	default:
		return "sys-unknown-id", nil
	}
}

func (m *mockChargeRepository) ProcessSettlement(ctx context.Context, charge *models.Charge, entries []*models.LedgerEntry, settlementWalletID string) error {
	m.settledCharges = append(m.settledCharges, charge)
	m.settledEntries = append(m.settledEntries, entries)
	return nil
}

func TestChargeService_ProcessCharge_Success(t *testing.T) {
	mockRepo := newMockChargeRepository()
	chargeService := services.NewChargeService(mockRepo)

	merchant := &models.Merchant{
		ID:                       "merchant-foodeli-uuid",
		Name:                     "Foodeli Inc.",
		Email:                    "merchant@foodeli.com",
		SettlementWalletID:       "wallet-foodeli-try",
		FeePercentageBasisPoints: 200, // 2.00%
		Status:                   models.MerchantStatusActive,
	}

	req := &dto.ChargeRequest{
		Amount:         10000, // 100.00 TRY
		Currency:       "TRY",
		CardNumber:     "4111111111110000", // Guaranteed success simulator card
		CardHolderName: "John Doe",
		ExpireMonth:    "12",
		ExpireYear:     "28",
		CVV:            "123",
	}

	ctx := context.Background()
	resp, statusCode, err := chargeService.ProcessCharge(ctx, merchant, "idemp-key-1", req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if statusCode != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", statusCode)
	}
	if !resp.Success {
		t.Fatalf("expected success true, got false (reason: %s)", resp.FailureReason)
	}
	if resp.Status != "SUCCEEDED" {
		t.Fatalf("expected status SUCCEEDED, got %s", resp.Status)
	}
	if resp.TransactionID == "" {
		t.Fatal("expected non-empty transaction_id")
	}

	// Verify Double-Entry Accounting Invariant
	if len(mockRepo.settledCharges) != 1 {
		t.Fatalf("expected 1 settled charge, got %d", len(mockRepo.settledCharges))
	}
	charge := mockRepo.settledCharges[0]
	if charge.Amount != 10000 {
		t.Fatalf("expected charge amount 10000, got %d", charge.Amount)
	}
	if charge.FeeAmount != 200 { // 2% of 10000 = 200
		t.Fatalf("expected fee amount 200, got %d", charge.FeeAmount)
	}
	if charge.NetAmount != 9800 { // 10000 - 200 = 9800
		t.Fatalf("expected net amount 9800, got %d", charge.NetAmount)
	}

	// Verify balanced ledger legs: sum(debit) == sum(credit)
	entries := mockRepo.settledEntries[0]
	if len(entries) != 3 {
		t.Fatalf("expected 3 ledger entries (acquiring, merchant, fee), got %d", len(entries))
	}

	var totalDebit, totalCredit int64
	for _, entry := range entries {
		if entry.EntryType == models.EntryTypeDebit {
			totalDebit += entry.Amount
		} else if entry.EntryType == models.EntryTypeCredit {
			totalCredit += entry.Amount
		}
	}

	if totalDebit != totalCredit {
		t.Fatalf("ledger unbalanced! totalDebit=%d, totalCredit=%d", totalDebit, totalCredit)
	}
	if totalDebit != 10000 {
		t.Fatalf("expected total balanced ledger volume 10000, got %d", totalDebit)
	}
}

func TestChargeService_ProcessCharge_InsufficientFunds_Decline(t *testing.T) {
	mockRepo := newMockChargeRepository()
	chargeService := services.NewChargeService(mockRepo)

	merchant := &models.Merchant{
		ID:                       "merchant-foodeli-uuid",
		Name:                     "Foodeli Inc.",
		SettlementWalletID:       "wallet-foodeli-try",
		FeePercentageBasisPoints: 200,
		Status:                   models.MerchantStatusActive,
	}

	req := &dto.ChargeRequest{
		Amount:         10000,
		Currency:       "TRY",
		CardNumber:     "4111111111115001", // Deterministic Insufficient Funds Card
		CardHolderName: "John Doe",
		ExpireMonth:    "12",
		ExpireYear:     "28",
		CVV:            "123",
	}

	ctx := context.Background()
	resp, statusCode, err := chargeService.ProcessCharge(ctx, merchant, "", req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if statusCode != http.StatusPaymentRequired { // 402
		t.Fatalf("expected status 402 Payment Required, got %d", statusCode)
	}
	if resp.Success != false {
		t.Fatal("expected success false")
	}
	if resp.Status != "FAILED" {
		t.Fatalf("expected status FAILED, got %s", resp.Status)
	}
	if resp.FailureReason != "Insufficient funds" {
		t.Fatalf("expected 'Insufficient funds', got %s", resp.FailureReason)
	}
	if resp.TransactionID != "" {
		t.Fatalf("expected empty transaction_id on decline, got %s", resp.TransactionID)
	}

	// Verify no ledger entries written
	if len(mockRepo.settledCharges) != 0 {
		t.Fatal("settled charges should be 0 on decline")
	}
	if len(mockRepo.failedCharges) != 1 {
		t.Fatalf("expected 1 audit failed charge logged, got %d", len(mockRepo.failedCharges))
	}
}

func TestChargeService_ProcessCharge_IdempotencyReplay(t *testing.T) {
	mockRepo := newMockChargeRepository()
	chargeService := services.NewChargeService(mockRepo)

	merchant := &models.Merchant{
		ID:                       "merchant-foodeli-uuid",
		Name:                     "Foodeli Inc.",
		SettlementWalletID:       "wallet-foodeli-try",
		FeePercentageBasisPoints: 200,
		Status:                   models.MerchantStatusActive,
	}

	req := &dto.ChargeRequest{
		Amount:         10000,
		Currency:       "TRY",
		CardNumber:     "4111111111110000",
		CardHolderName: "John Doe",
		ExpireMonth:    "12",
		ExpireYear:     "28",
		CVV:            "123",
	}

	ctx := context.Background()
	idempKey := "unique-idemp-uuid-1234"

	// First Request: Should process & settle
	resp1, status1, err1 := chargeService.ProcessCharge(ctx, merchant, idempKey, req)
	if err1 != nil || status1 != http.StatusOK {
		t.Fatalf("first charge failed: err=%v, status=%d", err1, status1)
	}

	// Second Request with SAME Idempotency-Key: Should return cached response without re-settlement
	resp2, status2, err2 := chargeService.ProcessCharge(ctx, merchant, idempKey, req)
	if err2 != nil || status2 != http.StatusOK {
		t.Fatalf("second charge failed: err=%v, status=%d", err2, status2)
	}

	if resp1.TransactionID != resp2.TransactionID {
		t.Fatalf("expected identical transaction_id on replay, got %s vs %s", resp1.TransactionID, resp2.TransactionID)
	}

	// Verify settled charges was only executed ONCE in DB
	if len(mockRepo.settledCharges) != 1 {
		t.Fatalf("expected exactly 1 settled charge execution despite retry, got %d", len(mockRepo.settledCharges))
	}
}
