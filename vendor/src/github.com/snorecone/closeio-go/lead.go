package closeio

import (
	"encoding/json"
	"net/url"
)

type Lead struct {
	Name        string            `json:"name,omitempty"`
	Url         string            `json:"url,omitempty"`
	Description string            `json:"description,omitempty"`
	StatusId    string            `json:"status_id,omitempty"`
	Status      string            `json:"status,omitempty"`
	Contacts    []*Contact        `json:"contacts"`
	Custom      map[string]string `json:"custom,omitempty"`
	Addresses   []*Address        `json:"addresses"`
}

type LeadResp struct {
	StatusId       string             `json:"status_id"`
	StatusLabel    string             `json:"status_label"`
	DisplayName    string             `json:"display_name"`
	Description    string             `json:"description"`
	Addresses      []Address          `json:"addresses"`
	Custom         map[string]string  `json:"custom"`
	Name           string             `json:"name"`
	Contacts       []ContactResp      `json:"contacts"`
	Url            string             `json:"url"`
	Id             string             `json:"id"`
	DateUpdated    string             `json:"date_updated"`
	DateCreated    string             `json:"date_created"`
	CreatedBy      string             `json:"created_by"`
	UpdatedBy      string             `json:"updated_by"`
	OrganizationId string             `json:"organization_id"`
	HtmlUrl        string             `json:"html_url"`
	Opportunities  []*OpportunityResp `json:"opportunities"`
	//Tasks          []string          `json:"tasks"` // TODO: change this
}
type Leads struct {
	HasMore      bool       `json:"has_more"`
	TotalResults int        `json:"total_results"`
	Data         []LeadResp `json:"data"`
}
type LeadSearch struct {
	Query string
}

func (c *Closeio) Leads(ls *LeadSearch) (l *Leads, err error) {
	leadType := ""
	if ls == nil {
		leadType = "lead/"
	} else {
		// Set query and encode
		// TODO: Set limit, etc.
		v := url.Values{}
		v.Set("query", ls.Query)
		query := v.Encode()
		leadType = "lead/?" + query
	}
	resp, err := request(leadType, "GET", c.Token, nil)
	if err != nil {
		return nil, err
	}
	leads := Leads{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&leads)
	if err != nil {
		return nil, err
	}
	return &leads, nil
}

func (c *Closeio) CreateLead(lead *Lead) (l *LeadResp, err error) {
	data, err := marshal(lead)
	if err != nil {
		return nil, err
	}
	resp, err := request("lead/", "POST", c.Token, data)
	if err != nil {
		return nil, err
	}
	leadresp := LeadResp{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&leadresp)
	if err != nil {
		return nil, err
	}
	return &leadresp, nil
}

func (c *Closeio) GetLead(id string) (l *LeadResp, err error) {
	resp, err := request("lead/"+id+"/", "GET", c.Token, nil)
	if err != nil {
		return nil, err
	}
	lead := LeadResp{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&lead)
	if err != nil {
		return nil, err
	}
	return &lead, nil
}
func (c *Closeio) DeleteLead(id string) error {
	_, err := request("lead/"+id+"/", "DELETE", c.Token, nil)
	if err != nil {
		return err
	}
	return nil
}
