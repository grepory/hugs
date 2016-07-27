package pagerduty

import "github.com/hoisie/mustache"

var CheckPassing = `{
  "service_key": "{{service_key}}",
  "incident_key":"{{check_id}}",
  "event_type":"resolve"
}
`

func init() {
	tmpl, err := mustache.ParseString(CheckPassing)
	if err != nil {
		panic(err)
	}
	Templates["check-passing"] = tmpl
}
