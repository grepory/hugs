package notifier

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/hoisie/mustache"
	"github.com/opsee/basic/com"
	"github.com/opsee/hugs/obj"
	"github.com/opsee/hugs/store"
	pdtmpl "github.com/opsee/notification-templates/dist/go/pagerduty"
	log "github.com/sirupsen/logrus"
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

	postMessageRequest := &obj.PagerDutyRequest{}
	switch eventType {

	case "resolve":
		templateContent := map[string]interface{}{
			"service_key":  serviceKey,
			"incident_key": result.CheckId,
		}

		err = json.Unmarshal([]byte(pdTemplate.Render(templateContent)), postMessageRequest)
		if err != nil {
			return err
		}

	case "trigger":
		templateContent := map[string]interface{}{
			"service_key": serviceKey,
			"check_name":  result.CheckName,
			"check_id":    result.CheckId,
			"group_name":  result.Target.Id,
			"opsee_host":  "app.opsee.com",
		}

		if e.Nocap != nil {
			if e.Nocap.JSONUrl != "" {
				templateContent["json_url"] = url.QueryEscape(e.Nocap.JSONUrl)
			} else {
				templateContent["json_url"] = "?"
			}
		} else {
			templateContent["json_url"] = "?"
		}

		log.Debug(string(pdTemplate.Render(templateContent)))
		err = json.Unmarshal([]byte(pdTemplate.Render(templateContent)), postMessageRequest)
		if err != nil {
			return err
		}
		resultJson, _ := json.Marshal(result)
		postMessageRequest.Details = string(resultJson)

	}

	response, err := postMessageRequest.Do()
	log.Debug(response)
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
	failTemplate, err := mustache.ParseString(pdtmpl.CheckFailing)
	if err != nil {
		return nil, err
	}

	// initialize check passing template
	passTemplate, err := mustache.ParseString(pdtmpl.CheckPassing)
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
