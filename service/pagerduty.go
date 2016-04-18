package service

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/opsee/basic/com"
	"github.com/opsee/basic/tp"
	"github.com/opsee/hugs/notifier"
	"github.com/opsee/hugs/obj"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

func (s *Service) postPagerDutyTest() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		user, ok := ctx.Value(userKey).(*com.User)
		if !ok {
			return nil, http.StatusUnauthorized, errors.New("Unable to get User from request context")
		}

		pdSender, err := notifier.NewPagerDutySender()
		if err != nil {
			log.WithError(err).Error("Couldn't get pagerduty sender")
			return nil, http.StatusInternalServerError, errUnknown
		}

		request, ok := ctx.Value(requestKey).(*obj.Notifications)
		if !ok {
			return nil, http.StatusBadRequest, errUnknown
		}

		if len(request.Notifications) < 1 {
			log.Error("Invalid pagerduty test notification received from emissary")
			return nil, http.StatusBadRequest, fmt.Errorf("Must have at least one notification")
		}

		event := obj.GenerateFailingTestEvent()
		request.Notifications[0].CustomerID = user.CustomerID

		err = pdSender.Send(request.Notifications[0], event)
		if err != nil {
			log.WithError(err).Error("Error sending pagerduty notification")
			return nil, http.StatusBadRequest, err
		}

		return nil, http.StatusOK, nil
	}
}

// Fetch slack token from database, check to see if the token is active
func (s *Service) getPagerDutyToken() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		user, ok := ctx.Value(userKey).(*com.User)
		if !ok {
			return nil, http.StatusUnauthorized, errors.New("Unable to get User from request context")
		}

		oaResponse, err := s.db.GetPagerDutyOAuthResponse(user)
		if err != nil {
			log.WithError(err).Error("Didn't get oauth response from database.")
			return nil, http.StatusOK, fmt.Errorf("integration_inactive")
		}
		if oaResponse == nil {
			return nil, http.StatusOK, fmt.Errorf("integration_inactive")
		}

		return oaResponse, http.StatusOK, nil
	}
}

func (s *Service) postPagerDutyCode() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		user, ok := ctx.Value(userKey).(*com.User)
		if !ok {
			return nil, http.StatusUnauthorized, errors.New("Unable to get User from request context")
		}

		oaResponse, ok := ctx.Value(requestKey).(*obj.PagerDutyOAuthResponse)
		if !ok {
			return nil, http.StatusBadRequest, errUnknown
		}

		// allow the endpoint to be reused to disable pagerduty.
		if oaResponse.Enabled != false {
			oaResponse.Enabled = true
		}

		if err := oaResponse.Validate(); err != nil {
			return nil, http.StatusInternalServerError, err
		}

		err := s.db.PutPagerDutyOAuthResponse(user, oaResponse)
		if err != nil {
			log.WithError(err).Error("Couldn't write pagerduty oauth response to database")
			return nil, http.StatusInternalServerError, err
		}

		return oaResponse, http.StatusOK, nil
	}
}
