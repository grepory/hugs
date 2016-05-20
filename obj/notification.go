package obj

import (
	_ "github.com/lib/pq"
	"github.com/opsee/hugs/util"
)

type Notifications struct {
	CheckId       string          `json:"check-id"`
	Notifications []*Notification `json:"notifications" db:"notifications"`
}

func (this *Notifications) Validate() error {
	validator := &util.Validator{}
	for _, notification := range this.Notifications {
		if err := validator.Validate(notification); err != nil {
			return err
		}
	}
	return nil
}

type Notification struct {
	Id         int    `json:"id" db:"id"`
	CustomerId string `json:"customer_id" db:"customer_id"`
	UserId     int    `json:"user_id" db:"user_id"`
	CheckId    string `json:"check_id" db:"check_id"`
	Value      string `json:"value" db:"value" required:"true"`
	Type       string `json:"type" db:"type" required:"true"`
}

func (this *Notification) Validate() error {
	validator := &util.Validator{}
	return validator.Validate(this)
}
