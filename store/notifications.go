package store

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/opsee/basic/schema"
	"github.com/opsee/hugs/obj"
	log "github.com/opsee/logrus"
)

func (pg *Postgres) GetNotifications(user *schema.User, oldNotifications []*obj.Notification) ([]*obj.Notification, error) {
	notifications := []*obj.Notification{}
	for _, oldNotification := range oldNotifications {
		newNotification, err := pg.GetNotification(user, oldNotification.Id)
		if err != nil {
			log.WithError(err).Errorf("Failed to get notification %d, for customerId %s", oldNotification.Id, user.CustomerId)
		}
		notifications = append(notifications, newNotification)
	}

	return notifications, nil
}

func (pg *Postgres) GetNotification(user *schema.User, id int) (*obj.Notification, error) {
	notification := &obj.Notification{}
	err := pg.db.Get(notification, "SELECT * FROM notifications WHERE customer_id = $1 AND id = $2", user.CustomerId, id)
	return notification, err
}

func (pg *Postgres) GetDefaultNotifications(user *schema.User) ([]*obj.Notification, error) {
	var notifications []*obj.Notification
	err := pg.db.Select(&notifications, "SELECT type, value FROM default_notifications WHERE customer_id = $1 limit 100", user.CustomerId)
	return notifications, err
}

func (pg *Postgres) GetNotificationsByUser(user *schema.User) ([]*obj.Notification, error) {
	var notifications []*obj.Notification
	rows, err := pg.db.Queryx("SELECT * from notifications WHERE customer_id = $1", user.CustomerId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var notification obj.Notification
		err := rows.StructScan(&notification)
		if err != nil {
			log.WithError(err).Errorf("Couldn't scan notification for user %s", user.CustomerId)
			return nil, err
		}
		notifications = append(notifications, &notification)
	}

	return notifications, err
}

func (pg *Postgres) deleteNotification(x sqlx.Ext, notification *obj.Notification) error {
	_, err := sqlx.NamedExec(x, `delete from notifications where id = :id  AND customer_id = :customer_id`, notification)
	return err
}

// Deletes all notifications associated with the given notification's checkId
func (pg *Postgres) deleteNotificationsByCheckId(x sqlx.Ext, notification *obj.Notification) error {
	_, err := sqlx.NamedExec(x, `delete from notifications where check_id = :check_id  AND customer_id = :customer_id`, notification)
	return err
}

func (pg *Postgres) putNotification(x sqlx.Ext, notification *obj.Notification) error {
	_, err := sqlx.NamedExec(x,
		`INSERT INTO notifications (customer_id, user_id, check_id, value, type)
		VALUES (:customer_id, :user_id, :check_id, :value, :type)
		RETURNING id`, notification)
	return err
}

func (pg *Postgres) putDefaultNotification(x sqlx.Ext, notification *obj.Notification) error {
	_, err := sqlx.NamedExec(x,
		`INSERT INTO default_notifications (customer_id, value, type)
		VALUES (:customer_id, :value, :type)
		RETURNING id`, notification)
	return err
}

func (pg *Postgres) PutDefaultNotifications(user *schema.User, notifications []*obj.Notification) error {
	tx, err := pg.db.Beginx()
	if err != nil {
		return err
	}

	_, err = tx.NamedExec("delete from default_notifications where customer_id = :customer_id", user)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, notif := range notifications {
		err = pg.putDefaultNotification(tx, notif)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (pg *Postgres) PutNotifications(user *schema.User, notifications []*obj.Notification) error {
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

func (pg *Postgres) PutNotificationsMultiCheck(notificationsObjs []*obj.Notifications) error {
	tx, err := pg.db.Beginx()
	if err != nil {
		return err
	}

	for _, notificationsObj := range notificationsObjs {
		// delete all notifications associated with this checkId
		if len(notificationsObj.Notifications) > 0 {
			notification := notificationsObj.Notifications[0]
			err = pg.deleteNotificationsByCheckId(tx, notification)
			if err != nil {
				log.WithError(err).Errorf("Couldn't delete notifications for check %s for customerId %s", notification.CheckId, notification.CustomerId)
				if err := tx.Rollback(); err != nil {
					log.WithError(err).Error("Error rolling back transaction")
				}
				return fmt.Errorf("Couldn't delete notification.")
			}
		}

		// add all new notifications for a given check
		for _, notification := range notificationsObj.Notifications {
			err = pg.putNotification(tx, notification)
			if err != nil {
				log.WithError(err).Errorf("Couldn't put notification %d for customerId %s", notification.Id, notification.CustomerId)
				if err := tx.Rollback(); err != nil {
					log.WithError(err).Error("Error rolling back transaction")
				}
				return fmt.Errorf("Couldn't put notification.")
			}
		}
	}
	return tx.Commit()
}

func (pg *Postgres) DeleteNotifications(notifications []*obj.Notification) error {
	tx, err := pg.db.Beginx()
	if err != nil {
		return err
	}

	for _, notification := range notifications {
		err = pg.deleteNotification(tx, notification)
		if err != nil {
			log.WithError(err).Errorf("Couldn't delete notification %s for customerId %s", notification.Id, notification.CustomerId)
			if err := tx.Rollback(); err != nil {
				log.WithError(err).Error("Error rolling back transaction")
			}
			return fmt.Errorf("Couldn't delete notification.  Rolled back.")
		}
	}
	return tx.Commit()
}

func (pg *Postgres) GetNotificationsByCheckId(user *schema.User, checkId string) ([]*obj.Notification, error) {
	notifications := []*obj.Notification{}
	err := pg.db.Select(&notifications, "SELECT * from notifications WHERE check_id = $1 AND customer_id = $2", checkId, user.CustomerId)
	if err != nil {
		return nil, err
	}

	return notifications, nil
}

func (pg *Postgres) UnsafeGetNotificationsByCheckId(checkId string) ([]*obj.Notification, error) {
	notifications := []*obj.Notification{}
	err := pg.db.Select(&notifications, "SELECT * from notifications WHERE check_id = $1", checkId)
	if err != nil {
		return nil, err
	}

	return notifications, nil
}

func (pg *Postgres) DeleteNotification(user *schema.User, notification *obj.Notification) error {
	rows, err := pg.db.Queryx(`DELETE from notifications WHERE id=$1 AND customer_id=$2`, notification.Id, user.CustomerId)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
}

func (pg *Postgres) DeleteNotificationsByUser(user *schema.User) error {
	rows, err := pg.db.Queryx(`DELETE from notifications WHERE customer_id=$1`, user.CustomerId)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
}

func (pg *Postgres) DeleteNotificationsByCheckId(user *schema.User, checkId string) error {
	rows, err := pg.db.Queryx(`DELETE from notifications WHERE customer_id=$1 AND check_id=$2`, user.CustomerId, checkId)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
}
