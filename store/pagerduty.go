package store

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"

	"github.com/jmoiron/sqlx/types"
	"github.com/opsee/basic/schema"
	"github.com/opsee/hugs/obj"
)

func (pg *Postgres) GetPagerDutyOAuthResponse(user *schema.User) (*obj.PagerDutyOAuthResponse, error) {
	oaResponses, err := pg.GetPagerDutyOAuthResponses(user)
	if err != nil {
		return nil, err
	}

	if len(oaResponses) > 0 {
		return oaResponses[0], nil
	}

	return nil, nil
}

func (pg *Postgres) GetPagerDutyOAuthResponses(user *schema.User) ([]*obj.PagerDutyOAuthResponse, error) {
	oaResponses := []*obj.PagerDutyOAuthResponse{}
	rows, err := pg.db.Queryx("SELECT data from pagerduty_oauth_responses WHERE customer_id = $1", user.CustomerId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var wrappedOAResponse obj.PagerDutyOAuthResponseDBWrapper
		err := rows.StructScan(&wrappedOAResponse)
		if err != nil {
			log.Fatalln(err)
		}

		oaResponse := obj.PagerDutyOAuthResponse{}
		err = wrappedOAResponse.Data.Unmarshal(&oaResponse)
		if err != nil {
			continue
		}

		oaResponses = append(oaResponses, &oaResponse)
	}

	return oaResponses, err
}

func (pg *Postgres) UpdatePagerDutyOAuthResponse(user *schema.User, s *obj.PagerDutyOAuthResponse) error {
	datjson, err := json.Marshal(s)
	if err != nil {
		return err
	}
	data := types.JSONText(string(datjson))
	rows, err := pg.db.Queryx(`UPDATE pagerduty_oauth_responses SET data=$1 where customer_id=$2`, data, user.CustomerId)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
}

func (pg *Postgres) PutPagerDutyOAuthResponse(user *schema.User, s *obj.PagerDutyOAuthResponse) error {
	datjson, err := json.Marshal(s)
	if err != nil {
		return err
	}

	err = pg.DeletePagerDutyOAuthResponsesByUser(user)
	if err != nil {
		return err
	}

	wrapper := obj.PagerDutyOAuthResponseDBWrapper{
		CustomerId: user.CustomerId,
		Data:       types.JSONText(string(datjson)),
	}

	_, err = pg.db.NamedExec("INSERT INTO pagerduty_oauth_responses (customer_id, data) VALUES (:customer_id, :data)", wrapper)
	return err
}

func (pg *Postgres) DeletePagerDutyOAuthResponsesByUser(user *schema.User) error {
	rows, err := pg.db.Queryx(`DELETE from pagerduty_oauth_responses WHERE customer_id=$1`, user.CustomerId)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
}
