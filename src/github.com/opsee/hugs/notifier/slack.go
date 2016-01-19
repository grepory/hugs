package notifier

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/hoisie/mustache"
	"github.com/opsee/hugs/store"
	slacktmpl "github.com/opsee/notification-templates/dist/go/slack"
	"github.com/sirupsen/logrus"
)

type SlackSender struct {
	templates map[string]*mustache.Template
}

// Send notification to customer.  At this point we have done basic validation on notification and event
func (this SlackSender) Send(n *store.Notification, e Event) error {
	logrus.Info("Notifier: Requested slack notification.")

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

func NewSlackSender() SlackSender {
	// initialize check failing template
	failTemplate, err := mustache.ParseString(slacktmpl.CheckFailing)
	if err != nil {
		logrus.Warn("Notifier: Failed to get check failing template.")
	}

	// initialize check passing template
	passTemplate, err := mustache.ParseString(slacktmpl.CheckPassing)
	if err != nil {
		logrus.Warn("Notifier: Failed to get check passing template.")
	}

	templateMap := map[string]*mustache.Template{
		"check-failing": failTemplate,
		"check-passing": passTemplate,
	}

	return SlackSender{
		templates: templateMap,
	}
}
