package results

import (
	"github.com/opsee/basic/schema"
)

// Store is used to store CheckResults and snapshots of checks with results.
type Store interface {
	GetResultByCheckId(bastionId, customerId string) (*schema.CheckResult, error)
	PutResult(result *schema.CheckResult) error
	GetCheckSnapshot(transitionId int64, checkId string) (*schema.Check, error)
	PutCheckSnapshot(transitionId int64, check *schema.Check) error
}
