package notifier

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hoisie/mustache"
	"github.com/opsee/basic/com"
	"github.com/opsee/hugs/config"
	"github.com/opsee/hugs/obj"
	"github.com/opsee/hugs/store"
	slacktmpl "github.com/opsee/notification-templates/dist/go/slack"
	log "github.com/sirupsen/logrus"
)

type SlackBotSender struct {
	templates map[string]*mustache.Template
}

// Send notification to customer.  At this point we have done basic validation on notification and event
func (this SlackBotSender) Send(n *obj.Notification, e *obj.Event) error {
	result := e.Result

	templateKey := "check-passing"
	if !result.Passing {
		templateKey = "check-failing"
	}

	// Bleh. This is copypasta from email.go
	// TODO(greg): When we move to a generic model, we can figure out a way
	// to centralize all of this logic so that senders can finally be dumb.
	failingResponses := result.FailingResponses()

	// It's a possible error state that if the CheckResult.Passing field is false,
	// i.e. this is a failing event, that there are somehow no constituent failing
	// CheckResponse objects contained within the CheckResult. We cannot know _why_
	// these CheckResponse objects aren't failing. Because we cannot ordain the reason
	// for this error state, let us first err on the side of not bugging a customer.
	if len(failingResponses) < 1 && !result.Passing {
		return errors.New("Received failing CheckResult with no failing responses.")
	}

	if slackTemplate, ok := this.templates[templateKey]; ok {
		token, err := this.getSlackToken(n)
		if err != nil {
			return err
		}

		templateContent := map[string]interface{}{
			"check_id":       result.CheckId,
			"check_name":     result.CheckName,
			"group_name":     result.Target.Id,
			"first_response": failingResponses[0],
			"instance_count": len(result.Responses),
			"fail_count":     len(failingResponses),
			"token":          token,
			"channel":        n.Value,
		}

		postMessageRequest := &obj.SlackPostChatMessageRequest{}
		log.Debug(string(slackTemplate.Render(templateContent)))
		err = json.Unmarshal([]byte(slackTemplate.Render(templateContent)), postMessageRequest)
		if err != nil {
			return err
		}

		slackPostMessageResponse, err := postMessageRequest.Do("https://slack.com/api/chat.postMessage")
		if err != nil {
			log.WithFields(log.Fields{"slackbot": "Send", "error": err}).Error("Error sending notification to slack.")
			return err
		}
		if slackPostMessageResponse.OK != true {
			return fmt.Errorf(slackPostMessageResponse.Error)
		}
	}

	return nil
}

func (this SlackBotSender) getSlackToken(n *obj.Notification) (string, error) {
	s, err := store.NewPostgres(config.GetConfig().PostgresConn)
	if err != nil {
		return "", err
	}

	oaResponse, err := s.GetSlackOAuthResponse(&com.User{CustomerID: n.CustomerID})
	if err != nil {
		return "", err
	}

	// if for whatever reason we don't have a bot
	if oaResponse.Bot == nil {
		log.WithFields(log.Fields{"slackbot": "getSlackToken"}).Error("User does not have a bot token associated with this slack integration.")
		return "", fmt.Errorf("integration_inactive")

	}

	return oaResponse.Bot.BotAccessToken, nil
}

func NewSlackBotSender() (*SlackBotSender, error) {
	// initialize check failing template
	failTemplate, err := mustache.ParseString(slacktmpl.CheckFailing)
	if err != nil {
		return nil, err
	}

	// initialize check passing template
	passTemplate, err := mustache.ParseString(slacktmpl.CheckPassing)
	if err != nil {
		return nil, err
	}

	templateMap := map[string]*mustache.Template{
		"check-failing": failTemplate,
		"check-passing": passTemplate,
	}

	return &SlackBotSender{
		templates: templateMap,
	}, nil
}
