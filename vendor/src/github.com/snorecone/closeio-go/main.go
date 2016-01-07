package closeio

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

const baseURL = "https://app.close.io/api"
const version = "v1"

type Closeio struct {
	Token string
}

type Contact struct {
	Name   string   `json:"name"`
	Title  string   `json:"title"`
	Emails []*Email `json:"emails"`
	Phones []*Phone `json:"phones"`
}
type ContactResp struct {
	Name           string      `json:"name"`
	Title          string      `json:"title"`
	Emails         []EmailResp `json:"emails"`
	Phones         []PhoneResp `json:"phones"`
	CreatedBy      string      `json:"created_by"`
	UpdatedBy      string      `json:"updated_by"`
	Id             string      `json:"id"`
	OrganizationId string      `json:"organization_id"`
	DateCreated    string      `json:"date_created"`
	DateUpdated    string      `json:"date_updated"`
}
type Email struct {
	Type  string `json:"type"`
	Email string `json:"email"`
}
type EmailResp struct {
	Type       string `json:"type"`
	Email      string `json:"email"`
	EmailLower string `json:"email_lower"`
}
type Phone struct {
	Type  string `json:"type"`
	Phone string `json:"phone"`
}
type PhoneResp struct {
	Type           string `json:"type"`
	Phone          string `json:"phone"`
	PhoneFormatted string `json:"phone_formatted"`
}
type Address struct {
	Label    string `json:"label"`
	Address1 string `json:"address_1"`
	Address2 string `json:"address_2"`
	City     string `json:"city"`
	State    string `json:"state"`
	Zipcode  string `json:"zipcode"`
	Country  string `json:"country"`
}

func New(token string) *Closeio {
	return &Closeio{token}
}
func marshal(data interface{}) (jsonD []byte, err error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}
func request(urlPart string, reqType string, key string, data []byte) (resp *http.Response, err error) {
	client := &http.Client{}
	url := baseURL + "/" + version + "/" + urlPart
	body := bytes.NewBuffer(data)
	req, err := http.NewRequest(reqType, url, body)
	req.SetBasicAuth(key, "")

	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		bod, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(string(bod))
	}
	return resp, nil
}
