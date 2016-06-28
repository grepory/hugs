package store

import (
	"testing"

	log "github.com/opsee/logrus"
)

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
	notifications, err := Common.DBStore.GetNotificationsByUser(Common.User)

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

func TestStoreGetNotificationsByCheckId(t *testing.T) {
	checkId := "00001"
	log.Info("TestStoreGetNotificationsByCheckId: Getting Common.Notifications from store for CheckId ", checkId)

	notifications, err := Common.DBStore.GetNotificationsByCheckId(Common.User, checkId)
	if err != nil {
		log.Error(err)
		t.FailNow()
	}
	if len(notifications) != len(Common.Notifications) {
		log.Error("TestStoreGetNotificationsByCheckId: Got ", len(notifications), ".")
		t.FailNow()
	}
	log.Info("TestStoreGetNotificationsByCheckId: PASS.")
}

func TestStoreDeleteNotifications(t *testing.T) {
	notifications, err := Common.DBStore.GetNotificationsByUser(Common.User)
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
	notifications, err = Common.DBStore.GetNotificationsByUser(Common.User)
	if err != nil {
		log.Error(err)
		t.FailNow()
	} else if len(notifications) != 0 {
		log.Error("TestStoreDeleteNotifications: Deleted 3 Notifications and Expect 0, Got ", len(notifications), ".")
		t.FailNow()
	}
	log.Info("TestStoreDeleteNotifications: PASS.")
}
