package store

import (
	"encoding/json"

	"github.com/jmoiron/sqlx/types"
	"github.com/opsee/basic/schema"
	"github.com/opsee/hugs/obj"
	log "github.com/opsee/logrus"
)

func (pg *Postgres) DeleteSlackOAuthResponsesByUser(user *schema.User) error {
	rows, err := pg.db.Queryx(`DELETE from slack_oauth_responses WHERE customer_id=$1`, user.CustomerId)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
}

func (pg *Postgres) PutSlackOAuthResponse(user *schema.User, s *obj.SlackOAuthResponse) error {
	datjson, err := json.Marshal(s)
	if err != nil {
		return err
	}

	err = pg.DeleteSlackOAuthResponsesByUser(user)
	if err != nil {
		return err
	}

	wrapper := obj.SlackOAuthResponseDBWrapper{
		CustomerId: user.CustomerId,
		Data:       types.JSONText(string(datjson)),
	}

	_, err = pg.db.NamedExec("INSERT INTO slack_oauth_responses (customer_id, data) VALUES (:customer_id, :data)", wrapper)
	return err
}

func (pg *Postgres) GetSlackOAuthResponse(user *schema.User) (*obj.SlackOAuthResponse, error) {
	oaResponses, err := pg.GetSlackOAuthResponses(user)
	if err != nil {
		return nil, err
	}

	if len(oaResponses) > 0 {
		return oaResponses[0], nil
	}

	return nil, nil
}

func (pg *Postgres) GetSlackOAuthResponses(user *schema.User) ([]*obj.SlackOAuthResponse, error) {
	oaResponses := []*obj.SlackOAuthResponse{}
	rows, err := pg.db.Queryx("SELECT data from slack_oauth_responses WHERE customer_id = $1", user.CustomerId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var wrappedOAResponse obj.SlackOAuthResponseDBWrapper
		err := rows.StructScan(&wrappedOAResponse)
		if err != nil {
			log.Fatalln(err)
		}

		oaResponse := obj.SlackOAuthResponse{}
		err = wrappedOAResponse.Data.Unmarshal(&oaResponse)
		if err != nil {
			continue
		}

		oaResponses = append(oaResponses, &oaResponse)
	}

	return oaResponses, err
}

func (pg *Postgres) UpdateSlackOAuthResponse(user *schema.User, s *obj.SlackOAuthResponse) error {
	datjson, err := json.Marshal(s)
	if err != nil {
		return err
	}
	data := types.JSONText(string(datjson))
	rows, err := pg.db.Queryx(`UPDATE slack_oauth_responses SET data=$1 where customer_id=$2`, data, user.CustomerId)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
}
