package store

import (
	"github.com/opsee/basic/schema"
	"github.com/opsee/hugs/obj"
)

type Store interface {
	GetNotifications(*schema.User) ([]*obj.Notification, error)
	GetNotificationsByCheckId(*schema.User, string) ([]*obj.Notification, error)
	UnsafeGetNotificationsByCheckId(string) ([]*obj.Notification, error)
	PutNotification(*schema.User, *obj.Notification) error
	PutNotifications(*schema.User, []*obj.Notification) error
	UpdateNotification(*schema.User, *obj.Notification) error
	DeleteNotification(*schema.User, *obj.Notification) error
	DeleteNotifications(*schema.User, []*obj.Notification) error
	GetSlackOAuthResponse(*schema.User) (*obj.SlackOAuthResponse, error)
	UpdateSlackOAuthResponse(*schema.User) error
	PutSlackOAuthResponse(*schema.User, *obj.SlackOAuthResponse) error
}
