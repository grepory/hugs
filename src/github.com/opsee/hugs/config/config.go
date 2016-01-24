package config

import (
	"os"
	"strconv"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	log "github.com/sirupsen/logrus"
)

// TODO(dan) consider splitting this into configs and testconfigs for each module
type Config struct {
	PublicHost        string
	PostgresConn      string
	SqsUrl            string
	OpseeHost         string
	MandrillApiKey    string
	VapeEndpoint      string
	VapeKey           string
	MaxWorkers        int64
	MinWorkers        int64
	LogLevel          string
	SlackClientSecret string
	SlackClientID     string
	SlackTestToken    string
	AWSSession        *session.Session
}

var hugsConfig *Config
var once sync.Once

func (config *Config) getAWSSession() {
	creds := credentials.NewChainCredentials(
		[]credentials.Provider{
			&ec2rolecreds.EC2RoleProvider{
				Client: ec2metadata.New(session.New()),
			},
			&credentials.EnvProvider{},
		})

	config.AWSSession = session.New(&aws.Config{
		Credentials: creds,
		MaxRetries:  aws.Int(11),
		Region:      aws.String(os.Getenv("HUGS_AWS_REGION")),
	})
}

func (config *Config) setLogLevel() {
	if len(config.LogLevel) > 0 {
		level, err := log.ParseLevel(config.LogLevel)
		if err == nil {
			log.SetLevel(level)
			return
		}
	}
	log.Warn("Config: couldn't parse loglevel.")
}

func GetConfig() *Config {
	once.Do(func() {
		// try to safely get max workers from env
		defaultMaxWorkers := int64(25)
		defaultMinWorkers := int64(1)
		maxWorkers := defaultMaxWorkers
		minWorkers := defaultMinWorkers

		maxWorkersString := os.Getenv("HUGS_MAX_WORKERS")
		minWorkersString := os.Getenv("HUGS_MIN_WORKERS")

		if len(maxWorkersString) > 0 && len(minWorkersString) > 0 {
			maxWorkersEnv, err1 := strconv.ParseInt(maxWorkersString, 10, 32)
			minWorkersEnv, err2 := strconv.ParseInt(minWorkersString, 10, 32)
			if err1 == nil && err2 == nil {
				if minWorkersEnv > 0 && maxWorkersEnv > minWorkersEnv {
					maxWorkers = maxWorkersEnv
					minWorkers = minWorkersEnv
				}
			} else {
				log.Warn("Errors getting HUGS_MAX_WORKERS and HUGS_MIN_WORKERS: ", err1, " ", err2)
			}
		} else {
			log.Warn("Config: using default value for MaxWorkers and MinWorkers")
		}

		c := &Config{
			PublicHost:        os.Getenv("HUGS_HOST"),
			PostgresConn:      os.Getenv("HUGS_POSTGRES_CONN"),
			SqsUrl:            os.Getenv("HUGS_SQS_URL"),
			OpseeHost:         os.Getenv("HUGS_OPSEE_HOST"),
			MandrillApiKey:    os.Getenv("HUGS_MANDRILL_API_KEY"),
			VapeEndpoint:      os.Getenv("HUGS_VAPE_ENDPOINT"),
			VapeKey:           os.Getenv("HUGS_VAPE_KEYFILE"),
			LogLevel:          os.Getenv("HUGS_LOG_LEVEL"),
			SlackClientID:     os.Getenv("HUGS_SLACK_CLIENT_ID"),
			SlackClientSecret: os.Getenv("HUGS_SLACK_CLIENT_SECRET"),
			SlackTestToken:    os.Getenv("HUGS_SLACK_TEST_TOKEN"),
			MaxWorkers:        maxWorkers,
			MinWorkers:        minWorkers,
		}
		c.setLogLevel()
		c.getAWSSession()
		hugsConfig = c

		log.WithFields(log.Fields{"module": "config", "PublicHost": c.PublicHost, "SQSUrl": c.SqsUrl, "OpseeHost": c.OpseeHost, "MaxWorkers": c.MaxWorkers, "MinWorkers": c.MinWorkers}).Info("Created new config.")
	})

	return hugsConfig
}
