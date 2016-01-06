package service

import (
	"errors"

	"github.com/opsee/basic/com"
	"github.com/opsee/basic/tp"
)

const (
	serviceKey = iota
	userKey
	requestKey
	paramsKey
)

func (s *service) StartHTTP(addr string) {
	router := tp.NewHTTPRouter(context.Background())

	router.CORS(
		[]string{"GET", "POST", "DELETE", "HEAD"},
		[]string{`https?://localhost:8080`, `https://(\w+\.)?opsee\.com`},
	)

	router.Handle("GET", "/api/swagger.json", []tp.DecodeFunc{}, s.swagger())

	router.Handle("GET", "/notifications", decoders(com.User{}), s.listNotifications())
	//router.Handle("GET", "/notification/:id",
	//router.Handle("POST", "/notification",
}

func (s *service) listNotifications() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		user, ok := ctx.Value(userKey).(*com.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errors.New("Unable to get User from request context")
		}

		notifications, err := s.ListNotifications(user)
		if err != nil {
			return ctx, http.StatusInternalServerError, err
		}

		return notifications, http.StatusOK, nil
	}
}

func decoders(userType interface{}) []tp.DecodeFunc {
	return []tp.DecodeFunc{
		tp.AuthorizationDecodeFunc{userKey, userType},
		tp.ParamsDecoder(paramsKey),
	}
}
