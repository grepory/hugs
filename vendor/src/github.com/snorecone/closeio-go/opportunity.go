package closeio

import "encoding/json"

type Opportunity struct {
	Confidence  int    `json:"confidence,omitempty"`
	Status      string `json:"status,omitempty"`
	StatusId    string `json:"status,omitempty"`
	Value       int    `json:"value,omitempty"`
	ValuePeriod string `json:"value_period,omitempty"` //Monthly, annually, one-time
	Note        string `json:"note,omitempty"`
	LeadId      string `json:"lead_id"`
}

type OpportunityResp struct {
	StatusId       string `json:"status_id"`
	StatusLabel    string `json:"status_label"`
	StatusType     string `json:"status_type"`
	DateWon        string `json:"date_won"`
	Confidence     int    `json:"confidence"`
	Userid         string `json:"user_id"`
	ContactId      string `json:"contact_id"`
	UpdatedBy      string `json:"updated_by"`
	DateUpdated    string `json:"date_updated"`
	CreatedBy      string `json:"created_by"`
	LeadId         string `json:"lead_id"`
	Note           string `json:"note"`
	Value          int    `json:"value"`
	DateCreated    string `json:"date_created"`
	OrganizationId string `json:"organization_id"`
	LeadName       string `json:"lead_name"`
	UserName       string `json:"user_name"`
	Id             string `json:"id"`
	ValuePeriod    string `json:"value_period"`
}

func (c *Closeio) CreateOpportunity(opportunity *Opportunity) (l *OpportunityResp, err error) {
	data, err := marshal(opportunity)
	if err != nil {
		return nil, err
	}
	resp, err := request("opportunity/", "POST", c.Token, data)
	if err != nil {
		return nil, err
	}
	opportunityresp := OpportunityResp{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&opportunityresp)
	if err != nil {
		return nil, err
	}
	return &opportunityresp, nil
}
