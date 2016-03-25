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
	"github.com/opsee/hugs/obj"
)

type Postgres struct {
	db *sqlx.DB
}

func NewPostgres(connection string) (*Postgres, error) {
	db, err := sqlx.Connect("postgres", connection)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return &Postgres{
		db: db,
	}, nil
}

func (pg *Postgres) GetNotifications(user *com.User) ([]*obj.Notification, error) {
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
			log.WithFields(log.Fields{"postgres": "GetNotifications", "user": user, "notification": notification, "err": err}).Error("Couldn't scan notification.")
			return nil, err
		}
		notifications = append(notifications, &notification)
	}

	return notifications, err
}

func (pg *Postgres) UnsafeGetNotificationsByCheckID(checkID string) ([]*obj.Notification, error) {
	notifications := []*obj.Notification{}
	err := pg.db.Select(&notifications, "SELECT * from notifications WHERE check_id = $1", checkID)
	if err != nil {
		return nil, err
	}

	return notifications, nil
}

func (pg *Postgres) GetNotificationsByCheckID(user *com.User, checkID string) ([]*obj.Notification, error) {
	notifications := []*obj.Notification{}
	err := pg.db.Select(&notifications, "SELECT * from notifications WHERE check_id = $1 AND customer_id = $2", checkID, user.CustomerID)
	if err != nil {
		return nil, err
	}

	return notifications, nil
}

func (pg *Postgres) PutNotifications(user *com.User, notifications []*obj.Notification) error {
	tx, err := pg.db.Beginx()
	if err != nil {
		return err
	}

	for _, notification := range notifications {
		_, err := tx.NamedExec(
			`insert into notifications (customer_id, user_id, check_id, value, type)
			 values (:customer_id, :user_id, :check_id, :value, :type)
			 returning id`, notification)

		if err != nil {
			log.WithFields(log.Fields{"postgres": "PutNotifications", "user": user, "notification": notification, "error": err}).Error("Couldn't put notification.")
			if err := tx.Rollback(); err != nil {
				log.WithError(err).Error("Error rolling back transaction")
			}
			return fmt.Errorf("Couldn't put notification.")
		}
	}

	return tx.Commit()
}

func (pg *Postgres) UpdateNotification(user *com.User, notification *obj.Notification) error {
	oldNotification := obj.Notification{}
	err := pg.db.Get(&oldNotification, "SELECT * from notifications WHERE customer_id=$1 AND id=$2", user.CustomerID, notification.ID)
	if err != nil {
		return err
	}

	if oldNotification.CustomerID != user.CustomerID {
		log.WithFields(log.Fields{"postgres": "UpdateNotification", "user": user, "notification": notification}).Error("user.CustomerID, notification.CustomerID mistmatch!")
		return fmt.Errorf("Error: CustomerID associated with notification to be updated does not match CustomerID of requesting user.")
	}

	rows, err := pg.db.Queryx(`UPDATE notifications SET check_id=$1, value=$2, type=$3 WHERE id=$4`,
		notification.CheckID, notification.Value, notification.Type, notification.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	return nil
}

func (pg *Postgres) DeleteNotifications(user *com.User, notifications []*obj.Notification) error {
	for _, notification := range notifications {
		err := pg.DeleteNotification(user, notification)
		if err != nil {
			log.WithFields(log.Fields{"postgres": "DeleteNotifications", "user": user, "notification": notification, "error": err}).Error("Couldn't delete notification")
			return fmt.Errorf("Couldn't delete notificaiton.")
		}
	}

	return nil
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
