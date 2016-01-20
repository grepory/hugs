package notifier

import (
	"github.com/keighl/mandrill"
	"github.com/opsee/hugs/store"
)

type EmailSender struct {
	opseeHost  string
	mailClient *mandrill.Client
}

func (es EmailSender) Send(n *store.Notification, e Event) error {
	templateName := "check-pass"
	if e.FailCount > 0 {
		templateName = "check-fail"
	}

	templateContent := map[string]interface{}{
		"check_id":       e.CheckID,
		"check_name":     e.CheckName,
		"group_id":       e.GroupID,
		"first_response": e.FirstResponse,
		"instance_count": e.InstanceCount,
		"fail_count":     e.FailCount,
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
