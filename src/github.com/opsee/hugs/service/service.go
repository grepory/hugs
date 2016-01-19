package service

import (
	"errors"
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/julienschmidt/httprouter"
	"github.com/opsee/basic/com"
	"github.com/opsee/basic/tp"
	"github.com/opsee/hugs/config"
	"github.com/opsee/hugs/store"
)

const (
	ServiceKey = iota
	userKey
	requestKey
	paramsKey
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
		[]string{`https?://localhost:666`, `https://(\w+\.)?opsee\.com`},
	)

	rtr.Handle("GET", "/api/swagger.json", []tp.DecodeFunc{}, s.swagger())
	rtr.Handle("GET", "/notifications", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, com.User{}), tp.ParamsDecoder(paramsKey)}, s.getNotifications())
	rtr.Handle("POST", "/notifications", decoders(com.User{}, CheckNotifications{}), s.postNotifications())
	rtr.Handle("DELETE", "/notifications/:check_id", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, com.User{}), tp.ParamsDecoder(paramsKey)}, s.deleteNotificationsByCheckID())
	rtr.Handle("GET", "/notifications/:check_id", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, com.User{}), tp.ParamsDecoder(paramsKey)}, s.getNotificationsByCheckID())
	rtr.Handle("PUT", "/notifications/:check_id", decoders(com.User{}, CheckNotifications{}), s.putNotificationsByCheckID())

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
			return ctx, http.StatusInternalServerError, err
		}

		response := &CheckNotifications{Notifications: notifications}

		return response, http.StatusOK, nil
	}
}

func (s *Service) postNotifications() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		user, ok := ctx.Value(userKey).(*com.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errors.New("Unable to get User from request context")
		}

		request, ok := ctx.Value(requestKey).(*CheckNotifications)
		if !ok {
			return ctx, http.StatusBadRequest, errUnknown
		}

		err := s.db.PutNotifications(user, request.Notifications)
		if err != nil {
			return ctx, http.StatusInternalServerError, err
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
			return ctx, http.StatusInternalServerError, err
		}
		err = s.db.DeleteNotifications(user, notifications)
		if err != nil {
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
			return ctx, http.StatusInternalServerError, err
		}

		response := &CheckNotifications{Notifications: notifications}

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

		request, ok := ctx.Value(requestKey).(*CheckNotifications)
		if !ok {
			return ctx, http.StatusBadRequest, errUnknown
		}

		for _, n := range request.Notifications {
			n.CheckID = checkID
		}

		err := s.db.PutNotifications(user, request.Notifications)
		if err != nil {
			return ctx, http.StatusInternalServerError, err
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
