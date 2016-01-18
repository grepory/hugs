package store

import (
	//"encoding/json"

	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/opsee/basic/com"
)

type Postgres struct {
	db *sqlx.DB
}

func NewPostgres(connection string) (*Postgres, error) {
	db, err := sqlx.Connect("postgres", connection)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)

	return &Postgres{
		db: db,
	}, nil
}

func (pg *Postgres) GetNotifications(user *com.User) ([]*Notification, error) {
	var notifications []*Notification
	rows, err := pg.db.Queryx("SELECT * from notifications WHERE customer_id = $1", user.CustomerID)
	for rows.Next() {
		var notification Notification
		err := rows.StructScan(&notification)
		if err != nil {
			log.Fatalln(err)
		}
		notifications = append(notifications, &notification)
		fmt.Printf("%#v\n", notification)
	}

	return notifications, err
}

func (pg *Postgres) UnsafeGetNotificationsByCheckID(checkID string) ([]*Notification, error) {
	notifications := []*Notification{}
	err := pg.db.Select(&notifications, "SELECT * from notifications WHERE check_id = $1", checkID)

	return notifications, err
}

func (pg *Postgres) GetNotificationsByCheckID(user *com.User, checkID string) ([]*Notification, error) {
	notifications := []*Notification{}
	err := pg.db.Select(&notifications, "SELECT * from notifications WHERE check_id = $1", checkID)

	// check to ensure that user matches returned notifications user
	if err == nil {
		for _, notification := range notifications {
			// TODO(dan) Also check CustomerID?
			if notification.UserID != user.ID {
				return nil, fmt.Errorf("UserID does not match Notification UserID")
			}
		}
	}

	return notifications, err
}

func (pg *Postgres) PutNotifications(user *com.User, notifications []*Notification) error {
	// TODO(dan) Should we return after the first error?
	for _, notification := range notifications {
		err := pg.PutNotification(user, notification)
		if err != nil {
			return err
		}
	}
	return nil
}

func (pg *Postgres) UnsafePutNotification(notification *Notification) error {
	_, err := pg.db.NamedExec(
		`insert into notifications (customer_id, user_id, check_id, value, type)
                 values (:customer_id, :user_id, :check_id, :value, :type)
                 returning id`, notification)
	return err
}

func (pg *Postgres) PutNotification(user *com.User, notification *Notification) error {
	// TODO(dan) This is a little funky.  How do we know if a notification already exists??
	if notification.UserID != user.ID {
		return fmt.Errorf("User ID does not match Notification UserID")
	}

	_, err := pg.db.NamedExec(
		`insert into notifications (customer_id, user_id, check_id, value, type)
                 values (:customer_id, :user_id, :check_id, :value, :type)
                 returning id`, notification)
	return err
}

func (pg *Postgres) UpdateNotification(user *com.User, notification *Notification) error {
	// Check to ensure notification in db has CustomerID that matches authenticated user
	oldNotification := Notification{}
	err := pg.db.Get(&oldNotification, "SELECT * from notifications WHERE customer_id=$1 AND id=$2", user.CustomerID, notification.ID)
	if err != nil {
		return err
	}

	// TODO(dan) Also check CustomerID?
	if oldNotification.UserID != user.ID {
		return fmt.Errorf("User is not allowed to modify this notification")
	}

	_, err = pg.db.Queryx(`UPDATE notifications SET check_id=$1, value=$2, type=$3`,
		notification.CheckID, notification.Value, notification.Type)

	return err
}

func (pg *Postgres) DeleteNotifications(user *com.User, notifications []*Notification) error {
	// TODO(dan) Should we return after the first error?
	for _, notification := range notifications {
		err := pg.DeleteNotification(user, notification)
		if err != nil {
			return err
		}
	}
	return nil
}

func (pg *Postgres) DeleteNotification(user *com.User, notification *Notification) error {
	// Check to ensure notification in db has CustomerID that matches authenticated user
	oldNotification := Notification{}
	err := pg.db.Get(&oldNotification, "SELECT * from notifications WHERE customer_id=$1 AND id=$2", user.CustomerID, notification.ID)
	if err != nil {
		return err
	}

	// TODO(dan) Also check CustomerID?
	if notification.UserID != user.ID {
		return fmt.Errorf("User is not allowed to modify this notification")
	}

	// TODO(dan) Get Notification and check actual UserID prior to deleting
	if notification.UserID != user.ID {
		return fmt.Errorf("User ID does not match Notification UserID")
	}

	_, err = pg.db.Queryx(`DELETE from notifications WHERE id=$1`, oldNotification.ID)
	return err
}

func (pg *Postgres) DeleteNotificationsByUser(user *com.User) error {
	_, err := pg.db.Queryx(`DELETE from notifications WHERE customer_id=$1`, user.CustomerID)
	return err
}
