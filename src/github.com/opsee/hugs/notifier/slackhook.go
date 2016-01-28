package notifier

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/hoisie/mustache"
	"github.com/opsee/hugs/obj"
	slacktmpl "github.com/opsee/notification-templates/dist/go/slack"
)

type SlackHookSender struct {
	templates map[string]*mustache.Template
}

// Send notification to customer.  At this point we have done basic validation on notification and event
func (this *SlackHookSender) Send(n *obj.Notification, e Event) error {

	templateKey := "check-passing"
	if e.FailCount > 0 {
		templateKey = "check-failing"
	}

	if template, ok := this.templates[templateKey]; ok {
		templateContent := map[string]interface{}{
			"check_id":       e.CheckID,
			"check_name":     e.CheckName,
			"group_id":       e.GroupID,
			"first_response": e.FirstResponse,
			"instance_count": e.InstanceCount,
			"fail_count":     e.FailCount,
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
