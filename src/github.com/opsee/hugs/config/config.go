package config

import (
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/sirupsen/logrus"
)

var HugsConfig = NewConfig()

type Config struct {
	PublicHost     string
	PostgresConn   string
	SqsUrl         string
	OpseeHost      string
	MandrillApiKey string
	MaxWorkers     int64
	MinWorkers     int64
	AWSSession     *session.Session
}

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

func NewConfig() *Config {
	// try to safely get max workers from env
	defaultMaxWorkers := int64(10)
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
			logrus.Warn("Errors getting HUGS_MAX_WORKERS and HUGS_MIN_WORKERS: ", err1, " ", err2)
		}
	} else {
		logrus.Warn("Config: using default value for MaxWorkers and MinWorkers")
	}

	c := &Config{
		PublicHost:     os.Getenv("HUGS_HOST"),
		PostgresConn:   os.Getenv("HUGS_POSTGRES_CONN"),
		SqsUrl:         os.Getenv("HUGS_SQS_URL"),
		OpseeHost:      os.Getenv("HUGS_OPSEE_HOST"),
		MandrillApiKey: os.Getenv("HUGS_MANDRILL_API_KEY"),
		MaxWorkers:     maxWorkers,
		MinWorkers:     minWorkers,
	}
	c.getAWSSession()

	return c
}

func GetConfig() *Config {
	return HugsConfig
}
