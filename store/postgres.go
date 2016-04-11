package store

import (
	//"encoding/json"

	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	_ "github.com/lib/pq"
	"github.com/opsee/basic/com"
	"github.com/opsee/hugs/config"
	"github.com/opsee/hugs/obj"
)

type Postgres struct {
	db *sqlx.DB
}

func NewPostgres() (*Postgres, error) {
	return &Postgres{
		db: config.GetConfig().DBConnection,
	}, nil
}

// given a list of notifications, returns a list of notifications for each (if it exists)
// TODO(dan) problem is when one of these calls fails
func (pg *Postgres) GetNotifications(user *com.User, oldNotifications []*obj.Notification) ([]*obj.Notification, error) {
	notifications := []*obj.Notification{}
	for _, oldNotification := range oldNotifications {
		newNotification, err := pg.GetNotification(user, oldNotification.ID)
		if err != nil {
			log.WithError(err).Errorf("Failed to get notification %d, for user %s", oldNotification.ID, user.CustomerID)
		}
		notifications = append(notifications, newNotification)
	}

	return notifications, nil
}

func (pg *Postgres) GetNotification(user *com.User, id int) (*obj.Notification, error) {
	notification := &obj.Notification{}
	err := pg.db.Get(notification, "SELECT * FROM notifications WHERE customer_id = $1 AND id = $2", user.CustomerID, id)
	return notification, err
}

func (pg *Postgres) GetNotificationsByUser(user *com.User) ([]*obj.Notification, error) {
	var notifications []*obj.Notification
	rows, err := pg.db.Queryx("SELECT * from notifications WHERE customer_id = $1", user.CustomerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var notification obj.Notification
		err := rows.StructScan(&notification)
		if err != nil {
			log.WithError(err).Errorf("Couldn't scan notification for user %s", user.CustomerID)
			return nil, err
		}
		notifications = append(notifications, &notification)
	}

	return notifications, err
}

func (pg *Postgres) deleteNotification(x sqlx.Ext, notification *obj.Notification) error {
	_, err := sqlx.NamedExec(x, `delete from notifications where id = :id  AND customer_id = :customer_id`, notification)
	return err
}

func (pg *Postgres) updateNotification(x sqlx.Ext, notification *obj.Notification) error {
	_, err := sqlx.NamedExec(x,
		`UPDATE notifications set customer_id = :customer_id, user_id = :user_id, check_id = :check_id, value = :value, type = :type)
		WHERE notification_id = :id`, notification)
	return err
}

func (pg *Postgres) putNotification(x sqlx.Ext, notification *obj.Notification) error {
	_, err := sqlx.NamedExec(x,
		`INSERT INTO notifications (customer_id, user_id, check_id, value, type)
		VALUES (:customer_id, :user_id, :check_id, :value, :type)
		RETURNING id`, notification)
	return err
}

func (pg *Postgres) PutNotifications(notifications []*obj.Notification) error {
	tx, err := pg.db.Beginx()
	if err != nil {
		return err
	}

	for _, notification := range notifications {
		err = pg.deleteNotification(tx, notification)
		if err != nil {
			log.WithError(err).Errorf("Couldn't delete notification %d for user %s", notification.ID, notification.CustomerID)
			if err := tx.Rollback(); err != nil {
				log.WithError(err).Error("Error rolling back transaction")
			}
			return fmt.Errorf("Couldn't delete notification.")
		}

		err = pg.putNotification(tx, notification)
		if err != nil {
			log.WithError(err).Errorf("Couldn't put notification %d for user %s", notification.ID, notification.CustomerID)
			if err := tx.Rollback(); err != nil {
				log.WithError(err).Error("Error rolling back transaction")
			}
			return fmt.Errorf("Couldn't put notification.")
		}
	}
	return tx.Commit()
}

func (pg *Postgres) DeleteNotifications(notifications []*obj.Notification) error {
	tx, err := pg.db.Beginx()
	if err != nil {
		return err
	}

	for _, notification := range notifications {
		err = pg.deleteNotification(tx, notification)
		if err != nil {
			log.WithError(err).Errorf("Couldn't delete notification %s for user %s", notification.ID, notification.CustomerID)
			if err := tx.Rollback(); err != nil {
				log.WithError(err).Error("Error rolling back transaction")
			}
			return fmt.Errorf("Couldn't delete notification.  Rolled back.")
		}
	}
	return tx.Commit()
}

