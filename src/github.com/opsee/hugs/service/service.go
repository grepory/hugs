package service

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/julienschmidt/httprouter"
	"github.com/nlopes/slack"
	"github.com/opsee/basic/com"
	"github.com/opsee/basic/tp"
	"github.com/opsee/hugs/config"
	"github.com/opsee/hugs/notifier"
	"github.com/opsee/hugs/obj"
	"github.com/opsee/hugs/store"
	log "github.com/sirupsen/logrus"
)

const (
	ServiceKey = iota
	userKey
	requestKey
	paramsKey
	queryKey
)

var (
	errUnauthorized = errors.New("unauthorized.")
	errUnknown      = errors.New("unknown error.")
)

type Service struct {
	db     *store.Postgres
	router *tp.Router
	config *config.Config
}

func (s *Service) Start() error {
	rtr := s.NewRouter()
	return http.ListenAndServe(config.GetConfig().PublicHost, rtr)
}

func (s *Service) NewRouter() *tp.Router {
	rtr := tp.NewHTTPRouter(context.Background())

	rtr.CORS(
		[]string{"GET", "POST", "DELETE", "HEAD"},
		[]string{`https?://localhost:9097`, `https://(\w+\.)?opsee\.com`},
	)

	rtr.Handle("GET", "/api/swagger.json", []tp.DecodeFunc{}, s.swagger())
	rtr.Handle("GET", "/services/slack/code", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, com.User{}), tp.RequestDecodeFunc(requestKey, obj.SlackOAuthRequest{})}, s.getSlackCode())
	rtr.Handle("POST", "/services/slack/test", decoders(com.User{}, obj.Notifications{}), s.postSlackTest())
	rtr.Handle("POST", "/services/email/test", decoders(com.User{}, obj.Notifications{}), s.postEmailTest())
	rtr.Handle("GET", "/services/slack/channels", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, com.User{})}, s.getSlackChannels())
	rtr.Handle("POST", "/services/slack", decoders(com.User{}, obj.SlackOAuthRequest{}), s.postSlackCode())
	rtr.Handle("GET", "/services/slack", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, com.User{})}, s.getSlackToken())

	rtr.Handle("GET", "/notifications", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, com.User{}), tp.ParamsDecoder(paramsKey)}, s.getNotifications())
	rtr.Handle("POST", "/notifications", decoders(com.User{}, obj.Notifications{}), s.postNotifications())
	rtr.Handle("DELETE", "/notifications/:check_id", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, com.User{}), tp.ParamsDecoder(paramsKey)}, s.deleteNotificationsByCheckID())
	rtr.Handle("GET", "/notifications/:check_id", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, com.User{}), tp.ParamsDecoder(paramsKey)}, s.getNotificationsByCheckID())
	rtr.Handle("PUT", "/notifications/:check_id", decoders(com.User{}, obj.Notifications{}), s.putNotificationsByCheckID())

	rtr.Timeout(5 * time.Minute)

	return rtr
}

func (s *Service) swagger() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		return swaggerMap, http.StatusOK, nil
	}
}

func (s *Service) getNotifications() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		user, ok := ctx.Value(userKey).(*com.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errors.New("Unable to get User from request context")
		}

		notifications, err := s.db.GetNotifications(user)
		if err != nil {
			log.WithFields(log.Fields{"service": "getNotifications", "error": err}).Error("Couldn't get notifications from database.")
			return ctx, http.StatusBadRequest, err
		}

		response := &obj.Notifications{Notifications: notifications}

		return response, http.StatusOK, nil
	}
}

func (s *Service) postNotifications() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		user, ok := ctx.Value(userKey).(*com.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errors.New("Unable to get User from request context")
		}

		request, ok := ctx.Value(requestKey).(*obj.Notifications)
		if !ok {
			return ctx, http.StatusBadRequest, errUnknown
		}

		// Set notification userID and customerID
		for _, n := range request.Notifications {
			n.CustomerID = user.CustomerID
			n.UserID = user.ID
			n.CheckID = request.CheckID
		}

		err := s.db.PutNotifications(user, request.Notifications)
		if err != nil {
			log.WithFields(log.Fields{"service": "putNotifications", "error": err}).Error("Couldn't put notifications in database.")
			return ctx, http.StatusBadRequest, err
		}

		result, err := s.db.GetNotificationsByCheckID(user, request.CheckID)
		if err != nil {
			log.WithFields(log.Fields{"service": "putNotifications", "error": err}).Error("Failed to get notifications.")
		}

		notifs := &obj.Notifications{
			CheckID:       request.CheckID,
			Notifications: result,
		}

		return notifs, http.StatusCreated, nil
	}
}

