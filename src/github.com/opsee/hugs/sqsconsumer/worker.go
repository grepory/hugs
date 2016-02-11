package sqsconsumer

import (
	//"encoding/json"
	"encoding/base64"
	"net/http"
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
	ID                string
	SQS               *sqs.SQS
	SQSUrl            string
	Store             *store.Postgres
	Notifier          *notifier.Notifier
	errCount          int
	errCountThreshold int
}

func NewWorker(ID string, maxErr int, sqsUrl string) (*Worker, error) {
	s, err := store.NewPostgres(config.GetConfig().PostgresConn)
	if err != nil {
		return nil, err
	}

	// create new notifier and warn on errors
	notifier, errMap := notifier.NewNotifier()
	for k, v := range errMap {
		if v != nil {
			log.WithFields(log.Fields{"worker": ID, "error": v}).Info("Couldn't initialize notifier: ", k)
		}
	}

	return &Worker{
		ID:                ID,
		SQS:               sqs.New(config.GetConfig().AWSSession),
		SQSUrl:            sqsUrl,
		Store:             s,
		Notifier:          notifier,
		errCount:          0,
		errCountThreshold: maxErr,
	}, nil
}

func (w *Worker) Start() {
	log.WithFields(log.Fields{"worker": w.ID}).Info("Starting up.")
	for {
		w.Work()
	}
}

// TODO(greg): We need to be deleting messages from the queue. As it stands, we're just requeueing them over and over again.
func (w *Worker) Work() {
	log.WithFields(log.Fields{"worker": w.ID}).Info("Doing work...")

	input := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(w.SQSUrl),
		MaxNumberOfMessages: aws.Int64(10),
		WaitTimeSeconds:     aws.Int64(20),
	}
	message, err := w.SQS.ReceiveMessage(input)

	if err != nil {
		log.WithFields(log.Fields{"worker": w.ID, "err": err}).Error("Encountered error.  Sleeping...")

		w.errCount += 1
		if w.errCount >= w.errCountThreshold {
			w.errCount = w.errCountThreshold / 2
			return
		}
		time.Sleep((1 << uint(w.errCount+1)) * time.Millisecond * 10)
		return
	}

	if len(message.Messages) > 0 {
		log.WithFields(log.Fields{"worker": w.ID, "message_count": len(message.Messages)}).Info("Got messages...")
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

		doDelete := false
		if len(notifications) == 0 {
			log.WithFields(log.Fields{"worker": w.ID, "check": result.CheckId}).Info("Deleting check with no notifications.")
			doDelete = true
		}

		for _, notification := range notifications {
			// Send notification with Notifier
			sendErr := w.Notifier.Send(notification, event)
			if sendErr != nil {
				log.WithFields(log.Fields{"worker": w.ID, "err": sendErr}).Error("Error emitting notification")
			} else {
				// If we successfully send one notification, then we're going to delete the SQS Message.
				// TODO(greg): Separate queues per notification type.
				doDelete = true
			}
		}

		if doDelete {
			// TODO(dan) we can't wait too long here or the message will become visible again.
			deleteMessageInput := &sqs.DeleteMessageInput{
				QueueUrl:      aws.String(w.SQSUrl),
				ReceiptHandle: message.ReceiptHandle,
			}
			deletedMessage := false
			for deleteTry := 1; deleteTry < 5; deleteTry++ {
				_, err := w.SQS.DeleteMessage(deleteMessageInput)
				if err == nil {
					deletedMessage = true
					break
				}
				time.Sleep((1 << uint(deleteTry+1)) * time.Millisecond * 10)
			}
			if deletedMessage == false {
				log.WithFields(log.Fields{"worker": w.ID, "message": message}).Error("Couldn't delete message from queue.")
			}
		}
	}
}
