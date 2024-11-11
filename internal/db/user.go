package db

import (
	"NodeArt/internal"
	"NodeArt/internal/db/user_repo/users"
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
	"github.com/jackc/pgx/v5"
	"log"
)

const salt = "salt" // should be more secure :)

type Storage struct {
	Conn *pgx.Conn
}

func (s Storage) GetUser(email, password string) (string, error) {
	ctx := context.Background()

	queries := nodeart.New(s.Conn)
	getUserCondition := nodeart.GetUserParams{
		Email:    email,
		Password: getMD5Hash(password),
	}
	u, err := queries.GetUser(ctx, getUserCondition)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("fetching failed: %s", err)

		return "", err
	}

	return u.Email, nil
}

func (s Storage) InsertUser(email, password string) error {
	ctx := context.Background()

	queries := nodeart.New(s.Conn)
	getUserCondition := nodeart.GetUserParams{
		Email:    email,
		Password: getMD5Hash(password),
	}
	u, err := queries.GetUser(ctx, getUserCondition)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if u.Email != "" {
		return internal.UserAlreadyExists
	}
	data := nodeart.InsertUserParams{
		Username: email,
		Password: getMD5Hash(password),
		Email:    email,
	}

	_, err = queries.InsertUser(ctx, data)
	if err != nil {
		log.Printf("insertion failed: %s", err)

		return err
	}

	return nil
}

func getMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
