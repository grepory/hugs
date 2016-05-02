package pagerduty
var CheckPassing = `{
  "service_key": "{{service_key}}",
  "incident_key":"{{check_id}}",
  "event_type":"resolve"
}
`
