package notifier

import (
	"fmt"

	"github.com/opsee/hugs/config"
	"github.com/opsee/hugs/store"
)

// Interface implemented by everything that wants to send notifications
type Sender interface {
	Send(n *store.Notification, e Event) error
}

// A notifier is a map of Senders
// NOTE: This is clearly not threadsafe.  Use multiple notifiers per worker and let worker handle concurrency.
type Notifier struct {
	Senders map[string]Sender
}

// A collection of Senders, utilized by Workers to send notifications, return map of sender initialization errors to Warn on
func NewNotifier() (*Notifier, map[string]error) {
	errMap := make(map[string]error)
	notifier := &Notifier{
		Senders: map[string]Sender{},
	}

	// try add slack sender
	slackSender, err := NewSlackSender()
	if err != nil {
		errMap["slack"] = err
	} else {
		notifier.addSender("slack", slackSender)
	}

	// try add email sender
	emailSender, err := NewEmailSender(config.GetConfig().OpseeHost, config.GetConfig().MandrillApiKey)
	if err != nil {
		errMap["email"] = err
	} else {
		notifier.addSender("email", emailSender)
	}
	return notifier, errMap
}

func (n Notifier) addSender(key string, sender Sender) {
	n.Senders[key] = sender
}

func (n Notifier) getSender(t string) (Sender, error) {
	if sender, ok := n.Senders[t]; ok {
		return sender, nil
	}
	return nil, fmt.Errorf("Notifier doesn't have a Sender for that notification type")
}

// A Send should require only a notification (userID, type, value) and Event (check info)
func (n Notifier) Send(notification *store.Notification, event Event) error {
	sender, err := n.getSender(notification.Type)
	if err == nil {
		return sender.Send(notification, event)
	}
	return err
}
