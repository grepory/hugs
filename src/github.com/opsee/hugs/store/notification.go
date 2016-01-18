package store

import
//"encoding/json"

_ "github.com/lib/pq"

type Notification struct {
	ID         int    `json:"id" db:"id"`
	CustomerID string `json:"customer_id" db:"customer_id"`
	UserID     int    `json:"user_id" db:"user_id"`
	CheckID    string `json:"check_id" db:"check_id"`
	Value      string `json:"value" db:"value"`
	Type       string `json:"type" db:"type"`
}
