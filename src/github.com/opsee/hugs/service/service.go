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
	"github.com/opsee/hugs/apiutils"
	"github.com/opsee/hugs/config"
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
	rtr.Handle("GET", "/notifs", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, com.User{}), tp.ParamsDecoder(paramsKey)}, s.getNotifications())
	rtr.Handle("POST", "/notifs", decoders(com.User{}, obj.Notifications{}), s.postNotifications())
	rtr.Handle("DELETE", "/notifs/:check_id", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, com.User{}), tp.ParamsDecoder(paramsKey)}, s.deleteNotificationsByCheckID())
	rtr.Handle("GET", "/notifs/:check_id", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, com.User{}), tp.ParamsDecoder(paramsKey)}, s.getNotificationsByCheckID())
	rtr.Handle("PUT", "/notifs/:check_id", decoders(com.User{}, obj.Notifications{}), s.putNotificationsByCheckID())
	rtr.Handle("POST", "/services/slack", decoders(com.User{}, obj.SlackOAuthRequest{}), s.postSlackCode())
	rtr.Handle("GET", "/services/slack", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, com.User{})}, s.getSlackToken())
	rtr.Handle("GET", "/services/slack/channels", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, com.User{})}, s.getSlackChannels())
	rtr.Handle("GET", "/services/slack/code", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, com.User{}), tp.RequestDecodeFunc(requestKey, obj.SlackOAuthRequest{})}, s.getSlackCode())

	rtr.Handle("GET", "/services/slack/test/code", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, com.User{}), tp.RequestDecodeFunc(requestKey, obj.SlackOAuthRequest{})}, s.getSlackCodeTest())

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

		err := s.db.PutNotifications(user, request.Notifications)
		if err != nil {
			log.WithFields(log.Fields{"service": "putNotifications", "error": err}).Error("Couldn't put notifications in database.")
			return ctx, http.StatusBadRequest, err
		}

		return ctx, http.StatusOK, nil
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
			return ctx, http.StatusBadRequest, errors.New("Must specify check-id in request.")
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
			return ctx, http.StatusUnauthorized, errors.New("Unable to get User from request context")
		}

		params, ok := ctx.Value(paramsKey).(httprouter.Params)
		if ok && params.ByName("check_id") != "" {
			checkID = params.ByName("check_id")
		}

		if checkID == "" {
			return ctx, http.StatusBadRequest, errors.New("Must specify check-id in request.")
		}

		request, ok := ctx.Value(requestKey).(*obj.Notifications)
		if !ok {
			return ctx, http.StatusBadRequest, errUnknown
		}

		for _, n := range request.Notifications {
			n.CheckID = checkID
		}

		err := s.db.PutNotifications(user, request.Notifications)
		if err != nil {
			return ctx, http.StatusBadRequest, err
		}

		return nil, http.StatusOK, nil

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

		oaResponse, err := oaRequest.Do(apiutils.SlackOAuthEndpoint)
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

		oaResponse, err := oaRequest.Do(apiutils.SlackOAuthEndpoint)
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

// get code from GET params and return token
func (s *Service) getSlackCodeTest() tp.HandleFunc {
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
			ClientID:     config.GetConfig().SlackTestClientID,
			ClientSecret: config.GetConfig().SlackTestClientSecret,
			Code:         request.Code,
			RedirectURI:  request.RedirectURI,
		}

		oaResponse, err := oaRequest.Do(apiutils.SlackOAuthEndpoint)
		if err != nil {
			log.WithFields(log.Fields{"service": "getSlackCode", "error": err}).Error("Didn't get oauth response from slack.")
			return oaResponse, http.StatusBadRequest, err
		}

		err = s.db.PutSlackOAuthResponse(user, oaResponse)
		if err != nil {
			log.WithFields(log.Fields{"service": "getSlackCode", "error": err}).Error("Couldn't put oauth response received from slack.")
			return ctx, http.StatusInternalServerError, err
		}

		return oaResponse, http.StatusOK, nil
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
