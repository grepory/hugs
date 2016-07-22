package pagerduty
var CheckFailing = `{
  "service_key": "{{service_key}}",
  "incident_key":"{{check_id}}",
  "description":"{{check_name}} failure in {{group_name}}",
  "client_url":"https://{{opsee_host}}/check/{{check_id}}",
  "event_type":"trigger",
  "details":"{{{event_json}}}"
}
`
