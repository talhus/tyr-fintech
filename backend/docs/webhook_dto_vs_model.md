# DTO vs. Database Model for Webhook Queue Design

This document details the architectural reasoning for using a **Data Transfer Object (DTO)** instead of a **Database/Domain Model** for asynchronous webhook channels and worker queues.

---

## 1. Architectural Analysis: DTO vs. Database Model

When sending event notifications (like webhooks) from your services to background workers:

### Exposing Database Models (`models.Transaction`)
* **Tight Coupling**: Database models map directly to table schemas (e.g., specific columns, foreign keys like `from_wallet_id`, and DB-specific data types). If you change the database schema in a future migration, your webhook payload structure will break for external consumers without warning.
* **Data Leakage**: Database models often contain internal metadata (e.g., raw statuses, internal identifiers, system auditing timestamps like `created_at` or `deleted_at`) that should not be exposed to third-party webhooks.
* **Serialization Limitations**: Domain models are tagged with database attributes (e.g., `db:"id"`). They may not have the optimal `json` representation (camelCase vs snake_case, string formatting of large numbers, etc.) needed by webhook clients.

### Using a Dedicated Webhook DTO (`dto.TransactionWebhookEvent`)
* **Stable API Contract**: The webhook payload acts as a public-facing API contract. By defining a dedicated DTO, you guarantee that database schema changes will never alter or break the payload structure sent to consumers.
* **Contextual Data & Enrichment**: You can easily customize the DTO fields. For example, if a consumer needs to see human-readable currency symbols or formatted amounts, you can do this transformation in the service and map it to the DTO before queuing it.
* **Security**: Only fields explicitly defined in the DTO are serialized, eliminating any risk of leaking sensitive internal data.

---

## 2. Refactored Implementation

To decouple the database model and fix the Go type compiler error (`cannot use req (type *dto.TransferRequest) as *models.Transaction value`), we applied the following changes:

### 1. Created the Webhook DTO
In [transaction_dto.go](file:///home/iamtbay/projects/tyr-fintech/internal/dto/transaction_dto.go), we added a dedicated payload type:
```go
type TransactionWebhookEvent struct {
	TransactionID string `json:"transaction_id"`
	FromWalletID  string `json:"from_wallet_id"`
	ToWalletID    string `json:"to_wallet_id"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
}
```

### 2. Updated the Webhook Queue and Worker
In [webhook_worker.go](file:///home/iamtbay/projects/tyr-fintech/internal/worker/webhook_worker.go), we changed the channel from accepting `*models.Transaction` to accepting `*dto.TransactionWebhookEvent`:
```go
var WebHookQueue = make(chan *dto.TransactionWebhookEvent, 100)

func StartWebhookWorker() {
	log.Println("Webhook worker has started, tasks are awaiting")
	for tx := range WebHookQueue {
		sendWebhook(tx)
	}
}

func sendWebhook(tx *dto.TransactionWebhookEvent) {
	// ...
	jsonData, _ := json.Marshal(tx)
	// ... Log fields updated to tx.TransactionID
}
```

### 3. Updated the Transaction Service Mapping
In [transaction_service.go](file:///home/iamtbay/projects/tyr-fintech/internal/services/transaction_service.go), upon a successful transfer, we map the request to the DTO and send it to the queue:
```diff
 	err := s.repo.Transfer(ctx, req)
 	if err != nil {
 		return err
 	}
 	//send webhook
-	worker.WebHookQueue <- req
+	worker.WebHookQueue <- &dto.TransactionWebhookEvent{
+		TransactionID: req.TransactionID,
+		FromWalletID:  req.FromWalletID,
+		ToWalletID:    req.ToWalletID,
+		Amount:        req.Amount,
+		Status:        "COMPLETED",
+	}
 	return nil
```
