package store

import (
	"os"
	"testing"

	"github.com/opsee/basic/com"
	log "github.com/sirupsen/logrus"
)

type StoreTest struct {
	DBStore       *Postgres
	Notifications []*Notification
	User          *com.User
}

func NewStoreTest() *StoreTest {
	log.Info("Connecting to local test store")
	db, err := NewPostgres(os.Getenv("HUGS_POSTGRES_CONN"))
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
		Notifications: []*Notification{
			&Notification{
				ID:         0,
				CustomerID: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
				UserID:     13,
				CheckID:    "00000",
				Value:      "off",
				Type:       "slack_bot",
			},
			&Notification{
				ID:         1,
				CustomerID: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
				UserID:     13,
				CheckID:    "00000",
				Value:      "you",
				Type:       "email",
			},
			&Notification{
				ID:         2,
				CustomerID: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
				UserID:     13,
				CheckID:    "00001",
				Value:      "fuck",
				Type:       "slack_hook",
			},
		},
	}
}

var Common = NewStoreTest()

func TestStorePutNotifications(t *testing.T) {
	for i, _ := range Common.Notifications {
		log.Info("TestStorePutNotifications: Adding Common.Notifications[", i, "] To Store.")
		if err := Common.DBStore.PutNotification(Common.User, Common.Notifications[i]); err != nil {
			log.Error(err)
			t.FailNow()
		}
	}
	log.Info("TestStorePutNotifications: PASS.")
}

func TestStoreGetNotifications(t *testing.T) {
	log.Info("TestStoreGetNotifications: Getting Common.Notifications from store")
	if notifications, err := Common.DBStore.GetNotifications(Common.User); err != nil {
		log.Error(err)
		t.FailNow()
	} else if len(notifications) != 3 {
		log.Error("TestStoreGetNotifications: Inserted 3 Notifications, Got ", len(notifications), ".")
		t.FailNow()
	}

	log.Info("TestStoreGetNotifications: PASS.")
}

func TestStoreGetNotificationsByCheckID(t *testing.T) {
	checkID := "00000"
	log.Info("TestStoreGetNotificationsByCheckID: Getting Common.Notifications from store for CheckID", checkID)
	if notifications, err := Common.DBStore.GetNotificationsByCheckID(Common.User, checkID); err != nil {
		log.Error(err)
		t.FailNow()
	} else if len(notifications) != 2 {
		log.Error("TestStoreGetNotificationsByCheckID: Deleted 3 Notifications and Expect 0, Got ", len(notifications), ".")
		t.FailNow()
	}
	log.Info("TestStoreGetNotificationsByCheckID: PASS.")
}

func TestStoreUpdateNotification(t *testing.T) {
	checkID := "11111"
	log.Info("TestStoreUpdateNotification: Getting Common.Notifications from store for CheckID", checkID)
	notifications, err := Common.DBStore.GetNotifications(Common.User)
	if err != nil {
		log.Error(err)
		t.FailNow()
	}
	for i, _ := range notifications {
		log.Info("TestStoreUpdateNotification: Update notifications[", i, "] From Store (Set notifications[", i, "].CheckID to \"11111\").")
		notifications[i].CheckID = checkID
		if err := Common.DBStore.UpdateNotification(Common.User, notifications[i]); err != nil {
			log.Error(err)
			t.FailNow()
		}
	}
	log.Info("TestStoreUpdateNotification: Validating Changes. Fetching Notifications.")
	notifications, err = Common.DBStore.GetNotifications(Common.User)
	if err != nil {
		log.Error(err)
		t.FailNow()
	}
	for i, _ := range notifications {
		if notifications[i].CheckID != checkID {
			log.Error(err)
			t.FailNow()
		}
		log.Info("TestStoreUpdateNotification: notifications[", i, "].CheckID  was updated successfully.")
	}
}

func TestStoreDeleteNotifications(t *testing.T) {
	notifications, err := Common.DBStore.GetNotifications(Common.User)
	if err != nil {
		log.Error(err)
		t.FailNow()
	}
	for i, _ := range notifications {
		log.Info("TestStoreDeleteNotifications: Delete Common.Notifications[", i, "] From Store.")
		if err := Common.DBStore.DeleteNotification(Common.User, notifications[i]); err != nil {
			log.Error(err)
			t.FailNow()
		}
	}
	notifications, err = Common.DBStore.GetNotifications(Common.User)
	if err != nil {
		log.Error(err)
		t.FailNow()
	} else if len(notifications) != 0 {
		log.Error("TestStoreDeleteNotifications: Deleted 3 Notifications and Expect 0, Got ", len(notifications), ".")
		t.FailNow()
	}
	log.Info("TestStoreDeleteNotifications: PASS.")
}
