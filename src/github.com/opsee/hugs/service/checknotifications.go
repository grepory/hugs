package service

import (
	"errors"

	"github.com/opsee/hugs/store"
)

type CheckNotifications struct {
	Notifications []*store.Notification `json:"assertions"`
}

func (c *CheckNotifications) Validate() error {
	if len(c.Notifications) < 1 {
		return errors.New("There must be at least one notification.")
	}

	return nil
}
