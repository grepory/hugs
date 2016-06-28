package service

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/nlopes/slack"
	"github.com/opsee/basic/schema"
	"github.com/opsee/basic/tp"
	"github.com/opsee/hugs/config"
	"github.com/opsee/hugs/notifier"
	"github.com/opsee/hugs/obj"
	log "github.com/opsee/logrus"
	"golang.org/x/net/context"
)

// get code from GET params and return token
func (s *Service) getSlackCode() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		user, ok := ctx.Value(userKey).(*schema.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errors.New("Unable to get User from request context")
		}
		request, ok := ctx.Value(requestKey).(*obj.SlackOAuthRequest)
		if !ok {
			return ctx, http.StatusBadRequest, errUnknown
		}

		// Might need to pass state as well...
		oaRequest := &obj.SlackOAuthRequest{
			ClientId:     config.GetConfig().SlackClientId,
			ClientSecret: config.GetConfig().SlackClientSecret,
			Code:         request.Code,
			RedirectURI:  request.RedirectURI,
		}

		oaResponse, err := oaRequest.Do("https://slack.com/api/oauth.access")
		if err != nil {
			log.WithFields(log.Fields{"service": "getSlackCode", "error": err}).Error("Didn't get oauth response from slack.")
			return oaResponse, http.StatusBadRequest, err
		}

		// only insert the new oauth token if it's OK
		if oaResponse.OK {
			err = s.db.PutSlackOAuthResponse(user, oaResponse)
			if err != nil {
				log.WithFields(log.Fields{"service": "getSlackCode", "error": err}).Error("Couldn't put oauth response received from slack.")
				return ctx, http.StatusInternalServerError, err
			}
		}

		return oaResponse, http.StatusOK, nil
	}
}

func (s *Service) postSlackTest() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		user, ok := ctx.Value(userKey).(*schema.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errors.New("Unable to get User from request context")
		}

		slackSender, err := notifier.NewSlackBotSender(testResultCache{})
		if err != nil {
			log.WithFields(log.Fields{"service": "postSlackTest"}).Error("Couldn't get slack sender.")
			return ctx, http.StatusBadRequest, errUnknown
		}

		request, ok := ctx.Value(requestKey).(*obj.Notifications)
		if !ok {
			return ctx, http.StatusBadRequest, errUnknown
		}

		if len(request.Notifications) < 1 {
			log.WithFields(log.Fields{"service": "postSlackTest"}).Error("Invalid notification")
			return ctx, http.StatusBadRequest, fmt.Errorf("Must have at least one notification")
		}

		event := obj.GenerateTestEvent()
		request.Notifications[0].CustomerId = user.CustomerId

		err = slackSender.Send(request.Notifications[0], event)
		if err != nil {
			log.WithFields(log.Fields{"service": "postSlackTest", "error": err}).Error("Error sending notification to slack")
			return ctx, http.StatusBadRequest, err
		}

		return nil, http.StatusOK, nil
	}
}

// Gets users slack token from db, then gets channels from API.
// TODO(dan) maybe store them in case we can't connect to slack.
func (s *Service) getSlackChannels() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		user, ok := ctx.Value(userKey).(*schema.User)
		if !ok {
			return nil, http.StatusUnauthorized, errors.New("Unable to get User from request context")
		}

		oaResponse, err := s.db.GetSlackOAuthResponse(user)
		if err != nil {
			log.WithFields(log.Fields{"service": "getSlackChannels", "error": err}).Error("Didn't get oauth response from database.")
			return nil, http.StatusBadRequest, err
		}
		if oaResponse == nil || oaResponse.Bot == nil {
			return nil, http.StatusNotFound, nil
		}

		api := slack.New(oaResponse.Bot.BotAccessToken)
		channels, err := api.GetChannels(true)
		if err != nil {
			log.WithFields(log.Fields{"service": "getSlackChannels", "error": err}).Error("Couldn't get channels from slack.")
			return nil, http.StatusBadRequest, err
		}

		respChannels := []*obj.SlackChannel{}
		for _, channel := range channels {
			slackChan := &obj.SlackChannel{
				Id:   channel.ID,
				Name: channel.Name,
			}
			respChannels = append(respChannels, slackChan)
		}
		response := &obj.SlackChannels{
			Channels: respChannels,
		}

		return response, http.StatusOK, nil
	}
}

