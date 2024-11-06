package db

import (
	nodeart_spin "NodeArt/internal/db/spin_repo/spin"
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type SpinStorage struct {
	Conn *pgx.Conn
}

func (ss SpinStorage) AddSpin(email string, combination string) error {
	ctx := context.Background()

	queries := nodeart_spin.New(ss.Conn)
	var userID pgtype.Int4
	userID.Scan(1)
	var co pgtype.Text
	co.Scan(combination)
	params := nodeart_spin.AddSpinParams{
		UserID:      userID,
		Combination: co,
		Email:       email,
	}

	_, err := queries.AddSpin(ctx, params)

	return err
}

func (ss SpinStorage) GetSpinHistory(email string) ([]nodeart_spin.Spin, error) {
	ctx := context.Background()

	queries := nodeart_spin.New(ss.Conn)
	var userID pgtype.Int4
	userID.Scan(1)
	params := nodeart_spin.GetSpinHistoryParams{
		UserID: userID,
		Email:  email,
	}

	s, err := queries.GetSpinHistory(ctx, params)

	return s, err
}