func (s *Service) deleteNotificationsByCheckID() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		var checkID string

		user, ok := ctx.Value(userKey).(*com.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errors.New("Unable to get user from request context.")
		}

		params, ok := ctx.Value(paramsKey).(httprouter.Params)
		if ok && params.ByName("check_id") != "" {
			checkID = params.ByName("check_id")
		}

		if checkID == "" {
			return ctx, http.StatusBadRequest, errors.New("Must specify check_id in request.")
		}

		// Get notifications by checkID and then call delete on each one
		notifications, err := s.db.GetNotificationsByCheckID(user, checkID)
		if err != nil {
			log.WithFields(log.Fields{"service": "deleteNotificationsByCheckID", "error": err}).Error("Couldn't delete notifications from database.")
			return ctx, http.StatusBadRequest, err
		}
		err = s.db.DeleteNotifications(user, notifications)
		if err != nil {
			log.WithFields(log.Fields{"service": "deleteNotificationsByCheckID", "error": err}).Error("Couldn't delete notifications from database.")
			return ctx, http.StatusInternalServerError, err
		}

		return nil, http.StatusOK, nil
	}
}

func (s *Service) getNotificationsByCheckID() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		var checkID string

		user, ok := ctx.Value(userKey).(*com.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errors.New("Unable to get user from request context.")
		}

		params, ok := ctx.Value(paramsKey).(httprouter.Params)
		if ok && params.ByName("check_id") != "" {
			checkID = params.ByName("check_id")
		}

		if checkID == "" {
			return ctx, http.StatusBadRequest, errors.New("Must specify check-id in request.")
		}

		// Get notifications by checkID and then call delete on each one
		notifications, err := s.db.GetNotificationsByCheckID(user, checkID)
		if err != nil {
			log.WithFields(log.Fields{"service": "getNotificationsByCheckID", "error": err}).Error("Couldn't get notifications from database.")
			return ctx, http.StatusInternalServerError, err
		}

		response := &obj.Notifications{Notifications: notifications}

		return response, http.StatusOK, nil
	}
}

func (s *Service) putNotificationsByCheckID() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		var checkID string

		user, ok := ctx.Value(userKey).(*com.User)
		if !ok {
			return nil, http.StatusUnauthorized, errors.New("Unable to get User from request context")
		}

		params, ok := ctx.Value(paramsKey).(httprouter.Params)
		if ok && params.ByName("check_id") != "" {
			checkID = params.ByName("check_id")
		}

		if checkID == "" {
			return nil, http.StatusBadRequest, errors.New("Must specify check-id in request.")
		}

		request, ok := ctx.Value(requestKey).(*obj.Notifications)
		if !ok {
			return nil, http.StatusBadRequest, errUnknown
		}

		// First delete notifications for this check
		err := s.db.DeleteNotificationsByCheckId(user, checkID)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}

		// Set notification userID and customerID
		for _, n := range request.Notifications {
			n.CustomerID = user.CustomerID
			n.UserID = user.ID
			n.CheckID = checkID
		}

		if err := s.db.PutNotifications(user, request.Notifications); err != nil {
			return nil, http.StatusInternalServerError, err
		}

		return &obj.Notifications{checkID, request.Notifications}, http.StatusCreated, nil

	}
}

// Finish the oauth flow and get token from slack.
// Save the oauth response from slack and return token to front-end
// TODO(dan) Deprecate
func (s *Service) postSlackCode() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		user, ok := ctx.Value(userKey).(*com.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errors.New("Unable to get User from request context")
		}

		request, ok := ctx.Value(requestKey).(*obj.SlackOAuthRequest)
		if !ok {
			return ctx, http.StatusBadRequest, errUnknown
		}

		oaRequest := &obj.SlackOAuthRequest{
			ClientID:     config.GetConfig().SlackClientID,
			ClientSecret: config.GetConfig().SlackClientSecret,
			Code:         request.Code,
			RedirectURI:  request.RedirectURI,
		}

		oaResponse, err := oaRequest.Do("https://slack.com/api/oauth.access")
		if err != nil {
			log.WithFields(log.Fields{"service": "postSlackCode", "error": err}).Error("Couldn't get oauth response from slack.")
			return ctx, http.StatusBadRequest, err
		}

		if err = oaResponse.Validate(); err != nil {
			return ctx, http.StatusBadRequest, err
		}

		err = s.db.PutSlackOAuthResponse(user, oaResponse)
		if err != nil {
			log.WithFields(log.Fields{"service": "postSlackCode", "error": err}).Error("Couldn't write slack oauth response to database.")
			return ctx, http.StatusBadRequest, err
		}

		return oaResponse, http.StatusOK, nil
	}
}

