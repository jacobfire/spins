package db

import (
	nodeart_balance "NodeArt/internal/db/balance_repo/balance"
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"log"
)

type BalanceStorage struct {
	Conn *pgx.Conn
}

func (b BalanceStorage) AddDeposit(email string, amount float64) (float64, error) {
	ctx := context.Background()

	var fl pgtype.Float8
	fl.Scan(amount)
	depositParam := nodeart_balance.AddBalanceParams{
		Email:   email,
		Balance: fl,
	}
	queries := nodeart_balance.New(b.Conn)

	bl, err := queries.AddBalance(ctx, depositParam)
	if err != nil {
		log.Printf("fetching failed: %s", err)

		return 0, err
	}

	return bl.Balance.Float64, nil
}

func (b BalanceStorage) UpdateDeposit(email string, amount float64) error {
	ctx := context.Background()

	var fl pgtype.Float8
	fl.Scan(amount)
	depositParam := nodeart_balance.UpdateBalanceParams{
		Email:   email,
		Balance: fl,
	}
	queries := nodeart_balance.New(b.Conn)

	err := queries.UpdateBalance(ctx, depositParam)
	if err != nil {
		log.Printf("fetching failed: %s", err)

		return err
	}

	return nil
}

func (b BalanceStorage) UpdateWithNewValueDeposit(email string, amount float64) error {
	ctx := context.Background()

	var fl pgtype.Float8
	fl.Scan(amount)
	depositParam := nodeart_balance.UpdateWithNewValueBalanceParams{
		Email:   email,
		Balance: fl,
	}
	queries := nodeart_balance.New(b.Conn)

	err := queries.UpdateWithNewValueBalance(ctx, depositParam)
	if err != nil {
		log.Printf("fetching failed: %s", err)

		return err
	}

	return nil
}

func (b BalanceStorage) SubDeposit(email string, amount float64) (float64, error) {
	ctx := context.Background()

	var fl pgtype.Float8
	fl.Scan(amount)
	subParam := nodeart_balance.SubBalanceParams{
		Email:   email,
		Balance: fl,
	}
	queries := nodeart_balance.New(b.Conn)

	err := queries.SubBalance(ctx, subParam)
	if err != nil {
		log.Printf("fetching failed: %s", err)

		return 0, err
	}

	return 0, nil
}

func (b BalanceStorage) GetBalance(email string) (float64, error) {
	ctx := context.Background()

	depositParam := nodeart_balance.GetBalanceParams{
		Email: email,
	}
	queries := nodeart_balance.New(b.Conn)

	bl, err := queries.GetBalance(ctx, depositParam)
	if err != nil {
		log.Printf("fetching failed: %s", err)

		return 0, err
	}

	return bl.Float64, nil
}
