package repos

import (
	"context"
	"net/http"

	"github.com/iamtbay/tyr-fintech/internal/models"
	"github.com/iamtbay/tyr-fintech/pkg/apperrors"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CardRepository struct {
	db *pgxpool.Pool
}

func NewCardRepository(db *pgxpool.Pool) *CardRepository {
	return &CardRepository{db: db}
}

// CREATE
func (r *CardRepository) Create(ctx context.Context, card *models.Card) error {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM cards WHERE wallet_id = $1 AND status != 'CLOSED'`, card.WalletID).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return apperrors.New(http.StatusBadRequest, "A virtual card already exists for this wallet")
	}

	query := `INSERT INTO cards (id, user_id, wallet_id, card_number, cvv, expiry_month, expiry_year, limit_amount, spent_amount,status) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err = r.db.Exec(ctx, query,
		card.ID, card.UserID, card.WalletID, card.CardNumber, card.CVV, card.ExpiryMonth, card.ExpiryYear, card.LimitAmount, card.SpentAmount, card.Status)
	if err != nil {
		return err
	}
	_ = r.db.QueryRow(ctx, `SELECT currency FROM wallets WHERE id = $1`, card.WalletID).Scan(&card.Currency)
	return nil
}

// GET BY USER ID
func (r *CardRepository) GetByUserID(ctx context.Context, userID string) ([]models.Card, error) {
	query := `SELECT c.id, c.user_id, c.wallet_id, c.card_number, c.cvv, c.expiry_month, c.expiry_year, c.limit_amount, c.spent_amount, c.status, c.created_at, w.currency 
	          FROM cards c 
	          JOIN wallets w ON c.wallet_id = w.id 
	          WHERE c.user_id = $1 AND c.status != 'CLOSED'
	          ORDER BY c.created_at ASC`
	cards := []models.Card{}
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		card := models.Card{}
		err := rows.Scan(&card.ID, &card.UserID, &card.WalletID, &card.CardNumber, &card.CVV, &card.ExpiryMonth, &card.ExpiryYear, &card.LimitAmount, &card.SpentAmount, &card.Status, &card.CreatedAt, &card.Currency)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}
	return cards, nil
}

func (r *CardRepository) UpdateStatus(ctx context.Context, cardID, userID string, status models.CardStatus) error {
	query := `UPDATE cards SET status=$1 WHERE id=$2 AND user_id=$3`
	_, err := r.db.Exec(ctx, query, status, cardID, userID)
	return err
}

