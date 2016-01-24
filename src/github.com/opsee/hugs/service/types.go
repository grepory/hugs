package service

import (
	"errors"

	"github.com/opsee/hugs/store"
)

type CheckNotifications struct {
	Notifications []*store.Notification `json:"notifications"`
}

func (c *CheckNotifications) Validate() error {
	if len(c.Notifications) < 1 {
		return errors.New("There must be at least one notification.")
	}

	return nil
}

type SlackChannel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type SlackChannels struct {
	Channels []*SlackChannel `json:"channels"`
}

func (c *SlackChannels) Validate() error {
	if len(c.Channels) < 0 {
		return errors.New("You can't have a negative nubmer of channels!")
	}

	return nil
}
