package sqsconsumer

import (
	"encoding/json"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/opsee/hugs/config"
	"github.com/opsee/hugs/notifier"
	"github.com/opsee/hugs/store"
	"github.com/sirupsen/logrus"
)

type Worker struct {
	Site              *Site
	SQS               *sqs.SQS
	Store             *store.Postgres
	Notifier          *notifier.Notifier
	CommandChan       chan ForemanCommand
	errCount          int
	errCountThreshold int
}

func NewWorker(site *Site) *Worker {
	s, err := store.NewPostgres(site.DBUrl)
	if err != nil {
		logrus.Fatal("Unable to connect to postgres! ", err)
	}
	return &Worker{
		Site:              site,
		SQS:               sqs.New(config.GetConfig().AWSSession),
		Store:             s,
		Notifier:          notifier.NewNotifier(),
		CommandChan:       make(chan ForemanCommand),
		errCount:          0,
		errCountThreshold: 13,
	}
}

func (w *Worker) Start() {
	go func() {
		w.Site.WorkerPool <- w.CommandChan
		atomic.AddInt64(w.Site.CurrentWorkerCount, 1)
		for {
			select {
			case command := <-w.CommandChan:
				if command == Quit {
					logrus.Info("Stopping worker")
					atomic.AddInt64(w.Site.CurrentWorkerCount, -1)
					return
				}
			default:
				w.Work()
			}
		}
	}()
}

func (w *Worker) Work() {
	input := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(w.Site.QueueUrl),
		MaxNumberOfMessages: aws.Int64(10),
	}
	message, err := w.SQS.ReceiveMessage(input)

	if err != nil {
		w.errCount += 1
		logrus.Error(err)
		if w.errCount >= w.errCountThreshold {
			w.Stop()
		}
		time.Sleep((1 << uint(w.errCount+1)) * time.Millisecond * 10)
		return
	}
	w.errCount = 0

	// unmarshal sqs message json into events
	for _, message := range message.Messages {
		if message.Body == nil {
			continue
		}

		bodyBytes := []byte(*message.Body)
		event := notifier.Event{}
		json.Unmarshal(bodyBytes, &event)

		if ok := event.Validate(); ok {
			notifications, err := w.Store.UnsafeGetNotificationsByCheckID(event.CheckID)
			if err != nil {
				//TODO(dan) send message back to sqs if you can't get notifications
				// OR send notification to seperate SQS queue for redelivery
				logrus.Warn("Worker: Couldn't get notifications for event.")
			} else {
				for _, notification := range notifications {
					// Send notification with Notifier
					sendErr := w.Notifier.Send(notification, event)
					if sendErr != nil {
						logrus.WithFields(logrus.Fields{"Error": sendErr}).Error("Error emitting notification")
					}
				}
			}
		} else {
			logrus.Warn("Worker: event failed validation.")
		}
	}
}

func (w *Worker) Stop() {
	go func() {
		w.CommandChan <- Quit
	}()
}
