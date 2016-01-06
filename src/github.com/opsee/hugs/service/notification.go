package service

import (
	"github.com/opsee/basic/com"
)

func (s *service) ListNotifications(user *com.User) ([]*store.Notification, error) {
	notifications, err := s.db.GetNotifications(user)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{"customer_id": user.CustomerID}).Errorf("Error getting notifications from database")
	}

	return notifications, nil
}

func (s *service) PutNotification(user *com.User, notification *store.Notification) error {
	err := s.db.PutNotification(notification)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{"customer_id": user.CustomerID, "notification_id": notification.ID}).Errorf("Error putting notification in database.")
	}

	return err
}

func (s *service) UpdateNotification(user *com.User, notification *store.Notification) error {
	err := s.db.UpdateNotification(notification)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{"customer_id": user.CustomerID, "notification_id": notification.ID}).Errorf("Error updating notification in database.")
	}

	return err
}

func (s *service) DeleteNotification(user *com.Userk, notification *store.Notification) error {
	err := s.db.DeleteNotification(notification)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{"customer_id": user.CustomerID, "notification_id": notification.ID}).Errorf("Error deleting notification from database.")
	}

	return err
}