// Fetch slack token from database, check to see if the token is active
func (s *Service) getSlackToken() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		user, ok := ctx.Value(userKey).(*schema.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errors.New("Unable to get User from request context")
		}

		oaResponse, err := s.db.GetSlackOAuthResponse(user)
		if err != nil {
			log.WithFields(log.Fields{"service": "getSlackToken", "error": err}).Error("Didn't get oauth response from database.")
			return nil, http.StatusInternalServerError, err
		}

		if oaResponse == nil || oaResponse.Bot == nil {
			return nil, http.StatusNotFound, fmt.Errorf("integration_inactive")
		}

		// confirm that the token is good and set team_domain
		// NOTE: the team_domain can change. nbd if we fail to save the new one.
		api := slack.New(oaResponse.Bot.BotAccessToken)
		teamInfo, err := api.GetTeamInfo()
		if err != nil {
			// case in which slack integration has been deactivated
			if err.Error() == "invalid_auth" || err.Error() == "account_inactive" {
				log.WithError(err).Error("Slack integration is inactive")
				return nil, http.StatusBadRequest, fmt.Errorf("integration_inactive")
			}
			// case in which we have no TeamDomain and cant get one, pass through slack err
			if oaResponse.TeamDomain == "" {
				log.WithError(err).Error("Failed to get team_domain from slack")
				return nil, http.StatusBadRequest, err
			}
			log.WithError(err).Warn("Couldn't get team_domain from slack.")
		} else {
			oaResponse.TeamDomain = teamInfo.Domain
			if oaResponse.TeamDomain != teamInfo.Domain {
				err = s.db.PutSlackOAuthResponse(user, oaResponse)
				if err != nil {
					log.WithError(err).Error("Couldn't write team info to database.")
				}
			}
		}

		return oaResponse, http.StatusOK, nil
	}
}

func (s *Service) postSlackCode() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		user, ok := ctx.Value(userKey).(*schema.User)
		if !ok {
			return nil, http.StatusUnauthorized, errors.New("Unable to get User from request context")
		}

		request, ok := ctx.Value(requestKey).(*obj.SlackOAuthRequest)
		if !ok {
			return nil, http.StatusBadRequest, errUnknown
		}

		oaRequest := &obj.SlackOAuthRequest{
			ClientId:     config.GetConfig().SlackClientId,
			ClientSecret: config.GetConfig().SlackClientSecret,
			Code:         request.Code,
			RedirectURI:  request.RedirectURI,
		}

		oaResponse, err := oaRequest.Do("https://slack.com/api/oauth.access")
		if err != nil {
			log.WithFields(log.Fields{"service": "postSlackCode", "error": err}).Error("Couldn't get oauth response from slack.")
			return nil, http.StatusBadRequest, err
		}

		if err = oaResponse.Validate(); err != nil {
			return nil, http.StatusBadRequest, err
		}

		// NOTE: If we fail to get the team domain, don't worry about it.  We will refetch on GET later
		api := slack.New(oaResponse.Bot.BotAccessToken)
		teamInfo, err := api.GetTeamInfo()
		if err != nil {
			log.WithError(err).Warn("Couldn't get team_domain from slack during integration creation.")
		} else {
			oaResponse.TeamDomain = teamInfo.Domain
		}

		err = s.db.PutSlackOAuthResponse(user, oaResponse)
		if err != nil {
			log.WithFields(log.Fields{"service": "postSlackCode", "error": err}).Error("Couldn't write slack oauth response to database.")
			return nil, http.StatusBadRequest, err
		}

		return oaResponse, http.StatusOK, nil
	}
}