// Gets users slack token from db, then gets channels from API.
// TODO(dan) maybe store them in case we can't connect to slack.
func (s *Service) getSlackChannels() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		user, ok := ctx.Value(userKey).(*com.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errors.New("Unable to get User from request context")
		}

		oaResponse, err := s.db.GetSlackOAuthResponse(user)
		if err != nil {
			log.WithFields(log.Fields{"service": "getSlackChannels", "error": err}).Error("Didn't get oauth response from database.")
			return ctx, http.StatusBadRequest, err
		}
		if oaResponse == nil || oaResponse.Bot == nil {
			return ctx, http.StatusNotFound, nil
		}

		api := slack.New(oaResponse.Bot.BotAccessToken)
		channels, err := api.GetChannels(true)
		if err != nil {
			log.WithFields(log.Fields{"service": "getSlackChannels", "error": err}).Error("Couldn't get channels from slack.")
			return ctx, http.StatusBadRequest, err
		}

		respChannels := []*obj.SlackChannel{}
		for _, channel := range channels {
			slackChan := &obj.SlackChannel{
				ID:   channel.ID,
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
		user, ok := ctx.Value(userKey).(*com.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errors.New("Unable to get User from request context")
		}

		oaResponse, err := s.db.GetSlackOAuthResponse(user)
		if err != nil {
			log.WithFields(log.Fields{"service": "getSlackToken", "error": err}).Error("Didn't get oauth response from database.")
			return ctx, http.StatusInternalServerError, err
		}

		// TODO(dan) need to handle inactive tokens differently for slack bots vs webhooks.
		// For now let's assume that we require a bot token
		if oaResponse == nil || oaResponse.Bot == nil {
			return ctx, http.StatusNotFound, fmt.Errorf("integration_inactive")
		}

		// confirm that the token is good
		api := slack.New(oaResponse.Bot.BotAccessToken)
		_, err = api.AuthTest()
		if err != nil {
			return ctx, http.StatusNotFound, fmt.Errorf("integration_inactive")
		}

		return oaResponse, http.StatusOK, nil
	}
}

// get code from GET params and return token
func (s *Service) getSlackCode() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		user, ok := ctx.Value(userKey).(*com.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errors.New("Unable to get User from request context")
		}
		request, ok := ctx.Value(requestKey).(*obj.SlackOAuthRequest)
		if !ok {
			return ctx, http.StatusBadRequest, errUnknown
		}

		// Might need to pass state as well...
		oaRequest := &obj.SlackOAuthRequest{
			ClientID:     config.GetConfig().SlackClientID,
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
		user, ok := ctx.Value(userKey).(*com.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errors.New("Unable to get User from request context")
		}

		slackSender, err := notifier.NewSlackBotSender()
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
		request.Notifications[0].CustomerID = user.CustomerID

		err = slackSender.Send(request.Notifications[0], event)
		if err != nil {
			log.WithFields(log.Fields{"service": "postSlackTest", "error": err}).Error("Error sending notification to slack")
			return ctx, http.StatusBadRequest, err
		}

		return nil, http.StatusOK, nil
	}
}

func (s *Service) postEmailTest() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		user, ok := ctx.Value(userKey).(*com.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errors.New("Unable to get User from request context")
		}

		emailSender, err := notifier.NewEmailSender(config.GetConfig().OpseeHost, config.GetConfig().MandrillApiKey)
		if err != nil {
			log.WithFields(log.Fields{"service": "postEmailTest"}).Error("Couldn't get email sender.")
			return ctx, http.StatusBadRequest, errUnknown
		}

		request, ok := ctx.Value(requestKey).(*obj.Notifications)
		if !ok {
			return ctx, http.StatusBadRequest, errUnknown
		}

		if len(request.Notifications) < 1 {
			log.WithFields(log.Fields{"service": "postEmailTest"}).Error("Invalid notification")
			return ctx, http.StatusBadRequest, fmt.Errorf("Must have at least one notification")
		}

		event := obj.GenerateTestEvent()
		request.Notifications[0].CustomerID = user.CustomerID

		err = emailSender.Send(request.Notifications[0], event)
		if err != nil {
			log.WithFields(log.Fields{"service": "postEmailTest", "error": err}).Error("Error sending notification via email.")
			return ctx, http.StatusBadRequest, err
		}

		return nil, http.StatusOK, nil
	}
}

func decoders(userType interface{}, requestType interface{}) []tp.DecodeFunc {
	return []tp.DecodeFunc{
		tp.AuthorizationDecodeFunc(userKey, userType),
		tp.RequestDecodeFunc(requestKey, requestType),
		tp.ParamsDecoder(paramsKey),
	}
}

func NewService() (*Service, error) {
	dbmaybe, err := store.NewPostgres(config.GetConfig().PostgresConn)
	if err != nil {
		return nil, err
	}
	return &Service{
		db:     dbmaybe,
		config: config.GetConfig(),
	}, nil
}
