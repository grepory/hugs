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
}

func (this *Event) Validate() error {
	validator := &util.Validator{}
	return validator.Validate(this)
}
