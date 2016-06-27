package service

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/opsee/basic/schema"
	"github.com/opsee/basic/tp"
	"github.com/opsee/hugs/notifier"
	"github.com/opsee/hugs/obj"
	log "github.com/opsee/logrus"
	"golang.org/x/net/context"
)

func (s *Service) postWebHookTest() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		user, ok := ctx.Value(userKey).(*schema.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errors.New("Unable to get User from request context")
		}

		webHookSender, err := notifier.NewWebHookSender()
		if err != nil {
			log.WithError(err).Error("Couldn't get web hook sender")
			return ctx, http.StatusBadRequest, errUnknown
		}

		request, ok := ctx.Value(requestKey).(*obj.Notifications)
		if !ok {
			return ctx, http.StatusBadRequest, errUnknown
		}

		if len(request.Notifications) < 1 {
			log.WithFields(log.Fields{"service": "postWebHookTest"}).Error("Invalid notification")
			return ctx, http.StatusBadRequest, fmt.Errorf("Must have at least one notification")
		}

		event := obj.GenerateTestEvent()
		request.Notifications[0].CustomerId = user.CustomerId

		err = webHookSender.Send(request.Notifications[0], event)
		if err != nil {
			log.WithError(err).Error("Error sending notification to endpoint")
			return ctx, http.StatusBadRequest, err
		}

		return nil, http.StatusOK, nil
	}
}
