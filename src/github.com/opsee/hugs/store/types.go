package store

import (
	"github.com/jmoiron/sqlx/types"
	_ "github.com/lib/pq"
)

type Notification struct {
	ID         int    `json:"id" db:"id"`
	CustomerID string `json:"customer_id" db:"customer_id"`
	UserID     int    `json:"user_id" db:"user_id"`
	CheckID    string `json:"check_id" db:"check_id"`
	Value      string `json:"value" db:"value"`
	Type       string `json:"type" db:"type"`
}

type SlackOAuthResponseDBWrapper struct {
	ID         int            `json:"id" db:"id"`
	CustomerID string         `json:"customer_id" db:"customer_id"`
	Data       types.JSONText `json:"data" db:"data"`
}

type Customer struct {
	ID string `json:"id" db:"id"`
}
