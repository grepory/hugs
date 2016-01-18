package slack

var CheckPassing = `{
  "text": "Check Passing",
  "username": "OpseeNotifier",
  "icon_url": "https://s3-us-west-1.amazonaws.com/opsee-public-images/slack-avi-48-red.png",
  "attachments": [
    {
      "text": "{{check_name}} passing in {{group_name}}",
      "color": "#f44336"
    },
    {
      "text":"All Instances Passing",
      "color": "#424242"
    }
  ]
}`
