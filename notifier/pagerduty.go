package notifier

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/hoisie/mustache"
	"github.com/opsee/basic/com"
	"github.com/opsee/hugs/obj"
	"github.com/opsee/hugs/store"
	slacktmpl "github.com/opsee/notification-templates/dist/go/slack"
)

type PagerDutySender struct {
	templates map[string]*mustache.Template
}

// Send notification to customer.  At this point we have done basic validation on notification and event
func (this PagerDutySender) Send(n *obj.Notification, e *obj.Event) error {
	result := e.Result

	templateKey := "check-passing"
	eventType := "resolve"
	if !result.Passing {
		eventType = "trigger"
		templateKey = "check-failing"
	}

	failingResponses := result.FailingResponses()

	if len(failingResponses) < 1 && !result.Passing {
		return errors.New("Received failing CheckResult with no failing responses.")
	}

	if _, ok := this.templates[templateKey]; !ok {
		return fmt.Errorf("Template key not found")
	}
	pdTemplate := this.templates[templateKey]

	serviceKey, err := this.getPagerDutyServiceKey(n)
	if err != nil {
		return err
	}

	templateContent := map[string]interface{}{
		"check_id":       result.CheckId,
		"check_name":     result.CheckName,
		"group_name":     result.Target.Id,
		"instance_count": len(result.Responses),
		"fail_count":     len(failingResponses),
	}

	contexts := []*obj.PagerDutyContext{}
	if e.Nocap != nil && e.Nocap.JSONUrl != "" {
		contexts = append(contexts, &obj.PagerDutyContext{
			Type: "link",
			// TODO(dan) remove the url from the fasdsdffl template
			Href: fmt.Sprintf("https://app.opsee.com/check/%s/%s/event?json=%s&utm_medium=pagerduty&utm_campaign=app", url.QueryEscape(e.Nocap.JSONUrl)),
			Text: "View complete check response.",
		})
	}

	postMessageRequest := &obj.PagerDutyRequest{
		ServiceKey:  serviceKey,
		IncidentKey: result.CheckId,
		EventType:   eventType,
	}

	if eventType == "trigger" {
		postMessageRequest.Description = pdTemplate.Render(templateContent)
		postMessageRequest.Details = templateContent
		postMessageRequest.Contexts = contexts
	}

	_, err = postMessageRequest.Do()
	return err
}

func (this PagerDutySender) getPagerDutyServiceKey(n *obj.Notification) (string, error) {
	s, err := store.NewPostgres()
	if err != nil {
		return "", err
	}

	oaResponse, err := s.GetPagerDutyOAuthResponse(&com.User{CustomerID: n.CustomerID})
	if err != nil {
		return "", err
	}
	if oaResponse.Enabled == false {
		return "", fmt.Errorf("integration_disabled")
	}

	return oaResponse.ServiceKey, nil
}

func NewPagerDutySender() (*PagerDutySender, error) {
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

	return &PagerDutySender{
		templates: templateMap,
	}, nil
}
