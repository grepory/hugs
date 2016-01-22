package service

type j map[string]interface{}
type k []string

var swaggerMap = j{
	"basePath": "/",
	"swagger":  "2.0",
	"info": j{
		"title":       "Notifications API",
		"version":     "0.0.1",
		"description": "API for bastion management",
	},
	"paths": j{
		"/services/slack": j{
			"post": j{
				"parameters": []j{
					j{
						"description": "",
						"in":          "body",
						"name":        "SlackOAuthRequest",
						"required":    true,
						"schema": j{
							"$ref": "#/definitions/SlackOAuthRequest",
						},
					},
				},
				"responses": j{
					"200": j{
						"description": "",
						"schema": j{
							"$ref": "#/definitions/SlackOAuthResponse",
						},
					},
				},
				"summary": "Create a new notification.",
				"tags":    k{"slackauthcode"},
			},
		},
		"/notifications": j{
			"get": j{
				"responses": j{
					"200": j{
						"description": "",
						"schema": j{
							"items": j{
								"$ref": "#/definitions/CheckNotifications",
							},
							"type": "array",
						},
					},
				},
				"summary": "Retrieve all of a customer's notifications",
				"tags":    k{"notifications"},
			},
			"post": j{
				"parameters": []j{
					j{
						"description": "",
						"in":          "body",
						"name":        "CheckNotifications",
						"required":    true,
						"schema": j{
							"$ref": "#/definitions/CheckNotifications",
						},
					},
				},
				"responses": j{
					"200": j{
						"description": "",
						"schema": j{
							"$ref": "#/definitions/CheckNotifications",
						},
					},
				},
				"summary": "Create a new notification.",
				"tags":    k{"notifications"},
			},
		},
		"/notifications/j {check_id}": j{
			"delete": j{
				"parameters": []j{
					j{
						"description": "",
						"in":          "path",
						"name":        "check_id",
						"required":    true,
						"type":        "string",
					},
				},
				"responses": j{
					"default": j{
						"description": "",
					},
				},
				"summary": "Deletes a notification.",
				"tags":    k{"notifications"},
			},
			"get": j{
				"parameters": []j{
					j{
						"description": "",
						"in":          "path",
						"name":        "check_id",
						"required":    true,
						"type":        "string",
					},
				},
				"responses": j{
					"200": j{
						"description": "",
						"schema": j{
							"$ref": "#/definitions/CheckNotifications",
						},
					},
				},
				"summary": "Retrieves a notification.",
				"tags":    k{"notifications"},
			},
			"put": j{
				"parameters": []j{
					j{
						"description": "",
						"in":          "body",
						"name":        "CheckNotifications",
						"required":    true,
						"schema": j{
							"$ref": "#/definitions/CheckNotifications",
						},
					},
					j{
						"description": "",
						"in":          "path",
						"name":        "check_id",
						"required":    true,
						"type":        "string",
					},
				},
				"responses": j{
					"200": j{
						"description": "",
						"schema": j{
							"$ref": "#/definitions/CheckNotifications",
						},
					},
				},
				"summary": "Replaces a notification.",
				"tags":    k{"notifications"},
			},
		},
	},
	"definitions": j{
		"CheckNotifications": j{
			"properties": j{
				"check-id": j{
					"type": "string",
				},
				"notifications": j{
					"items": j{
						"$ref": "#/definitions/Notification",
					},
					"type": "array",
				},
			},
			"required": k{
				"check-id",
				"notifications",
			},
			"type": "object",
		},
		"Notification": j{
			"properties": j{
				"type": j{
					"type": "string",
				},
				"value": j{
					"type": "string",
				},
			},
			"required": k{
				"type",
				"value",
			},
			"type": "object",
		},
		"SlackOAuthRequest": j{
			"properties": j{
				"client_id": j{
					"type": "string",
				},
				"client_secret": j{
					"type": "string",
				},
				"code": j{
					"type": "string",
				},
				"redirect_uri": j{
					"type": "string",
				},
			},
			"required": k{
				"code",
			},
			"type": "object",
		},
		"SlackOAuthResponse": j{
			"properties": j{
				"access_token": j{
					"type": "string",
				},
				"scope": j{
					"type": "string",
				},
				"team_name": j{
					"type": "string",
				},
				"team_id": j{
					"type": "string",
				},
			},
			"required": k{
				"access_token", "scope", "team_name", "team_id",
			},
			"type": "object",
		},
	},

	"consumes": j{},
	"produces": j{},
}
