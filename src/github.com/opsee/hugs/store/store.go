package store

import (
	"github.com/opsee/basic/com"
	"github.com/opsee/hugs/apiutils"
)

type Store interface {
	GetNotifications(*com.User) ([]*Notification, error)
	GetNotificationsByCheckID(*com.User, string) ([]*Notification, error)
	UnsafeGetNotificationsByCheckID(string) ([]*Notification, error)
	PutNotification(*com.User, *Notification) error
	UnsafePutNotification(*Notification) error
	PutNotifications(*com.User, []*Notification) error
	UpdateNotification(*com.User, *Notification) error
	DeleteNotification(*com.User, *Notification) error
	DeleteNotifications(*com.User, []*Notification) error
	GetSlackOAuthResponse(*com.User) (*apiutils.SlackOAuthResponse, error)
	UpdateSlackOAuthResponse(*com.User) error
	PutSlackOAuthResponse(*com.User, *apiutils.SlackOAuthResponse) error
}
