package notifier

type Event struct {
	CheckID       string `json:"check_id" validate:"min=1"`
	CheckName     string `json:"check_name" validate:"min=1"`
	GroupID       string `json:"group_id" validate:"min=1"`
	FirstResponse string `json:"first_response" validate:"min=0"`
	InstanceCount int    `json:"instance_count" validate:"min=0"`
	FailCount     int    `json:"fail_count" validate:"min=0"`
}

func (e Event) Validate() bool {
	return true
	/*
		if err := validator.Validate(e); err != nil {
			return true
		}
		return false
	*/
}
