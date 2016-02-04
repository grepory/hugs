package obj

import (
	"github.com/opsee/hugs/checker"
	"github.com/opsee/hugs/util"
)

type NocapResponse struct {
	Images  map[string]string `json:"images"`
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
	checkResult := &checker.CheckResult{
		CheckId:   "00002",
		CheckName: `You'll never believe what we found`,
		Target: &checker.Target{
			Id: `your AWS Environment`,
		},

		Responses: []*checker.CheckResponse{
			&checker.CheckResponse{
				Target: &checker.Target{
					Id: "Your whole aws",
				},
				Error:   "You won't believe what we found.",
				Passing: false,
			},
		},
		Version: 1,
	}
	event := &Event{
		Result: checkResult,
		Test:   true,
	}
	return event
}
