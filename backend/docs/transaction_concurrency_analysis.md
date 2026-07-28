# Transaction Locking & Concurrency Analysis

This document explains the behavior of `FOR UPDATE` in database transactions, addresses whether it is needed for all queries within a transaction, and details the fix for the critical concurrency bug found in [transaction_repository.go](file:///home/iamtbay/projects/tyr-fintech/internal/repos/transaction_repository.go).

---

## 1. Do you need `FOR UPDATE` on all SQL queries inside a transaction?

**No, you do not.** 

Here is why:
1. **`FOR UPDATE` is only valid on `SELECT` statements**: Using `FOR UPDATE` on an `UPDATE`, `INSERT`, or `DELETE` statement is a syntax error in PostgreSQL (and most other relational databases).
2. **`UPDATE` statements lock rows automatically**: In databases like PostgreSQL, any `UPDATE` (or `DELETE`) statement automatically acquires an exclusive row-level lock on the rows it matches and modifies. Any concurrent transaction attempting to read the same rows with `FOR UPDATE` or write to them will block until your transaction completes (commits or rolls back).
3. **Lock inheritance inside a transaction**: Once a row is locked in a transaction (either explicitly via `SELECT ... FOR UPDATE` or implicitly via `UPDATE`), subsequent operations on that row within the *same* transaction do not need any special keywords to retain the lock.

---

## 2. The Critical Bug in `transaction_repository.go`

In the original code, the query was:
```sql
UPDATE wallets SET balance=balance-$1 WHERE id=$2 AND balance>=$1 FOR UPDATE
```
This had two issues:
1. **Syntax Error**: The `FOR UPDATE` clause caused a database syntax error (SQLSTATE `42601`).
2. **Concurrency / Race Condition (Double Spend)**: 
   If two concurrent transactions ($Tx_A$ and $Tx_B$) tried to transfer money from the same wallet with a balance of $100$:
   * $Tx_A$ locks the wallet and subtracts $100$. The balance becomes $0$.
   * $Tx_B$ is blocked waiting for $Tx_A$'s lock.
   * $Tx_A$ successfully completes the rest of the queries (updates the receiver, logs the transaction) and commits.
   * $Tx_B$ acquires the lock and executes its `UPDATE` statement. Because the balance is now $0$, the condition `balance >= 100` is **false**.
   * In PostgreSQL, when an `UPDATE` matches 0 rows, it returns successfully but updates 0 rows.
   * The Go code did not check the number of affected rows, so it proceeded to update the receiver's wallet (adding $100$) and successfully committed!
   * **Result**: The sender lost $100$ once, but the receiver gained $200$.

---

## 3. How We Fixed It

We updated the transaction repository in [transaction_repository.go](file:///home/iamtbay/projects/tyr-fintech/internal/repos/transaction_repository.go):

1. **Removed the invalid `FOR UPDATE` keyword** from the `UPDATE` query.
2. **Checked `RowsAffected()`**: We inspect the result returned by `tx.Exec`. If `RowsAffected() == 0`, we immediately return an error, which triggers the deferred `tx.Rollback(ctx)` to abort the transaction.

### Code Diff

```diff
-	//sender
-	_, err = tx.Exec(ctx, `UPDATE wallets SET balance=balance-$1 WHERE id=$2 AND balance>=$1 FOR UPDATE`, req.Amount, req.FromWalletID)
-	if err != nil {
-		return err
-	}
+	//sender (UPDATE automatically acquires an exclusive row-level lock on the row)
+	res, err := tx.Exec(ctx, `UPDATE wallets SET balance=balance-$1 WHERE id=$2 AND balance>=$1`, req.Amount, req.FromWalletID)
+	if err != nil {
+		return err
+	}
+	if res.RowsAffected() == 0 {
+		return errors.New("insufficient balance or wallet not found")
+	}
```

---

## 4. Alternative Approach: Explicit `SELECT ... FOR UPDATE`

If you prefer to perform business logic in Go before updating the database, you can use `SELECT ... FOR UPDATE` to lock the sender's wallet row at the beginning of the transaction:

```go
// 1. Lock and read the sender's wallet
var balance int64
err = tx.QueryRow(ctx, "SELECT balance FROM wallets WHERE id=$1 FOR UPDATE", req.FromWalletID).Scan(&balance)
if err != nil {
    return err // e.g. wallet not found
}

// 2. Perform validation in Go
if balance < req.Amount {
    return errors.New("insufficient balance")
}

// 3. Perform the update (no other transaction can modify this row until tx commits/rolls back)
_, err = tx.Exec(ctx, "UPDATE wallets SET balance=balance-$1 WHERE id=$2", req.Amount, req.FromWalletID)
if err != nil {
    return err
}
```
Both the direct `UPDATE + RowsAffected()` check and the `SELECT FOR UPDATE + UPDATE` patterns are concurrency-safe and correct.
