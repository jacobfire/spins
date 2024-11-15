package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log"
)

type Config struct {
	Host     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

func NewConnection(ctx context.Context, c Config) (*pgx.Conn, error) {
	log.Printf("host=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Username, c.Password, c.DBName, c.SSLMode)
	connString := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Username, c.Password, c.DBName, c.SSLMode)
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		log.Printf("connection not established: %s", err)
		return nil, err
	}

	return conn, nil
}
