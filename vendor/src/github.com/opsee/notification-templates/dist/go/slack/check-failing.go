package slack
var CheckFailing = `{
  "token": "{{token}}",
  "channel":"{{channel}}",
  "text": "*{{check_name}}* failing in *{{group_name}}*",
  "username": "OpseeBot",
  "icon_url": "https://s3-us-west-1.amazonaws.com/opsee-public-images/slack-avi-48-red.png",
  "attachments": [
    {
      "text":"{{fail_count}} of {{instance_count}} Failing",
      "color": "#f44336"
    }
  ]
}
`
