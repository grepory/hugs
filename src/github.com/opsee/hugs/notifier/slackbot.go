package notifier

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/bluele/slack"
	"github.com/opsee/hugs/obj"
)

type SlackBotSender struct {
}

// Send notification to customer.  At this point we have done basic validation on notification and event
func (this SlackBotSender) Send(n *obj.Notification, e *obj.Event) error {
	result := e.Result

	state := "passing"
	if !result.Passing {
		state = "failing"
	}
	msg := fmt.Sprintf(`Check "%s" is *%s*.`, result.CheckName, state)

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

func (this SlackBotSender) getSlackToken(n *obj.Notification) (*string, error) {
	//TODO get tokn from vape usin' that tokn and that customer id
	return aws.String("xoxb-10894110067-HaeaISFpxwILScVehYTRVzZ1"), nil
}

func NewSlackBotSender() (*SlackBotSender, error) {
	return &SlackBotSender{}, nil
}
