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

func (es EmailSender) Send(n *obj.Notification, e *checker.CheckResult) error {
	templateName := "check-pass"
	if !e.Passing {
		templateName = "check-fail"
	}

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

	templateContent := map[string]interface{}{
		"check_id":       e.CheckId,
		"check_name":     e.CheckName,
		"group_id":       e.Target.Id,
		"first_response": firstFailingResponse,
		"instance_count": len(e.Responses),
		"fail_count":     failCount,
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
