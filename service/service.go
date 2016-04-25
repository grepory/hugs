package service

import (
	"errors"
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/opsee/basic/com"
	"github.com/opsee/basic/tp"
	"github.com/opsee/hugs/config"
	"github.com/opsee/hugs/obj"
	"github.com/opsee/hugs/store"
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
	// swagger
	rtr.Handle("GET", "/api/swagger.json", []tp.DecodeFunc{}, s.swagger())

	// slack
	rtr.Handle("GET", "/services/slack/code", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, com.User{}), tp.RequestDecodeFunc(requestKey, obj.SlackOAuthRequest{})}, s.getSlackCode())
	rtr.Handle("POST", "/services/slack/test", decoders(com.User{}, obj.Notifications{}), s.postSlackTest())
	rtr.Handle("GET", "/services/slack/channels", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, com.User{})}, s.getSlackChannels())
	rtr.Handle("POST", "/services/slack", decoders(com.User{}, obj.SlackOAuthRequest{}), s.postSlackCode())
	rtr.Handle("GET", "/services/slack", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, com.User{})}, s.getSlackToken())

	// pagerduty
	rtr.Handle("POST", "/services/pagerduty", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, com.User{}), tp.RequestDecodeFunc(requestKey, obj.PagerDutyOAuthResponse{})}, s.postPagerDutyCode())
	rtr.Handle("GET", "/services/pagerduty", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, com.User{})}, s.getPagerDutyToken())
	rtr.Handle("POST", "/services/pagerduty/test", decoders(com.User{}, obj.Notifications{}), s.postPagerDutyTest())

	// email
	rtr.Handle("POST", "/services/email/test", decoders(com.User{}, obj.Notifications{}), s.postEmailTest())

	// webhooks
	rtr.Handle("POST", "/services/webhook/test", decoders(com.User{}, obj.Notifications{}), s.postWebHookTest())

	// notifications
	rtr.Handle("GET", "/notifications", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, com.User{}), tp.ParamsDecoder(paramsKey)}, s.getNotifications())
	rtr.Handle("POST", "/notifications", decoders(com.User{}, obj.Notifications{}), s.postNotifications())
	rtr.Handle("POST", "/notifications-multicheck", decoders(com.User{}, []*obj.Notifications{}), s.postNotificationsMultiCheck())
	rtr.Handle("DELETE", "/notifications", decoders(com.User{}, obj.Notifications{}), s.deleteNotifications())
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

func decoders(userType interface{}, requestType interface{}) []tp.DecodeFunc {
	return []tp.DecodeFunc{
		tp.AuthorizationDecodeFunc(userKey, userType),
		tp.RequestDecodeFunc(requestKey, requestType),
		tp.ParamsDecoder(paramsKey),
	}
}

func NewService() (*Service, error) {
	dbmaybe, err := store.NewPostgres()
	if err != nil {
		return nil, err
	}
	return &Service{
		db:     dbmaybe,
		config: config.GetConfig(),
	}, nil
}
