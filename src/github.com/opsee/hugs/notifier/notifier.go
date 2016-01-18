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

// A collection of Senders, utilized by Workers to send notifications
func NewNotifier() *Notifier {
	notifier := &Notifier{
		Senders: map[string]Sender{},
	}

	// add all of our sender types, the keys correspond to the Type field in store.Notification
	notifier.addSender("email", NewEmailSender(config.GetConfig().OpseeHost, config.GetConfig().MandrillApiKey))
	notifier.addSender("slack", NewSlackSender())
	return notifier
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
