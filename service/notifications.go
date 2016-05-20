package service

import (
	"errors"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/opsee/basic/schema"
	"github.com/opsee/basic/tp"
	"github.com/opsee/hugs/obj"
	"golang.org/x/net/context"
)

func (s *Service) getNotifications() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		user, ok := ctx.Value(userKey).(*schema.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errors.New("Unable to get User from request context")
		}

		notifications, err := s.db.GetNotificationsByUser(user)
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
		user, ok := ctx.Value(userKey).(*schema.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errors.New("Unable to get User from request context")
		}

		request, ok := ctx.Value(requestKey).(*obj.Notifications)
		if !ok {
			return ctx, http.StatusBadRequest, errUnknown
		}

		// Set notification userId and customerId
		for _, n := range request.Notifications {
			n.CustomerId = user.CustomerId
			n.UserId = int(user.Id)
			n.CheckId = request.CheckId
		}

		err := s.db.PutNotifications(user, request.Notifications)
		if err != nil {
			log.WithFields(log.Fields{"service": "putNotifications", "error": err}).Error("Couldn't put notifications in database.")
			return ctx, http.StatusBadRequest, err
		}

		result, err := s.db.GetNotificationsByCheckId(user, request.CheckId)
		if err != nil {
			log.WithFields(log.Fields{"service": "putNotifications", "error": err}).Error("Failed to get notifications.")
		}

		notifs := &obj.Notifications{
			CheckId:       request.CheckId,
			Notifications: result,
		}

		return notifs, http.StatusCreated, nil
	}
}

// creates or updates all notifications in the included Notifications array
func (s *Service) postNotificationsMultiCheck() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		user, ok := ctx.Value(userKey).(*schema.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errors.New("Unable to get User from request context")
		}

		responseNotificationsObjArray, ok := ctx.Value(requestKey).(*[]*obj.Notifications)
		if !ok {
			return nil, http.StatusBadRequest, errUnknown
		}
		notificationsObjArray := *responseNotificationsObjArray

		updatedNotificationsObjMap := make(map[string]*obj.Notifications)
		for _, notificationsObj := range notificationsObjArray {
			updatedNotificationsObjMap[notificationsObj.CheckId] = &obj.Notifications{CheckId: notificationsObj.CheckId}
			for _, notification := range notificationsObj.Notifications {
				notification.CustomerId = user.CustomerId
				notification.UserId = int(user.Id)
				notification.CheckId = notificationsObj.CheckId
			}
		}

		err := s.db.PutNotificationsMultiCheck(notificationsObjArray)
		if err != nil {
			log.WithError(err).Error("Couldn't post notifications in database.")
			return nil, http.StatusBadRequest, err
		}

		// return the notifications for each check in the deebee
		updatedNotificationsObjs := make([]*obj.Notifications, 0, len(updatedNotificationsObjMap))
		for checkId, updatedNotificationsObj := range updatedNotificationsObjMap {
			updatedNotificationsArray, err := s.db.GetNotificationsByCheckId(user, checkId)
			if err != nil {
				log.WithError(err).Error("Couldn't get updated list of notifications")
			}

			updatedNotificationsObj.Notifications = updatedNotificationsArray
			updatedNotificationsObjs = append(updatedNotificationsObjs, updatedNotificationsObj)
		}

		return updatedNotificationsObjs, http.StatusOK, nil
	}
}

func (s *Service) deleteNotificationsByCheckId() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		var checkId string

		user, ok := ctx.Value(userKey).(*schema.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errors.New("Unable to get user from request context.")
		}

		params, ok := ctx.Value(paramsKey).(httprouter.Params)
		if ok && params.ByName("check_id") != "" {
			checkId = params.ByName("check_id")
		}

		if checkId == "" {
			return ctx, http.StatusBadRequest, errors.New("Must specify check_id in request.")
		}

		notifications, err := s.db.GetNotificationsByCheckId(user, checkId)
		if err != nil {
			log.WithFields(log.Fields{"service": "deleteNotificationsByCheckId", "error": err}).Error("Couldn't delete notifications from database.")
			return ctx, http.StatusBadRequest, err
		}

		err = s.db.DeleteNotifications(notifications)
		if err != nil {
			log.WithFields(log.Fields{"service": "deleteNotificationsByCheckId", "error": err}).Error("Couldn't delete notifications from database.")
			return ctx, http.StatusInternalServerError, err
		}

		return nil, http.StatusOK, nil
	}
}

func (s *Service) getNotificationsByCheckId() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		var checkId string

		user, ok := ctx.Value(userKey).(*schema.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errors.New("Unable to get user from request context.")
		}

		params, ok := ctx.Value(paramsKey).(httprouter.Params)
		if ok && params.ByName("check_id") != "" {
			checkId = params.ByName("check_id")
		}

		if checkId == "" {
			return ctx, http.StatusBadRequest, errors.New("Must specify check-id in request.")
		}

		notifications, err := s.db.GetNotificationsByCheckId(user, checkId)
		if err != nil {
			log.WithFields(log.Fields{"service": "getNotificationsByCheckId", "error": err}).Error("Couldn't get notifications from database.")
			return ctx, http.StatusInternalServerError, err
		}

		return &obj.Notifications{Notifications: notifications}, http.StatusOK, nil
	}
}

func (s *Service) putNotificationsByCheckId() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		var checkId string

		user, ok := ctx.Value(userKey).(*schema.User)
		if !ok {
			return nil, http.StatusUnauthorized, errors.New("Unable to get User from request context")
		}

		params, ok := ctx.Value(paramsKey).(httprouter.Params)
		if ok && params.ByName("check_id") != "" {
			checkId = params.ByName("check_id")
		}

		if checkId == "" {
			return nil, http.StatusBadRequest, errors.New("Must specify check-id in request.")
		}

		request, ok := ctx.Value(requestKey).(*obj.Notifications)
		if !ok {
			return nil, http.StatusBadRequest, errUnknown
		}

		// First delete notifications for this check
		err := s.db.DeleteNotificationsByCheckId(user, checkId)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}

		// Set notification userId and customerId
		for _, n := range request.Notifications {
			n.CustomerId = user.CustomerId
			n.UserId = int(user.Id)
			n.CheckId = checkId
		}

		if err := s.db.PutNotifications(user, request.Notifications); err != nil {
			return nil, http.StatusInternalServerError, err
		}

		return &obj.Notifications{checkId, request.Notifications}, http.StatusCreated, nil

	}
}

func (s *Service) deleteNotifications() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		user, ok := ctx.Value(userKey).(*schema.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errors.New("Unable to get User from request context")
		}

		request, ok := ctx.Value(requestKey).(*obj.Notifications)
		if !ok {
			return ctx, http.StatusBadRequest, errUnknown
		}

		// Set notifications customer Id
		for _, notification := range request.Notifications {
			notification.CustomerId = user.CustomerId
			notification.UserId = int(user.Id)
		}

		notifications := obj.Notifications{
			Notifications: request.Notifications,
		}
		// Validate all notifications
		if err := notifications.Validate(); err != nil {
			return ctx, http.StatusInternalServerError, err
		}

		err := s.db.DeleteNotifications(notifications.Notifications)
		if err != nil {
			log.WithError(err).Error("Couldn't post notifications in database.")
			return ctx, http.StatusBadRequest, err
		}

		return nil, http.StatusOK, nil
	}
}
