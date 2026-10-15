package pg

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Connection struct {
	dsn           string
	pgxConnection *pgx.Conn
}

func NewConnection(ctx context.Context, dsn string) (*Connection, error) {
	// dsn will be the postgres connection string pg:// "postgres://username:password@localhost:5432/database_name"
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return &Connection{
		dsn:           dsn,
		pgxConnection: conn,
	}, nil
}