// GET CARD TRANSACTIONS / SPENDINGS
func (r *CardRepository) GetCardTransactions(ctx context.Context, cardID, userID string) ([]models.Transaction, error) {
	query := `SELECT t.id, COALESCE(t.from_wallet_id, '00000000-0000-0000-0000-000000000000'), 
	                 COALESCE(t.to_wallet_id, '00000000-0000-0000-0000-000000000000'), 
	                 COALESCE(fw.wallet_number, 0), COALESCE(tw.wallet_number, 0), 
	                 t.amount, COALESCE(t.converted_amount, t.amount), t.status, t.created_at, 
	                 t.card_id, COALESCE(t.merchant_name, 'Online Purchase')
	          FROM transactions t
	          JOIN cards c ON t.card_id = c.id
	          LEFT JOIN wallets fw ON t.from_wallet_id = fw.id
	          LEFT JOIN wallets tw ON t.to_wallet_id = tw.id
	          WHERE t.card_id = $1 AND c.user_id = $2
	          ORDER BY t.created_at DESC`
	rows, err := r.db.Query(ctx, query, cardID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := []models.Transaction{}
	for rows.Next() {
		var tx models.Transaction
		var cardIDVal, merchantVal string
		err := rows.Scan(&tx.ID, &tx.FromWalletID, &tx.ToWalletID, &tx.FromWalletNumber, &tx.ToWalletNumber, &tx.Amount, &tx.ConvertedAmount, &tx.Status, &tx.CreatedAt, &cardIDVal, &merchantVal)
		if err != nil {
			return nil, err
		}
		tx.CardID = &cardIDVal
		tx.MerchantName = &merchantVal
		transactions = append(transactions, tx)
	}
	return transactions, nil
}

// GET UNMASKED CARD DETAILS
func (r *CardRepository) GetCardDetails(ctx context.Context, cardID, userID string) (*models.Card, error) {
	query := `SELECT c.id, c.user_id, c.wallet_id, c.card_number, c.cvv, c.expiry_month, c.expiry_year, c.limit_amount, c.spent_amount, c.status, c.created_at, w.currency 
	          FROM cards c 
	          JOIN wallets w ON c.wallet_id = w.id 
	          WHERE c.id = $1 AND c.user_id = $2 AND c.status != 'CLOSED'`
	card := &models.Card{}
	err := r.db.QueryRow(ctx, query, cardID, userID).Scan(&card.ID, &card.UserID, &card.WalletID, &card.CardNumber, &card.CVV, &card.ExpiryMonth, &card.ExpiryYear, &card.LimitAmount, &card.SpentAmount, &card.Status, &card.CreatedAt, &card.Currency)
	if err != nil {
		return nil, err
	}
	return card, nil
}

//PROCESS PAYMENT

func (r *CardRepository) ProcessPayment(ctx context.Context, transactionID, cardID, cvv string, expiryMonth, expiryYear int, amount int64, merchantName string) (*models.CardPaymentResult, error) {

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var card models.Card
	var userEmail, userName string
	// get card and user details from db
	query := `SELECT c.id, c.user_id, u.email, u.name, c.wallet_id, c.limit_amount, c.spent_amount, c.status, c.cvv, c.expiry_month, c.expiry_year 
	          FROM cards c 
	          JOIN users u ON c.user_id = u.id 
	          WHERE c.id = $1 FOR UPDATE`
	err = tx.QueryRow(ctx, query, cardID).Scan(&card.ID, &card.UserID, &userEmail, &userName, &card.WalletID, &card.LimitAmount, &card.SpentAmount, &card.Status, &card.CVV, &card.ExpiryMonth, &card.ExpiryYear)
	if err != nil {
		return nil, err
	}

	// check if card is active
	if card.Status != models.CardStatusActive {
		return nil, apperrors.New(http.StatusBadRequest, "Card is not active")
	}
	// check if card details are valid
	if card.CVV != cvv || card.ExpiryMonth != expiryMonth || card.ExpiryYear != expiryYear {
		return nil, apperrors.New(http.StatusBadRequest, "Invalid card details")
	}
	// check if user has enough funds
	if amount+card.SpentAmount > card.LimitAmount {
		return nil, apperrors.New(http.StatusBadRequest, "Limit Exceeded")
	}

	// get wallet balance
	var walletBalance int64
	err = tx.QueryRow(ctx, `SELECT balance FROM wallets WHERE id=$1 FOR UPDATE`, card.WalletID).Scan(&walletBalance)
	if err != nil {
		return nil, err
	}
	// check wallet has enough balance
	if walletBalance < amount {
		return nil, apperrors.New(http.StatusBadRequest, "Insufficient funds")
	}

	// update spent_amount and balance
	_, err = tx.Exec(ctx, `UPDATE wallets SET balance=balance-$1 WHERE id=$2`, amount, card.WalletID)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx, `UPDATE cards SET spent_amount=spent_amount+$1 WHERE id=$2`, amount, card.ID)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO transactions(id, from_wallet_id, to_wallet_id, amount, status, card_id, merchant_name) VALUES($1, $2, $3, $4, $5, $6, $7)`,
		transactionID, card.WalletID, nil, amount, models.StatusCompleted, card.ID, merchantName)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &models.CardPaymentResult{
		TransactionID: transactionID,
		UserID:        card.UserID,
		UserEmail:     userEmail,
		UserName:      userName,
		MerchantName:  merchantName,
		Amount:        amount,
	}, nil
}