func (pg *Postgres) GetNotificationsByCheckID(user *com.User, checkID string) ([]*obj.Notification, error) {
	notifications := []*obj.Notification{}
	err := pg.db.Select(&notifications, "SELECT * from notifications WHERE check_id = $1 AND customer_id = $2", checkID, user.CustomerID)
	if err != nil {
		return nil, err
	}

	return notifications, nil
}

func (pg *Postgres) UnsafeGetNotificationsByCheckID(checkID string) ([]*obj.Notification, error) {
	notifications := []*obj.Notification{}
	err := pg.db.Select(&notifications, "SELECT * from notifications WHERE check_id = $1", checkID)
	if err != nil {
		return nil, err
	}

	return notifications, nil
}

func (pg *Postgres) DeleteNotification(user *com.User, notification *obj.Notification) error {
	rows, err := pg.db.Queryx(`DELETE from notifications WHERE id=$1 AND customer_id=$2`, notification.ID, user.CustomerID)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
}

func (pg *Postgres) DeleteNotificationsByUser(user *com.User) error {
	rows, err := pg.db.Queryx(`DELETE from notifications WHERE customer_id=$1`, user.CustomerID)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
}

func (pg *Postgres) DeleteNotificationsByCheckId(user *com.User, checkId string) error {
	rows, err := pg.db.Queryx(`DELETE from notifications WHERE customer_id=$1 AND check_id=$2`, user.CustomerID, checkId)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
}

func (pg *Postgres) DeleteSlackOAuthResponsesByUser(user *com.User) error {
	rows, err := pg.db.Queryx(`DELETE from slack_oauth_responses WHERE customer_id=$1`, user.CustomerID)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
}

func (pg *Postgres) PutSlackOAuthResponse(user *com.User, s *obj.SlackOAuthResponse) error {
	datjson, err := json.Marshal(s)
	if err != nil {
		return err
	}

	err = pg.DeleteSlackOAuthResponsesByUser(user)
	if err != nil {
		return err
	}

	wrapper := obj.SlackOAuthResponseDBWrapper{
		CustomerID: user.CustomerID,
		Data:       types.JSONText(string(datjson)),
	}

	_, err = pg.db.NamedExec("INSERT INTO slack_oauth_responses (customer_id, data) VALUES (:customer_id, :data)", wrapper)
	return err
}

func (pg *Postgres) GetSlackOAuthResponse(user *com.User) (*obj.SlackOAuthResponse, error) {
	oaResponses, err := pg.GetSlackOAuthResponses(user)
	if err != nil {
		return nil, err
	}

	if len(oaResponses) > 0 {
		return oaResponses[0], nil
	}

	return nil, nil
}

func (pg *Postgres) GetSlackOAuthResponses(user *com.User) ([]*obj.SlackOAuthResponse, error) {
	oaResponses := []*obj.SlackOAuthResponse{}
	rows, err := pg.db.Queryx("SELECT data from slack_oauth_responses WHERE customer_id = $1", user.CustomerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var wrappedOAResponse obj.SlackOAuthResponseDBWrapper
		err := rows.StructScan(&wrappedOAResponse)
		if err != nil {
			log.Fatalln(err)
		}

		oaResponse := obj.SlackOAuthResponse{}
		err = wrappedOAResponse.Data.Unmarshal(&oaResponse)
		if err != nil {
			continue
		}

		oaResponses = append(oaResponses, &oaResponse)
	}

	return oaResponses, err
}

func (pg *Postgres) UpdateSlackOAuthResponse(user *com.User, s *obj.SlackOAuthResponse) error {
	datjson, err := json.Marshal(s)
	if err != nil {
		return err
	}
	data := types.JSONText(string(datjson))
	rows, err := pg.db.Queryx(`UPDATE slack_oauth_responses SET data=$1 where customer_id=$2`, data, user.CustomerID)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
}
