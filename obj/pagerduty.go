package obj

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/jmoiron/sqlx/types"
	"github.com/opsee/hugs/util"
	log "github.com/sirupsen/logrus"
)

const PagerDutyIntegrationsAPIEndpoint = "https://events.pagerduty.com/generic/2010-04-15/create_event.json"

type PagerDutyContext struct {
	Type string `json:"type" required:"true"`
	Href string `json:"href" required:"true"`
	Text string `json:"text"`
}

type PagerDutyRequest struct {
	ServiceKey  string      `json:"service_key"`
	EventType   string      `json:"event_type"`
	Description string      `json:"description"`
	IncidentKey string      `json:"incident_key"`
	Client      string      `json:"client,omitempty"`
	ClientURL   string      `json:"client_url,omitempty"`
	Details     interface{} `json:"details,omitempty"`
}

func (pd *PagerDutyRequest) Validate() error {
	validator := &util.Validator{}
	return validator.Validate(pd)
}

func (pdr *PagerDutyRequest) Do() (*PagerDutyResponse, error) {
	if err := pdr.Validate(); err != nil {
		return nil, err
	}

	reqBody, err := json.Marshal(pdr)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(PagerDutyIntegrationsAPIEndpoint, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Info(string(bodyBytes))
	pdResponse := &PagerDutyResponse{}
	err = json.Unmarshal(bodyBytes, pdResponse)
	if err != nil {
		return nil, err
	}

	return pdResponse, nil
}

type PagerDutyBadRequest struct {
	Errors string `json:"errors"`
}

type PagerDutyResponse struct {
	Status      string `json:"status"`
	Message     string `json:"message"`
	IncidentKey string `json:"incident_key"`
	PagerDutyBadRequest
}

type PagerDutyOAuthResponseDBWrapper struct {
	Id         int            `json:"id" db:"id"`
	CustomerId string         `json:"customer_id" db:"customer_id" required:"true"`
	Data       types.JSONText `json:"data" db:"data"`
}

func (pd *PagerDutyOAuthResponseDBWrapper) Validate() error {
	validator := &util.Validator{}
	return validator.Validate(pd)
}

// Oath response for slack token
type PagerDutyOAuthResponse struct {
	Account     string `json:"account" db:"account" required:"true"`
	ServiceKey  string `json:"service_key" db:"service_key"`
	ServiceName string `json:"service_name" db:"service_name"`
	Enabled     bool   `json:"enabled" db:"enabled"`
	PagerDutyResponse
}

func (pd *PagerDutyOAuthResponse) Validate() error {
	validator := &util.Validator{}
	return validator.Validate(pd)
}
