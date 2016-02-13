package obj

import (
	"github.com/opsee/hugs/checker"
	"github.com/opsee/hugs/util"
	log "github.com/sirupsen/logrus"
)

type NocapResponse struct {
	Images  map[string]string `json:"image_urls"`
	JSONUrl string            `json:"json_url"`
}

type Event struct {
	Result *checker.CheckResult
	Nocap  *NocapResponse
	Test   bool
}

func (this *Event) Validate() error {
	validator := &util.Validator{}
	return validator.Validate(this)
}

func GenerateTestEvent() *Event {

	httpResponse := &checker.HttpResponse{
		Code: 200,
		Body: "test",
		Host: "a host",
	}

	responseAny, err := checker.MarshalAny(httpResponse)
	if err != nil {
		log.WithFields(log.Fields{"service": "GenerateTestEvent", "error": err}).Error("Error marshalling HttpResponse.")
	}

	checkResult := &checker.CheckResult{
		CheckId:   "00002",
		CheckName: `Test Check`,
		Target: &checker.Target{
			Id: `Test Target`,
		},
		Responses: []*checker.CheckResponse{
			&checker.CheckResponse{
				Target: &checker.Target{
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
