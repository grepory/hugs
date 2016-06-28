package store

import (
	"testing"

	"github.com/opsee/hugs/obj"
	log "github.com/opsee/logrus"
)

func TestStorePutPagerDutyOAuthResponse(t *testing.T) {
	pdOAuthResponse := &obj.PagerDutyOAuthResponse{
		Account:     "test",
		ServiceKey:  "test",
		ServiceName: "test",
	}

	err := Common.DBStore.PutPagerDutyOAuthResponse(Common.User, pdOAuthResponse)
	if err != nil {
		log.Error(err)
		t.FailNow()
	}
}
