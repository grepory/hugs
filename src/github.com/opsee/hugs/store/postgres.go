package store

import (
	//"encoding/json"

	"encoding/json"
	"fmt"
	"log"

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

	_, err = pg.db.Queryx(`UPDATE notifications SET check_id=$1, value=$2, type=$3 WHERE id=$4`,
		notification.CheckID, notification.Value, notification.Type, notification.ID)

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

func (pg *Postgres) DeleteSlackOAuthResponsesByUser(user *com.User) error {
	_, err := pg.db.Queryx(`DELETE from slack_oauth_responses WHERE customer_id=$1`, user.CustomerID)
	return err
}

// TODO(dan) decide whether we want to limit customer-ids to one slack integration
// TODO(dan) right now we delete all of the existing responses prior to adding one
func (pg *Postgres) PutSlackOAuthResponse(user *com.User, s *apiutils.SlackOAuthResponse) error {
	datjson, err := json.Marshal(s)
	if err != nil {
		return err
	}

	// Ensure we only have one for right now
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

// TODO(dan) Operating under the assumption that one token/user, this will return that one token
// leaving the GetSlackOAuthReponses in case we allow more than one integration per customer
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

// TODO(dan) decide whether we want to limit customer-ids to one slack integration
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

// TODO(dan) decide whether we want to limit customer-ids to one slack integration
func (pg *Postgres) UpdateSlackOAuthResponse(user *com.User, s *apiutils.SlackOAuthResponse) error {
	datjson, err := json.Marshal(s)
	if err != nil {
		return err
	}
	data := types.JSONText(string(datjson))
	_, err = pg.db.Queryx(`UPDATE slack_oauth_responses SET data=$1 where customer_id=$2`, data, user.CustomerID)
	return err
}
