package sqsconsumer

import (
	//"encoding/json"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/golang/protobuf/proto"
	"github.com/opsee/basic/schema"
	"github.com/opsee/hugs/config"
	"github.com/opsee/hugs/notifier"
	"github.com/opsee/hugs/store"
	log "github.com/sirupsen/logrus"
	"github.com/yeller/yeller-golang"
)

var (
	httpClient = &http.Client{
		Timeout: 15 * time.Second,
	}
)

type Worker struct {
	Id                string
	SQS               *sqs.SQS
	SQSUrl            string
	Store             *store.Postgres
	Notifier          *notifier.Notifier
	errCount          int
	errCountThreshold int
}

func NewWorker(Id string, maxErr int, sqsUrl string) (*Worker, error) {
	s, err := store.NewPostgres()
	if err != nil {
		return nil, err
	}

	// create new notifier and warn on errors
	notifier, errMap := notifier.NewNotifier()
	for k, v := range errMap {
		if v != nil {
			log.WithFields(log.Fields{"worker": Id, "error": v}).Info("Couldn't initialize notifier: ", k)
		}
	}

	return &Worker{
		Id:                Id,
		SQS:               sqs.New(config.GetConfig().AWSSession),
		SQSUrl:            sqsUrl,
		Store:             s,
		Notifier:          notifier,
		errCount:          0,
		errCountThreshold: maxErr,
	}, nil
}

func (w *Worker) Start() {
	log.WithFields(log.Fields{"worker": w.Id}).Info("Starting up.")

	for {
		w.Work()
	}
}

func (w *Worker) deleteMessage(handle *string) error {
	var err error

	if handle == nil {
		return errors.New("No message handle for SQS message")
	}

	// TODO(dan) we can't wait too long here or the message will become visible again.
	deleteMessageInput := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(w.SQSUrl),
		ReceiptHandle: handle,
	}

	for deleteTry := 1; deleteTry < 5; deleteTry++ {
		_, err = w.SQS.DeleteMessage(deleteMessageInput)
		if err != nil {
			time.Sleep((1 << uint(deleteTry+1)) * time.Millisecond * 10)
		}
	}

	return err
}

func (w *Worker) Work() {
	log.WithFields(log.Fields{"worker": w.Id}).Info("Doing work...")

	input := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(w.SQSUrl),
		MaxNumberOfMessages: aws.Int64(10),
		WaitTimeSeconds:     aws.Int64(20),
	}
	message, err := w.SQS.ReceiveMessage(input)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{"worker": w.Id}).Error("Encountered error polling SQS.  Sleeping...")

		w.errCount += 1
		if w.errCount >= w.errCountThreshold {
			w.errCount = w.errCountThreshold / 2
			return
		}
		time.Sleep((1 << uint(w.errCount+1)) * time.Millisecond * 10)
		return
	}

	if len(message.Messages) > 0 {
		log.WithFields(log.Fields{"worker": w.Id, "message_count": len(message.Messages)}).Info("Got messages...")
		w.errCount = 0
	}

	// unmarshal sqs message json into events
	for _, message := range message.Messages {
		if message.Body == nil {
			continue
		}

		bodyBytes, err := base64.StdEncoding.DecodeString(*message.Body)
		if err != nil {
			log.WithFields(log.Fields{"worker": w.Id, "err": err, "message": *message.Body}).Error("Cannot decode message body")
			info := make(map[string]interface{})
			info["message"] = string(bodyBytes)
			yeller.NotifyInfo(err, info)
			if err := w.deleteMessage(message.ReceiptHandle); err != nil {
				log.WithError(err).WithFields(log.Fields{"worker": w.Id, "message": *message.Body}).Error("Cannot delete message from SQS.")
			}
		}

		result := &schema.CheckResult{}
		err = proto.Unmarshal(bodyBytes, result)
		if err != nil {
			log.WithFields(log.Fields{"worker": w.Id, "err": err, "message": *message.Body}).Error("Cannot unmarshal message body")
			info := make(map[string]interface{})
			info["message"] = string(bodyBytes)
			yeller.NotifyInfo(err, info)
			if err := w.deleteMessage(message.ReceiptHandle); err != nil {
				log.WithError(err).WithFields(log.Fields{"worker": w.Id, "message": *message.Body}).Error("Cannot delete message from SQS.")
			}
		}
		log.WithFields(log.Fields{"worker": w.Id, "CheckResult": result.String()}).Info("Unmarshalled CheckResult.")

		notifications, err := w.Store.UnsafeGetNotificationsByCheckId(result.CheckId)
		if err != nil {
			//TODO(dan) send message back to sqs if you can't get notifications
			// OR send notification to seperate SQS queue for redelivery
			log.WithFields(log.Fields{"worker": w.Id}).Warn("Worker: Couldn't get notifications for event.")
			info := make(map[string]interface{})
			info["message"] = string(bodyBytes)
			yeller.NotifyInfo(err, info)
			continue
		}

		if len(notifications) < 1 {
			log.WithFields(log.Fields{"worker": w.Id, "check": result.CheckId}).Info("Deleting check with no notifications.")
			if err := w.deleteMessage(message.ReceiptHandle); err != nil {
				log.WithError(err).WithFields(log.Fields{"worker": w.Id, "message": *message.Body}).Error("Cannot delete message from SQS.")
			}
		} else {
			event := buildEvent(notifications[0], result)

			var msg string
			if event.Result.Passing {
				msg = "Sending passing notifications to customer."
			} else {
				msg = "Sending failing notifications to customer."
			}

			log.WithFields(log.Fields{
				"customer_id": event.Result.CustomerId,
				"check_id":    event.Result.CheckId,
			}).Info(msg)

			for _, notification := range notifications {
				// Send notification with Notifier
				sendErr := w.Notifier.Send(notification, event)
				if sendErr != nil {
					log.WithFields(log.Fields{"worker": w.Id, "err": sendErr}).Error("Error emitting notification")
				} else {
					log.WithFields(log.Fields{
						"customer_id": event.Result.CustomerId,
						"check_id":    event.Result.CheckId,
					}).Info(fmt.Sprintf("Sent %s notification to customer.", notification.Type))
					// If we successfully send one notification, then we're going to delete the SQS Message.
					// TODO(greg): Separate queues per notification type.
					if err := w.deleteMessage(message.ReceiptHandle); err != nil {
						log.WithError(err).WithFields(log.Fields{"worker": w.Id, "message": *message.Body}).Error("Cannot delete message from SQS.")
					}
				}
			}
		}
	}
}
