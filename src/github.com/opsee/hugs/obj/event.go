package obj

import "github.com/opsee/hugs/util"

type Event struct {
	CheckID       string `json:"check_id" required:"true"`
	CheckName     string `json:"check_name" required:"true"`
	GroupID       string `json:"group_id" required:"true"`
	FirstResponse string `json:"first_response" required:"true"`
	InstanceCount int    `json:"instance_count" `
	FailCount     int    `json:"fail_count" required:"true"`
}

func (this *Event) Validate() error {
	validator := &util.Validator{}
	return validator.Validate(this)
}
