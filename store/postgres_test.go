package store

import (
	"testing"

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

func TestStorePutNotifications(t *testing.T) {
	Common.DBStore.DeleteNotificationsByUser(Common.User)
	log.Info("TestStorePutNotifications: Adding ", len(Common.Notifications), " To Store.")
	if err := Common.DBStore.PutNotifications(Common.User, Common.Notifications); err != nil {
		log.Error(err)
		t.FailNow()
	}
	log.Info("TestStorePutNotifications: PASS.")
}

func TestStoreGetNotifications(t *testing.T) {
	log.Info("TestStoreGetNotifications: Getting Common.Notifications from store")
	notifications, err := Common.DBStore.GetNotifications(Common.User)

	if err != nil {
		log.Error(err)
		t.FailNow()
	}

	if len(notifications) != len(Common.Notifications) {
		log.Error("TestStoreGetNotifications: Got ", len(notifications), ".")
		t.FailNow()
	}

	log.Info("TestStoreGetNotifications: PASS.")
}

func TestStoreGetNotificationsByCheckID(t *testing.T) {
	checkID := "00001"
	log.Info("TestStoreGetNotificationsByCheckID: Getting Common.Notifications from store for CheckID ", checkID)

	notifications, err := Common.DBStore.GetNotificationsByCheckID(Common.User, checkID)
	if err != nil {
		log.Error(err)
		t.FailNow()
	}
	if len(notifications) != len(Common.Notifications) {
		log.Error("TestStoreGetNotificationsByCheckID: Got ", len(notifications), ".")
		t.FailNow()
	}
	log.Info("TestStoreGetNotificationsByCheckID: PASS.")
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

func TestStorePutSlackOAuthResponse(t *testing.T) {
	slackOAuthResponse := &obj.SlackOAuthResponse{
		AccessToken: "test",
		Scope:       "test",
		TeamName:    "test",
		TeamID:      "test",
		IncomingWebhook: &obj.SlackIncomingWebhook{
			URL:              "test",
			Channel:          "test",
			ConfigurationURL: "test",
		},
		Bot: &obj.SlackBotCreds{
			BotUserID:      "test",
			BotAccessToken: "test",
		},
	}

	err := Common.DBStore.PutSlackOAuthResponse(Common.User, slackOAuthResponse)
	if err != nil {
		log.Error(err)
		t.FailNow()
	}
}

func TestStoreUpdateSlackOAuthResponse(t *testing.T) {
	slackOAuthResponse := &obj.SlackOAuthResponse{
		AccessToken: "test",
		Scope:       "test",
		TeamName:    "feck",
		TeamID:      "test",
		IncomingWebhook: &obj.SlackIncomingWebhook{
			URL:              "test",
			Channel:          "test",
			ConfigurationURL: "test",
		},
		Bot: &obj.SlackBotCreds{
			BotUserID:      "test",
			BotAccessToken: "test",
		},
	}

	err := Common.DBStore.UpdateSlackOAuthResponse(Common.User, slackOAuthResponse)
	if err != nil {
		log.Error(err)
		t.FailNow()
	}
}

func TestStoreGetSlackOAuthResponse(t *testing.T) {
	response, err := Common.DBStore.GetSlackOAuthResponse(Common.User)
	if err != nil {
		log.Error(err)
		t.FailNow()
	}
	log.Info("Got OAuthResponse: ", response)
}

func TestStoreGetSlackOAuthResponses(t *testing.T) {
	responses, err := Common.DBStore.GetSlackOAuthResponses(Common.User)
	if err != nil {
		log.Error(err)
		t.FailNow()
	}

	if len(responses) == 0 {
		t.FailNow()
	}

	log.Info("Got OAuthResponse: ", responses[0])
}
