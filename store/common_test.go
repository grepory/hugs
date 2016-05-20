package store

import (
	"github.com/opsee/basic/schema"
	"github.com/opsee/hugs/obj"
	log "github.com/sirupsen/logrus"
)

type StoreTest struct {
	DBStore       *Postgres
	Notifications []*obj.Notification
	User          *schema.User
}

func NewStoreTest() *StoreTest {
	log.Info("Connecting to local test store")
	db, err := NewPostgres()
	if err != nil {
		panic(err)
	}

	user := &schema.User{
		Id:         13,
		CustomerId: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
	}

	log.Info("Clearing local test store of notifications")
	err = db.DeleteNotificationsByUser(user)
	if err != nil {
		log.Warn("Warning: Couldn't clear local test store of notifications")
	}

	return &StoreTest{
		DBStore: db,
		User:    user,
		Notifications: []*obj.Notification{
			&obj.Notification{
				Id:         0,
				CustomerId: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
				UserId:     13,
				CheckId:    "00001",
				Value:      "off",
				Type:       "slack_bot",
			},
			&obj.Notification{
				Id:         1,
				CustomerId: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
				UserId:     13,
				CheckId:    "00001",
				Value:      "you",
				Type:       "email",
			},
			&obj.Notification{
				Id:         2,
				CustomerId: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
				UserId:     13,
				CheckId:    "00001",
				Value:      "fuck",
				Type:       "webhook",
			},
		},
	}
}

var Common = NewStoreTest()
