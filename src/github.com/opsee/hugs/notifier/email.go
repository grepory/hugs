package notifier

import (
	"errors"

	"github.com/keighl/mandrill"
	"github.com/opsee/hugs/checker"
	"github.com/opsee/hugs/obj"
)

type EmailSender struct {
	opseeHost  string
	mailClient *mandrill.Client
}

func (es EmailSender) Send(n *obj.Notification, e *obj.Event) error {
	result := e.Result
	var (
		templateName string
	)

	failingResponses := result.FailingResponses()

	// It's a possible error state that if the CheckResult.Passing field is false,
	// i.e. this is a failing event, that there are somehow no constituent failing
	// CheckResponse objects contained within the CheckResult. We cannot know _why_
	// these CheckResponse objects aren't failing. Because we cannot ordain the reason
	// for this error state, let us first err on the side of not bugging a customer.
	if len(failingResponses) < 1 && !result.Passing {
		return errors.New("Received failing CheckResult with no failing responses.")
	}
	failingInstances := []*checker.Target{}
	for _, resp := range failingResponses {
		failingInstances = append(failingInstances, resp.Target)
	}

	templateContent := map[string]interface{}{
		"check_id":       result.CheckId,
		"check_name":     result.CheckName,
		"group_id":       result.Target.Id,
		"group_name":     result.Target.Id,
		"first_response": failingResponses[0],
		"instance_count": len(result.Responses),
		"instances":      failingInstances,
		"fail_count":     len(failingResponses),
	}

	if result.Target.Name != "" {
		templateContent["group_name"] = result.Target.Name
	}

	if e.Nocap != nil {
		nocap := e.Nocap
		templateContent["json_url"] = nocap.JSONUrl
		// TODO(greg): The images will have sizes in nocap response soon.
		templateContent["img_400"] = nocap.Images["default"]
		templateContent["img_400"] = nocap.Images["default"]
		if result.Passing {
			templateName = "check-pass-image"
		} else {
			templateName = "check-fail-image"
		}
	} else {
		if result.Passing {
			templateName = "check-pass"
		} else {
			templateName = "check-fail"
		}
	}

	mergeVars := make(map[string]interface{})
	mergeVars["opsee_host"] = es.opseeHost
	message := &mandrill.Message{}
	message.AddRecipient(n.Value, n.Value, "to")
	message.Merge = true
	message.MergeLanguage = "handlebars"
	message.MergeVars = []*mandrill.RcptMergeVars{mandrill.MapToRecipientVars(n.Value, mergeVars)}

	_, err := es.mailClient.MessagesSendTemplate(message, templateName, templateContent)
	return err
}

func NewEmailSender(host string, mandrillKey string) (*EmailSender, error) {
	return &EmailSender{
		opseeHost:  host,
		mailClient: mandrill.ClientWithKey(mandrillKey),
	}, nil
}
