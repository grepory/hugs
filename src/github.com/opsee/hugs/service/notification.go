package service

import (
	log "github.com/Sirupsen/logrus"
	"github.com/opsee/basic/com"
	"github.com/opsee/hugs/store"
)

// returns notifications by user.ID
func (s *service) GetNotifications(user *com.User) ([]*store.Notification, error) {
	notifications, err := s.db.GetNotifications(user)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{"customer_id": user.CustomerID}).Error("Error getting notifications from database")
	}

	return notifications, nil
}

func (s *service) PutNotifications(user *com.User, req *CheckNotifications) error {
	err := s.db.PutNotifications(user, req.Notifications)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{"customer_id": user.CustomerID, "Error": err.Error()}).Error("Error putting notification in database.")
	}

	return err
}

func (s *service) DeleteNotifications(user *com.User, req *CheckNotifications) error {
	err := s.db.DeleteNotifications(user, req.Notifications)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{"customer_id": user.CustomerID, "Error": err.Error()}).Error("Error deleting notification from database.")
	}

	return err
}

func (s *service) GetNotificationsByCheckID(user *com.User, checkID string) ([]*store.Notification, error) {
	notifications, err := s.db.GetNotificationsByCheckID(user, checkID)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{"customer_id": user.CustomerID, "Error": err.Error()}).Error("Error getting notifications from database by")
		return nil, err
	}

	return notifications, nil
}
