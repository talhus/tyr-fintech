package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresDB struct {
	DB *pgxpool.Pool
}

func (p *PostgresDB) Close() {
	p.DB.Close()
}

func Connect(dsn string) (*PostgresDB, error) {

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	err = pool.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}
	_, err = pool.Exec(context.Background(), `UPDATE wallets SET wallet_number = nextval('wallet_number_seq') WHERE wallet_number IS NULL;`)
	if err != nil {
		fmt.Printf("Warning: failed to backfill missing wallet numbers: %v\n", err)
	}
	_, err = pool.Exec(context.Background(), `ALTER TABLE transactions ADD COLUMN IF NOT EXISTS converted_amount BIGINT;`)
	if err != nil {
		fmt.Printf("Warning: failed to add converted_amount column to transactions table: %v\n", err)
	}
	fmt.Println("Successfully connected to PostgresDB")
	return &PostgresDB{DB: pool}, nil
}
