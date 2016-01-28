package store

import (
	"github.com/opsee/basic/com"
	"github.com/opsee/hugs/obj"
)

type Store interface {
	GetNotifications(*com.User) ([]*obj.Notification, error)
	GetNotificationsByCheckID(*com.User, string) ([]*obj.Notification, error)
	UnsafeGetNotificationsByCheckID(string) ([]*obj.Notification, error)
	PutNotification(*com.User, *obj.Notification) error
	UnsafePutNotification(*obj.Notification) error
	PutNotifications(*com.User, []*obj.Notification) error
	UpdateNotification(*com.User, *obj.Notification) error
	DeleteNotification(*com.User, *obj.Notification) error
	DeleteNotifications(*com.User, []*obj.Notification) error
	GetSlackOAuthResponse(*com.User) (*obj.SlackOAuthResponse, error)
	UpdateSlackOAuthResponse(*com.User) error
	PutSlackOAuthResponse(*com.User, *obj.SlackOAuthResponse) error
}
