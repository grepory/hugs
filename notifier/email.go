package notifier

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/keighl/mandrill"
	"github.com/opsee/basic/schema"
	"github.com/opsee/cats/checks/results"
	"github.com/opsee/hugs/config"
	"github.com/opsee/hugs/obj"
	log "github.com/sirupsen/logrus"
)

type EmailSender struct {
	opseeHost   string
	mailClient  *mandrill.Client
	resultStore results.Store
}

func (es EmailSender) Send(n *obj.Notification, e *obj.Event) error {
	result := e.Result
	var (
		templateName string
		responses    []*schema.CheckResponse
	)

	if result.Passing {
		responses = result.Responses
	} else {
		responses = result.FailingResponses()
	}

	log.WithFields(log.Fields{"responses": responses}).Debug("Got responses.")

	// It's a possible error state that if the CheckResult.Passing field is false,
	// i.e. this is a failing event, that there are somehow no constituent failing
	// CheckResponse objects contained within the CheckResult. We cannot know _why_
	// these CheckResponse objects aren't failing. Because we cannot ordain the reason
	// for this error state, let us first err on the side of not bugging a customer.
	if len(responses) < 1 && !result.Passing {
		return errors.New("Received failing CheckResult with no failing responses.")
	}

	instances := []*schema.Target{}
	for _, resp := range responses {
		instances = append(instances, resp.Target)
	}
	log.WithFields(log.Fields{"instances": instances}).Info("Got instances.")

	responseJson, err := json.MarshalIndent(responses[0], "", "  ")
	if err != nil {
		return err
	}

	templateContent := map[string]interface{}{
		"check_id":       result.CheckId,
		"check_name":     result.CheckName,
		"group_id":       result.Target.Id,
		"group_name":     result.Target.Id,
		"first_response": string(responseJson),
		"instance_count": len(result.Responses),
		"instances":      instances,
		"fail_count":     result.FailingCount(),
		"opsee_host":     config.GetConfig().OpseeHost,
	}
	log.WithFields(log.Fields{"template_content": templateContent}).Debug("Build template content")

	if result.Target.Name != "" {
		templateContent["group_name"] = result.Target.Name
	}

	if e.Nocap != nil {
		nocap := e.Nocap
		templateContent["json_url"] = nocap.JSONUrl

		if result.Passing {
			templateName = "check-pass-json"
		} else {
			templateName = "check-fail-json"
		}
	} else {
		if result.Passing {
			templateName = "check-pass"
		} else {
			templateName = "check-fail"
		}
	}

	// use different tempalte for RDS instances
	if result.Target.Type == "dbinstance" {
		if result.Passing {
			templateName = "check-pass-rds"
		} else {
			templateName = "check-fail-rds"
		}
		templateContent["rds_db_name"] = result.Target.Id
	}

	// use different template for external hosts
	if result.Target.Type == "external_host" {
		if result.Passing {
			templateName = "check-pass-url"
		} else {
			templateName = "check-fail-url"
		}

		results, err := es.resultStore.GetResultsByCheckId(result.CheckId)
		if err != nil {
			return err
		}

		var (
			instanceCount = len(results)
			failCount     int
		)

		for _, r := range results {
			failCount += r.FailingCount()
		}

		// we have inconsistent results, so don't do anything
		if !result.Passing && failCount == 0 {
			return fmt.Errorf("Failing result, but fail count == 0")
		}

		templateContent["instance_count"] = instanceCount
		templateContent["fail_count"] = failCount
	}

	mergeVars := templateContent
	mergeVars["opsee_host"] = es.opseeHost
	message := &mandrill.Message{}
	message.AddRecipient(n.Value, n.Value, "to")
	message.Merge = true
	message.MergeLanguage = "handlebars"
	message.MergeVars = []*mandrill.RcptMergeVars{mandrill.MapToRecipientVars(n.Value, mergeVars)}

	log.Debug(message)

	_, err = es.mailClient.MessagesSendTemplate(message, templateName, templateContent)
	return err
}

func NewEmailSender(host string, mandrillKey string) (*EmailSender, error) {
	return &EmailSender{
		opseeHost:  host,
		mailClient: mandrill.ClientWithKey(mandrillKey),
		resultStore: &results.S3Store{
			S3Client:   dynamodb.New(session.New(&aws.Config{Region: aws.String("us-west-2")})),
			BucketName: "opsee-results-production"},
	}, nil
}
