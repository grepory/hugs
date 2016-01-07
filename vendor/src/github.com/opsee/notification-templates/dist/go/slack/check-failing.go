package slack

var CheckFailing = `{
  "text": "Check Failing",
  "username": "OpseeNotifier",
  "icon_url": "https://s3-us-west-1.amazonaws.com/opsee-public-images/slack-avi-48-red.png",
  "attachments": [
    {
      "text": "{{check_name}} failure in {{group_name}}",
      "color": "#f44336"
    },
    {
      "text":"{{fail_count}} of {{instance_count}} Instances Failing",
      "color": "#424242"
    }
  ]
}`
