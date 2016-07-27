package slack

import "github.com/hoisie/mustache"

var NewCc = `{
  "text": ":pepe-trump: :bread: :garf: NEW CREDIT CARD ENTERED :garf: :bread: :pepe-trump:",
  "username": "CustomerBot",
  "icon_url": "https://s3-us-west-1.amazonaws.com/opsee-public-images/slack-avi-48-green.png",
  "attachments": [
    {
      "text": "{{team_name}} - {{user_email}}",
      "color": "#81C784",
      "fields": [
          {
              "title": "Plan",
              "value": "{{subscription_plan}}",
              "short": true
          },
          {
              "title": "# Checks",
              "value": "{{subscription_quantity}}",
              "short": true
          },
          {
              "title": "Amount",
              "value": "{{subscription_amount}}",
              "short": true
          },
        ]
    }
  ]
}
`

func init() {
	tmpl, err := mustache.ParseString(NewCc)
	if err != nil {
		panic(err)
	}
	Templates["new-cc"] = tmpl
}
