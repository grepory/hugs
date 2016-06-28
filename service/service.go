package service

import (
	"errors"
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/opsee/basic/schema"
	"github.com/opsee/basic/tp"
	"github.com/opsee/hugs/config"
	"github.com/opsee/hugs/notifier"
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

type testResultCache struct{}

func (tr testResultCache) Results(checkId string) (*notifier.ResultCacheItem, error) {
	return &notifier.ResultCacheItem{}, nil
}

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
	rtr.Handle("GET", "/services/slack/code", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, schema.User{}), tp.RequestDecodeFunc(requestKey, obj.SlackOAuthRequest{})}, s.getSlackCode())
	rtr.Handle("POST", "/services/slack/test", decoders(schema.User{}, obj.Notifications{}), s.postSlackTest())
	rtr.Handle("GET", "/services/slack/channels", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, schema.User{})}, s.getSlackChannels())
	rtr.Handle("POST", "/services/slack", decoders(schema.User{}, obj.SlackOAuthRequest{}), s.postSlackCode())
	rtr.Handle("GET", "/services/slack", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, schema.User{})}, s.getSlackToken())

	// pagerduty
	rtr.Handle("POST", "/services/pagerduty", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, schema.User{}), tp.RequestDecodeFunc(requestKey, obj.PagerDutyOAuthResponse{})}, s.postPagerDutyCode())
	rtr.Handle("GET", "/services/pagerduty", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, schema.User{})}, s.getPagerDutyToken())
	rtr.Handle("POST", "/services/pagerduty/test", decoders(schema.User{}, obj.Notifications{}), s.postPagerDutyTest())

	// email
	rtr.Handle("POST", "/services/email/test", decoders(schema.User{}, obj.Notifications{}), s.postEmailTest())

	// webhooks
	rtr.Handle("POST", "/services/webhook/test", decoders(schema.User{}, obj.Notifications{}), s.postWebHookTest())

	// notifications
	rtr.Handle("GET", "/notifications", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, schema.User{}), tp.ParamsDecoder(paramsKey)}, s.getNotifications())
	rtr.Handle("POST", "/notifications", decoders(schema.User{}, obj.Notifications{}), s.postNotifications())
	rtr.Handle("POST", "/notifications-default", decoders(schema.User{}, obj.Notifications{}), s.postNotificationsDefault())
	rtr.Handle("GET", "/notifications-default", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, schema.User{})}, s.getNotificationsDefault())
	rtr.Handle("POST", "/notifications-multicheck", decoders(schema.User{}, []*obj.Notifications{}), s.postNotificationsMultiCheck())
	rtr.Handle("DELETE", "/notifications", decoders(schema.User{}, obj.Notifications{}), s.deleteNotifications())
	rtr.Handle("DELETE", "/notifications/:check_id", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, schema.User{}), tp.ParamsDecoder(paramsKey)}, s.deleteNotificationsByCheckId())
	rtr.Handle("GET", "/notifications/:check_id", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, schema.User{}), tp.ParamsDecoder(paramsKey)}, s.getNotificationsByCheckId())
	rtr.Handle("PUT", "/notifications/:check_id", decoders(schema.User{}, obj.Notifications{}), s.putNotificationsByCheckId())
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
