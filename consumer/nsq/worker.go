package nsq

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/nsqio/go-nsq"
	"github.com/opsee/basic/schema"
	hugsconsumer "github.com/opsee/hugs/consumer"
	"github.com/opsee/hugs/notifier"
	"github.com/opsee/hugs/store"
	log "github.com/opsee/logrus"
)

var nsqTopic = "alerts"

type Worker struct {
	Id       string
	Store    *store.Postgres
	Notifier *notifier.Notifier
}

func NewWorker(Id string) (*Worker, error) {
	s, err := store.NewPostgres()
	if err != nil {
		return nil, err
	}

	// create new notifier and warn on errors
	notifier, errMap := notifier.NewNotifier()
	for k, v := range errMap {
		if v != nil {
			log.WithFields(log.Fields{"worker": Id, "error": v}).Info("Couldn't initialize notifier: ", k)
			return nil, v
		}
	}

	return &Worker{
		Id:       Id,
		Store:    s,
		Notifier: notifier,
	}, nil
}

func (w *Worker) Start() error {
	logger := log.WithFields(log.Fields{"worker": w.Id})
	logger.Info("Starting up.")

	config := nsq.NewConfig()
	config.MaxInFlight = 4
	consumer, err := nsq.NewConsumer(nsqTopic, nsqTopic, config)
	if err != nil {
		log.WithError(err).Error("couldn't create nsq consumer")
		return err
	}

	consumer.AddConcurrentHandlers(w, 4)
	if err := consumer.ConnectToNSQLookupds([]string{"nsqlookupd.in.opsee.com:4161"}); err != nil {
		return err
	}

	consumer.SetLogger(nsq.LogLevelError)

	return nil
}

func (w *Worker) HandleMessage(message *nsq.Message) error {
	log.WithFields(log.Fields{"worker": w.Id}).Info("Doing work...")

	result := &schema.CheckResult{}
	err := proto.Unmarshal(message.Body, result)
	if err != nil {
		log.WithError(err).Error("couldn't unmarshal checkresult")
		return err
	}

	notifications, err := w.Store.UnsafeGetNotificationsByCheckId(result.CheckId)
	if err != nil {
		log.WithError(err).Error("couldn't get notifications from the db")
		return err
	}

	if len(notifications) < 1 {
		log.Infof("no notifications found, skipping check id: %s", result.CheckId)
		return nil
	}

	event, err := hugsconsumer.BuildEvent(notifications[0], result)
	if err != nil {
		return err
	}

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
			return sendErr
		}
		log.WithFields(log.Fields{
			"customer_id": event.Result.CustomerId,
			"check_id":    event.Result.CheckId,
		}).Info(fmt.Sprintf("Sent %s notification to customer.", notification.Type))
	}

	return nil
}
