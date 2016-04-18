package store

import (
	//"encoding/json"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/opsee/hugs/config"
)

type Postgres struct {
	db *sqlx.DB
}

func NewPostgres() (*Postgres, error) {
	return &Postgres{
		db: config.GetConfig().DBConnection,
	}, nil
}
