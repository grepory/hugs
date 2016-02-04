package sqsconsumer

import (
	//"encoding/json"
	"encoding/base64"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/golang/protobuf/proto"
	"github.com/opsee/hugs/checker"
	"github.com/opsee/hugs/config"
	"github.com/opsee/hugs/notifier"
	"github.com/opsee/hugs/store"
	log "github.com/sirupsen/logrus"
)

var (
	httpClient = &http.Client{
		Timeout: 15 * time.Second,
	}
)

type Worker struct {
	ID                int64
	Site              *Site
	SQS               *sqs.SQS
	Store             *store.Postgres
	Notifier          *notifier.Notifier
	CommandChan       chan ForemanCommand
	errCount          int
	errCountThreshold int
}

func NewWorker(site *Site) (*Worker, error) {
	s, err := store.NewPostgres(site.DBUrl)
	if err != nil {
		return nil, err
	}

	// create new notifier and warn on errors
	notifier, errMap := notifier.NewNotifier()
	for k, v := range errMap {
		if v != nil {
			log.WithFields(log.Fields{"worker": "initializing", "error": v}).Info("Couldn't initialize notifier: ", k)
		}
	}

	return &Worker{
		ID:                -1,
		Site:              site,
		SQS:               sqs.New(config.GetConfig().AWSSession),
		Store:             s,
		Notifier:          notifier,
		CommandChan:       make(chan ForemanCommand),
		errCount:          0,
		errCountThreshold: 12,
	}, nil
}

func (w *Worker) Start() {
	go func() {
		w.Site.WorkerPool <- w.CommandChan
		w.ID = atomic.AddInt64(w.Site.CurrentWorkerCount, 1)
		for {
			select {
			case command := <-w.CommandChan:
				if command == Quit {
					log.WithFields(log.Fields{"worker": w.ID}).Info("Quitting.")
					atomic.AddInt64(w.Site.CurrentWorkerCount, -1)
					return
				}
			default:
				w.Work()
			}
		}
	}()
}

// TODO(greg): We need to be deleting messages from the queue. As it stands, we're just requeueing them over and over again.
func (w *Worker) Work() {
	input := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(w.Site.QueueUrl),
		MaxNumberOfMessages: aws.Int64(10),
	}
	message, err := w.SQS.ReceiveMessage(input)

	if err != nil || len(message.Messages) == 0 {
		w.errCount += 1
		if w.errCount >= w.errCountThreshold {
			w.errCount = w.errCountThreshold / 2
			return
		}
		if err != nil {
			log.WithFields(log.Fields{"worker": w.ID, "err": err}).Error("Encountered error.  Sleeping...")
		}
		time.Sleep((1 << uint(w.errCount+1)) * time.Millisecond * 10)
		return
	}
	if len(message.Messages) > 0 {
		w.errCount = 0
	}

	// unmarshal sqs message json into events
	for _, message := range message.Messages {
		if message.Body == nil {
			continue
		}

		bodyBytes, err := base64.StdEncoding.DecodeString(*message.Body)
		if err != nil {
			log.WithFields(log.Fields{"worker": w.ID, "err": err, "message": *message.Body}).Error("Cannot decode message body")
			continue
		}
		result := &checker.CheckResult{}
		err = proto.Unmarshal(bodyBytes, result)
		if err != nil {
			log.WithFields(log.Fields{"worker": w.ID, "err": err, "message": *message.Body}).Error("Cannot unmarshal message body")
			continue
		}
		log.WithFields(log.Fields{"worker": w.ID, "CheckResult": result.String()}).Info("Unmarshalled CheckResult.")

		notifications, err := w.Store.UnsafeGetNotificationsByCheckID(result.CheckId)
		if err != nil {
			//TODO(dan) send message back to sqs if you can't get notifications
			// OR send notification to seperate SQS queue for redelivery
			log.WithFields(log.Fields{"worker": w.ID}).Warn("Worker: Couldn't get notifications for event.")
			continue
		}

		if len(notifications) < 1 {
			log.WithFields(log.Fields{"worker": w.ID}).Warn("Worker: No notifications for event.")
			continue
		}

		event := buildEvent(notifications[0], result)

		for _, notification := range notifications {
			// Send notification with Notifier
			sendErr := w.Notifier.Send(notification, event)
			if sendErr != nil {
				log.WithFields(log.Fields{"worker": w.ID, "err": sendErr}).Error("Error emitting notification")
			}
		}
	}
}

func (w *Worker) Stop() {
	go func() {
		w.CommandChan <- Quit
	}()
}
