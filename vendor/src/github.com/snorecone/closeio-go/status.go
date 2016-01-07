package closeio

import (
	"encoding/json"
)

type Statuses struct {
	HasMore bool     `json:"has_more"`
	Data    []Status `json:"data"`
}
type Status struct {
	Id    string `json:"id"`
	Label string `json:"label"`
}

func (c *Closeio) Statuses() (*Statuses, error) {
	resp, err := request("status/lead/", "GET", c.Token, nil)
	if err != nil {
		return nil, err
	}
	statuses := Statuses{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&statuses)
	if err != nil {
		return nil, err
	}
	return &statuses, nil
}
