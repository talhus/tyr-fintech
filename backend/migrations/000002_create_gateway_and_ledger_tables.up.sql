-- 1. Merchants Table: Represents B2B clients (e.g., Foodeli) integrating with TyrFintech gateway
CREATE TABLE IF NOT EXISTS merchants (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    settlement_wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE RESTRICT,
    fee_percentage_basis_points INT NOT NULL DEFAULT 200, -- 200 bps = 2.00% gateway fee
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'SUSPENDED')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. Merchant API Keys Table: Secure SHA-256 hashed API keys for authentication
CREATE TABLE IF NOT EXISTS merchant_api_keys (
    id UUID PRIMARY KEY,
    merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    key_hash VARCHAR(64) NOT NULL UNIQUE, -- SHA-256 hex string of the raw API key
    prefix VARCHAR(20) NOT NULL,          -- e.g., 'tyr_live_' or 'tyr_test_'
    masked_key VARCHAR(50) NOT NULL,      -- e.g., 'tyr_live_••••1234' for dashboard viewing
    environment VARCHAR(20) NOT NULL DEFAULT 'live' CHECK (environment IN ('live', 'test')),
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'REVOKED')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. System Accounts: Central internal accounts for settlement and revenue
CREATE TABLE IF NOT EXISTS system_accounts (
    id UUID PRIMARY KEY,
    account_type VARCHAR(50) NOT NULL UNIQUE CHECK (account_type IN ('ACQUIRING_CLEARING', 'PLATFORM_FEE_REVENUE')),
    currency VARCHAR(3) NOT NULL,
    balance BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 4. Charges Table: Tracks every payment request lifecycle, card metadata, and status
CREATE TABLE IF NOT EXISTS charges (
    id VARCHAR(64) PRIMARY KEY, -- e.g. 'trx_9876543210'
    merchant_id UUID NOT NULL REFERENCES merchants(id),
    amount BIGINT NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL,
    fee_amount BIGINT NOT NULL DEFAULT 0,
    net_amount BIGINT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL CHECK (status IN ('PENDING', 'SUCCEEDED', 'FAILED')),
    failure_reason VARCHAR(255) DEFAULT '',
    card_last4 VARCHAR(4) NOT NULL,
    card_holder_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 5. Ledger Entries: Immutable double-entry bookkeeping records
-- For every charge, sum(DEBIT) must strictly equal sum(CREDIT)
CREATE TABLE IF NOT EXISTS ledger_entries (
    id UUID PRIMARY KEY,
    charge_id VARCHAR(64) NOT NULL REFERENCES charges(id) ON DELETE CASCADE,
    account_type VARCHAR(50) NOT NULL CHECK (account_type IN ('ACQUIRING_CLEARING', 'MERCHANT_PAYABLE', 'PLATFORM_FEE_REVENUE')),
    account_id UUID NOT NULL, -- references system_accounts(id) or wallets(id)
    entry_type VARCHAR(10) NOT NULL CHECK (entry_type IN ('DEBIT', 'CREDIT')),
    amount BIGINT NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 6. Charge Idempotency Table: Stores cached HTTP status code and response body per (merchant_id, idempotency_key)
CREATE TABLE IF NOT EXISTS charge_idempotency (
    idempotency_key VARCHAR(255) NOT NULL,
    merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    request_hash VARCHAR(64) NOT NULL,
    response_status INT NOT NULL,
    response_body TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (merchant_id, idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_charges_merchant_id ON charges(merchant_id);
CREATE INDEX IF NOT EXISTS idx_ledger_entries_charge_id ON ledger_entries(charge_id);
CREATE INDEX IF NOT EXISTS idx_merchant_api_keys_hash ON merchant_api_keys(key_hash);
