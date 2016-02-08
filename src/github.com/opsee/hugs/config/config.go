package config

import (
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/opsee/hugs/util"
	log "github.com/sirupsen/logrus"
)

// TODO(dan) consider splitting this into configs and testconfigs for each module
type Config struct {
	PublicHost            string `required:"true"`
	PostgresConn          string `required:"true"`
	SqsUrl                string `required:"true"`
	AWSRegion             string `required:"true"`
	OpseeHost             string `required:"true"`
	MandrillApiKey        string `required:"true"`
	VapeEndpoint          string `required:"true"`
	VapeKey               string `required:"true"`
	LogLevel              string
	SlackClientSecret     string `required:"true"`
	SlackClientID         string `required:"true"`
	SlackTestToken        string
	SlackTestClientSecret string
	SlackTestClientID     string
	AWSSession            *session.Session
	NotificaptionEndpoint string
	BartnetEndpoint       string
}

func (this *Config) Validate() error {
	validator := &util.Validator{}
	if err := validator.Validate(this); err != nil {
		return err
	}
	return nil
}

var hugsConfig *Config
var once sync.Once

func (this *Config) getAWSSession() {
	creds := credentials.NewChainCredentials(
		[]credentials.Provider{
			&ec2rolecreds.EC2RoleProvider{
				Client: ec2metadata.New(session.New()),
			},
			&credentials.EnvProvider{},
		})

	this.AWSSession = session.New(&aws.Config{
		Credentials: creds,
		MaxRetries:  aws.Int(11),
		Region:      aws.String(this.AWSRegion),
	})
}

func (this *Config) setLogLevel() {
	if len(this.LogLevel) > 0 {
		level, err := log.ParseLevel(this.LogLevel)
		if err == nil {
			log.SetLevel(level)
			return
		}
	}
	log.WithFields(log.Fields{"config": "setLogLevel"}).Warn("Could not set log level!")
}

func GetConfig() *Config {
	once.Do(func() {
		c := &Config{
			PublicHost:            os.Getenv("HUGS_HOST"),
			PostgresConn:          os.Getenv("HUGS_POSTGRES_CONN"),
			SqsUrl:                os.Getenv("HUGS_SQS_URL"),
			AWSRegion:             os.Getenv("HUGS_AWS_REGION"),
			OpseeHost:             os.Getenv("HUGS_OPSEE_HOST"),
			MandrillApiKey:        os.Getenv("HUGS_MANDRILL_API_KEY"),
			VapeEndpoint:          os.Getenv("HUGS_VAPE_ENDPOINT"),
			VapeKey:               os.Getenv("HUGS_VAPE_KEYFILE"),
			LogLevel:              os.Getenv("HUGS_LOG_LEVEL"),
			SlackClientID:         os.Getenv("HUGS_SLACK_CLIENT_ID"),
			SlackClientSecret:     os.Getenv("HUGS_SLACK_CLIENT_SECRET"),
			SlackTestToken:        os.Getenv("HUGS_TEST_SLACK_TOKEN"),
			SlackTestClientID:     os.Getenv("HUGS_TEST_SLACK_CLIENT_ID"),
			SlackTestClientSecret: os.Getenv("HUGS_TEST_SLACK_CLIENT_SECRET"),
			NotificaptionEndpoint: os.Getenv("HUGS_NOTIFICAPTION_ENDPOINT"),
			BartnetEndpoint:       os.Getenv("HUGS_BARTNET_ENDPOINT"),
		}
		if err := c.Validate(); err == nil {
			c.setLogLevel()
			c.getAWSSession()
			hugsConfig = c
		} else {
			log.WithFields(log.Fields{"config": "Validate", "error": err}).Fatal("Error generating config.")
		}
	})

	return hugsConfig
}
