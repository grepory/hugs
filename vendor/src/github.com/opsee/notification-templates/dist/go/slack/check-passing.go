package slack
var CheckPassing = `{
 "token": "{{token}}",
  "channel":"{{channel}}",
  "text": "*{{check_name}}* passing in *{{group_name}}*",
  "username": "OpseeBot",
  "icon_url": "https://s3-us-west-1.amazonaws.com/opsee-public-images/slack-avi-48-green.png",
  "attachments": [
    {
      "text": "Everything is fine",
      "color": "#69a92c"
    }
  ]
}
`
