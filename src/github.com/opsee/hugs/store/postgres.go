package store

import (
	"database/sql"
	"encoding/json"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/opsee/basic/com"
)

type Postgres struct {
	db *sqlx.DB
}

func NewPostgres(connection string) (Store, error) {
	db, err := sqlx.Open("postgres", connection)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(64)
	db.SetMaxIdleConns(8)

	return &Postgres{
		db: db,
	}, nil
}

type Notification struct {
	ID         int    `json:"id" db:"id"`
	CustomerID string `json:"customer_id" db:"customer_id"`
	CheckID    string `json:"check_id" db:"check_id"`
	Value      string `json:"value" db:"value"`
}

func (pg *Postgres) GetNotifications(user *com.User) ([]*Notification, error) {
	notifications := []*Notification{}
	err := pg.db.Get(notifications,
		"select * from notifications where customer_id = $1",
		user.CustomerID,
	)

	return notifications, err
}

func (pg *Postgres) PutNotification(notification *Notification) error {
	var id string
	err := sqlx.NamedExec(
		pg.db,
		&id,
		`insert into notifications (customer_id, user_id, check_id, value)
                 values (:customer_id, :user_id, :check_id, :value)
                 returning id`,
		notification)

	notification.ID = id
	return err
}

func (pg *Postgres) UpdateNotification(notification *Notification) error {
	_, err := pg.db.NamedExec(
		`update notifications set customer_id = :customer_id, user_id = :user_id, value := value`,
		notification)
	return err
}

func (pg *Postgres) DeleteNotification(notification *Notification) error {
	_, err := pg.db.NamedExec(
		`delete from notifications where id = :id`,
		notification,
	)

	return err
}
