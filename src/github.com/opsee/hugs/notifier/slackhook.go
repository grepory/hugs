package notifier

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"

	"github.com/hoisie/mustache"
	"github.com/opsee/hugs/checker"
	"github.com/opsee/hugs/obj"
	slacktmpl "github.com/opsee/notification-templates/dist/go/slack"
)

type SlackHookSender struct {
	templates map[string]*mustache.Template
}

// Send notification to customer.  At this point we have done basic validation on notification and event
func (this *SlackHookSender) Send(n *obj.Notification, e *checker.CheckResult) error {
	templateKey := "check-passing"
	if !e.Passing {
		templateKey = "check-failing"
	}

	// Bleh. This is copypasta from email.go
	// TODO(greg): When we move to a generic model, we can figure out a way
	// to centralize all of this logic so that senders can finally be dumb.
	failCount := 0
	var firstFailingResponse *checker.CheckResponse

	for i := 0; i < len(e.Responses); i++ {
		if !e.Responses[i].Passing {
			failCount += 1
			if firstFailingResponse == nil {
				firstFailingResponse = e.Responses[i]
			}
		}
	}

	// It's a possible error state that if the CheckResult.Passing field is false,
	// i.e. this is a failing event, that there are somehow no constituent failing
	// CheckResponse objects contained within the CheckResult. We cannot know _why_
	// these CheckResponse objects aren't failing. Because we cannot ordain the reason
	// for this error state, let us first err on the side of not bugging a customer.
	if firstFailingResponse == nil && !e.Passing {
		return errors.New("Received failing CheckResult with no failing responses.")
	}

	if template, ok := this.templates[templateKey]; ok {
		templateContent := map[string]interface{}{
			"check_id":       e.CheckId,
			"check_name":     e.CheckName,
			"group_id":       e.Target.Id,
			"first_response": e.Responses[0],
			"instance_count": len(e.Responses),
			"fail_count":     failCount,
		}

		body := bytes.NewBufferString(template.Render(templateContent))
		resp, err := http.Post(n.Value, "application/json", body)
		if err != nil {
			return err
		}

		defer resp.Body.Close()
	} else {
		return fmt.Errorf("Slack Notifier: Could not find appropriate Slack template for event type (", templateKey, ")")
	}

	return nil
}

func NewSlackHookSender() (*SlackHookSender, error) {
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

	return &SlackHookSender{
		templates: templateMap,
	}, nil
}
