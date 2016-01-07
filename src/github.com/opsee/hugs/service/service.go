package service

import (
	"errors"

	"github.com/opsee/basic/com"
	"github.com/opsee/basic/tp"
	"github.com/opsee/hugs/config"
	"github.com/opsee/hugs/store"
)

var (
	errUnauthorized = errors.New("unauthorized.")
	errUnknown      = errors.New("unknown error.")
)

type Service interface {
	GetNotifications(*com.User)
	PutNotifications(*com.User, *CheckNotifications)
	DeleteNotifications(*com.User, *CheckNotifications)
}

type service struct {
	db     *store.Postgres
	router *tp.Router
	config *config.Config
}

func NewService(db *store.Postgres, cfg *config.Config) *service {
	return &service{
		db:     db,
		config: cfg,
	}
}

type CheckNotifications struct {
	Notifications []*store.Notification `json:"assertions"`
}

func (c *CheckNotifications) Validate() error {
	if len(c.Notifications) < 1 {
		return errors.New("There must be at least one notification.")
	}

	return nil
}
