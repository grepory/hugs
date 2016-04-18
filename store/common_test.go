package store

import (
	"github.com/opsee/basic/com"
	"github.com/opsee/hugs/obj"
	log "github.com/sirupsen/logrus"
)

type StoreTest struct {
	DBStore       *Postgres
	Notifications []*obj.Notification
	User          *com.User
}

func NewStoreTest() *StoreTest {
	log.Info("Connecting to local test store")
	db, err := NewPostgres()
	if err != nil {
		panic(err)
	}

	user := &com.User{
		ID:         13,
		CustomerID: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
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
				ID:         0,
				CustomerID: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
				UserID:     13,
				CheckID:    "00001",
				Value:      "off",
				Type:       "slack_bot",
			},
			&obj.Notification{
				ID:         1,
				CustomerID: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
				UserID:     13,
				CheckID:    "00001",
				Value:      "you",
				Type:       "email",
			},
			&obj.Notification{
				ID:         2,
				CustomerID: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
				UserID:     13,
				CheckID:    "00001",
				Value:      "fuck",
				Type:       "webhook",
			},
		},
	}
}

var Common = NewStoreTest()
