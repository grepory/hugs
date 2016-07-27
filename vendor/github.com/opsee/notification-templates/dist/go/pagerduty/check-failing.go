package pagerduty

import "github.com/hoisie/mustache"

var CheckFailing = `{
  "service_key": "{{service_key}}",
  "incident_key":"{{check_id}}",
  "description":"{{check_name}} failure in {{group_name}}",
  "client_url":"https://{{opsee_host}}/check/{{check_id}}",
  "event_type":"trigger",
  "details":"{{{event_json}}}"
}
`

func init() {
	tmpl, err := mustache.ParseString(CheckFailing)
	if err != nil {
		panic(err)
	}
	Templates["check-failing"] = tmpl
}
