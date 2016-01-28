package obj

import (
	_ "github.com/lib/pq"
	"github.com/opsee/hugs/util"
)

type Notifications struct {
	Notifications []*Notification `json:"notifications" db:"notifications"`
}

func (this *Notifications) Validate() error {
	return nil
}

type Notification struct {
	ID         int    `json:"id" db:"id" required:"true"`
	CustomerID string `json:"customer_id" db:"customer_id" required:"true"`
	UserID     int    `json:"user_id" db:"user_id" required:"true"`
	CheckID    string `json:"check_id" db:"check_id" required:"true"`
	Value      string `json:"value" db:"value" required:"true"`
	Type       string `json:"type" db:"type" required:"true"`
}

func (this *Notification) Validate() error {
	validator := &util.Validator{}
	return validator.Validate(this)
}

type Customer struct {
	ID string `json:"id" db:"id" required:"true"`
}

func (this *Customer) Validate() error {
	validator := &util.Validator{}
	return validator.Validate(this)
}
