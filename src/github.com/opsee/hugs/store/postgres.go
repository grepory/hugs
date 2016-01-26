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
	"github.com/opsee/hugs/apiutils"
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
			log.WithFields(log.Fields{"postgres": "GetNotifications", "user": user, "notification": notification, "err": err}).Error("Couldn't scan notification.")
			return nil, err
		}
		notifications = append(notifications, &notification)
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
	err := pg.db.Select(&notifications, "SELECT * from notifications WHERE check_id = $1 AND customer_id = $2", checkID, user.CustomerID)

	return notifications, err
}

func (pg *Postgres) PutNotifications(user *com.User, notifications []*Notification) error {
	for _, notification := range notifications {
		err := pg.PutNotification(user, notification)
		if err != nil {
			log.WithFields(log.Fields{"postgres": "PutNotifications", "user": user, "notification": notification, "error": err}).Error("Couldn't put notification.")
			return fmt.Errorf("Couldn't put notification.")
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
	if notification.CustomerID != user.CustomerID {
		return fmt.Errorf("Customer ID does not match notification ID")
	}
	_, err := pg.db.NamedExec(
		`insert into notifications (customer_id, user_id, check_id, value, type)
                 values (:customer_id, :user_id, :check_id, :value, :type)
                 returning id`, notification)
	return err
}

func (pg *Postgres) UpdateNotification(user *com.User, notification *Notification) error {
	oldNotification := Notification{}
	err := pg.db.Get(&oldNotification, "SELECT * from notifications WHERE customer_id=$1 AND id=$2", user.CustomerID, notification.ID)
	if err != nil {
		return err
	}

	if oldNotification.CustomerID != user.CustomerID {
		log.WithFields(log.Fields{"postgres": "UpdateNotification", "user": user, "notification": notification}).Error("user.CustomerID, notification.CustomerID mistmatch!")
		return fmt.Errorf("Error: CustomerID associated with notification to be updated does not match CustomerID of requesting user.")
	}

	_, err = pg.db.Queryx(`UPDATE notifications SET check_id=$1, value=$2, type=$3 WHERE id=$4`,
		notification.CheckID, notification.Value, notification.Type, notification.ID)

	return err
}

func (pg *Postgres) DeleteNotifications(user *com.User, notifications []*Notification) error {
	for _, notification := range notifications {
		err := pg.DeleteNotification(user, notification)
		if err != nil {
			log.WithFields(log.Fields{"postgres": "DeleteNotifications", "user": user, "notification": notification, "error": err}).Error("Couldn't delete notification")
			return fmt.Errorf("Couldn't delete notificaiton.")
		}
	}

	return nil
}

func (pg *Postgres) DeleteNotification(user *com.User, notification *Notification) error {
	_, err := pg.db.Queryx(`DELETE from notifications WHERE id=$1 AND customer_id=$2`, notification.ID, user.CustomerID)
	return err
}

func (pg *Postgres) DeleteNotificationsByUser(user *com.User) error {
	_, err := pg.db.Queryx(`DELETE from notifications WHERE customer_id=$1`, user.CustomerID)
	return err
}

func (pg *Postgres) DeleteSlackOAuthResponsesByUser(user *com.User) error {
	_, err := pg.db.Queryx(`DELETE from slack_oauth_responses WHERE customer_id=$1`, user.CustomerID)
	return err
}

func (pg *Postgres) PutSlackOAuthResponse(user *com.User, s *apiutils.SlackOAuthResponse) error {
	customer := Customer{}
	err := pg.db.Get(&customer, "SELECT * from customers WHERE id=$1", user.CustomerID)

	if customer.ID == "" {
		log.WithFields(log.Fields{"postgres": "PutSlackOAuthResponse", "CustomerID": user}).Error("Adding new customer.")
		pg.db.MustExec("INSERT INTO customers (id) VALUES ($1)", user.CustomerID)
	}

	datjson, err := json.Marshal(s)
	if err != nil {
		return err
	}

	err = pg.DeleteSlackOAuthResponsesByUser(user)
	if err != nil {
		return err
	}

	wrapper := SlackOAuthResponseDBWrapper{
		CustomerID: user.CustomerID,
		Data:       types.JSONText(string(datjson)),
	}

	_, err = pg.db.NamedExec("INSERT INTO slack_oauth_responses (customer_id, data) VALUES (:customer_id, :data)", wrapper)
	return err
}

func (pg *Postgres) GetSlackOAuthResponse(user *com.User) (*apiutils.SlackOAuthResponse, error) {
	oaResponses, err := pg.GetSlackOAuthResponses(user)
	if err != nil {
		return nil, err
	}

	if len(oaResponses) > 0 {
		return oaResponses[0], nil
	}

	return nil, nil
}

func (pg *Postgres) GetSlackOAuthResponses(user *com.User) ([]*apiutils.SlackOAuthResponse, error) {
	oaResponses := []*apiutils.SlackOAuthResponse{}
	rows, err := pg.db.Queryx("SELECT data from slack_oauth_responses WHERE customer_id = $1", user.CustomerID)
	if err != nil {
		return oaResponses, err
	}

	for rows.Next() {
		var wrappedOAResponse SlackOAuthResponseDBWrapper
		err := rows.StructScan(&wrappedOAResponse)
		if err != nil {
			log.Fatalln(err)
		}

		oaResponse := apiutils.SlackOAuthResponse{}
		err = wrappedOAResponse.Data.Unmarshal(&oaResponse)
		if err != nil {
			continue
		}

		oaResponses = append(oaResponses, &oaResponse)
	}

	return oaResponses, err
}

func (pg *Postgres) UpdateSlackOAuthResponse(user *com.User, s *apiutils.SlackOAuthResponse) error {
	datjson, err := json.Marshal(s)
	if err != nil {
		return err
	}
	data := types.JSONText(string(datjson))
	_, err = pg.db.Queryx(`UPDATE slack_oauth_responses SET data=$1 where customer_id=$2`, data, user.CustomerID)
	return err
}
