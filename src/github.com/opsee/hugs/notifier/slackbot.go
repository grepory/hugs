package notifier

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/bluele/slack"
	"github.com/opsee/hugs/store"
)

type SlackBotSender struct {
}

// Send notification to customer.  At this point we have done basic validation on notification and event
func (this SlackBotSender) Send(n *store.Notification, e Event) error {
	passing := "passing"
	if e.FailCount > 0 {
		passing = "failing"
	}
	msg := fmt.Sprintf(`Check "%s" is *%s*.`, e.CheckName, passing)

	token, err := this.getSlackToken(n)
	if err != nil {
		return err
	}

	api := slack.New(*token)
	err = api.ChatPostMessage(n.Value, msg, nil)
	if err != nil {
		return err
	}

	return nil
}

func (this SlackBotSender) getSlackToken(n *store.Notification) (*string, error) {
	//TODO get tokn from vape usin' that tokn and that customer id
	return aws.String("xoxb-10894110067-HaeaISFpxwILScVehYTRVzZ1"), nil
}

func NewSlackBotSender() (*SlackBotSender, error) {
	return &SlackBotSender{}, nil
}
