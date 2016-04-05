package obj

import (
	"github.com/opsee/basic/schema"
	"github.com/opsee/hugs/util"
	log "github.com/sirupsen/logrus"
)

type NocapResponse struct {
	Images  map[string]string `json:"image_urls"`
	JSONUrl string            `json:"json_url"`
}

type Event struct {
	Result *schema.CheckResult
	Nocap  *NocapResponse
	Test   bool
}

func (this *Event) Validate() error {
	validator := &util.Validator{}
	return validator.Validate(this)
}

func GenerateTestEvent() *Event {

	httpResponse := &schema.HttpResponse{
		Code: 200,
		Body: "test",
		Host: "a host",
	}

	responseAny, err := schema.MarshalAny(httpResponse)
	if err != nil {
		log.WithFields(log.Fields{"service": "GenerateTestEvent", "error": err}).Error("Error marshalling HttpResponse.")
	}

	checkResult := &schema.CheckResult{
		CheckId:   "00002",
		CheckName: `Test Check`,
		Target: &schema.Target{
			Id: `Test Target`,
		},
		Responses: []*schema.CheckResponse{
			&schema.CheckResponse{
				Target: &schema.Target{
					Id: "test-target",
				},
				Response: responseAny,
				Passing:  true,
			},
		},
		Passing: true,
		Version: 1,
	}
	event := &Event{
		Result: checkResult,
		Nocap: &NocapResponse{
			Images: map[string]string{
				"default": "https://opsee-notificaption-images.s3.amazonaws.com/dGhlIHJhcmVzdCBwZXBl_1454622727136_800.png",
			},
			JSONUrl: "https://opsee-notificaption-images.s3.amazonaws.com/dGhlIHJhcmVzdCBwZXBl_1454621842230.json",
		},
		Test: true,
	}
	return event
}
