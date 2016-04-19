package service

import (
	"errors"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/opsee/basic/com"
	"github.com/opsee/basic/tp"
	"github.com/opsee/hugs/obj"
	"golang.org/x/net/context"
)

func (s *Service) getNotifications() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		user, ok := ctx.Value(userKey).(*com.User)
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

// creates or updates all notifications in the included Notifications array
func (s *Service) postNotificationsMultiCheck() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		user, ok := ctx.Value(userKey).(*com.User)
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
			updatedNotificationsObjMap[notificationsObj.CheckID] = &obj.Notifications{CheckID: notificationsObj.CheckID}
			for _, notification := range notificationsObj.Notifications {
				notification.CustomerID = user.CustomerID
				notification.UserID = user.ID
				notification.CheckID = notificationsObj.CheckID
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
			updatedNotificationsArray, err := s.db.GetNotificationsByCheckID(user, checkId)
			if err != nil {
				log.WithError(err).Error("Couldn't get updated list of notifications")
			}

			updatedNotificationsObj.Notifications = updatedNotificationsArray
			updatedNotificationsObjs = append(updatedNotificationsObjs, updatedNotificationsObj)
		}

		return updatedNotificationsObjs, http.StatusOK, nil
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

		notifications, err := s.db.GetNotificationsByCheckID(user, checkID)
		if err != nil {
			log.WithFields(log.Fields{"service": "deleteNotificationsByCheckID", "error": err}).Error("Couldn't delete notifications from database.")
			return ctx, http.StatusBadRequest, err
		}

		err = s.db.DeleteNotifications(notifications)
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

		notifications, err := s.db.GetNotificationsByCheckID(user, checkID)
		if err != nil {
			log.WithFields(log.Fields{"service": "getNotificationsByCheckID", "error": err}).Error("Couldn't get notifications from database.")
			return ctx, http.StatusInternalServerError, err
		}

		return &obj.Notifications{Notifications: notifications}, http.StatusOK, nil
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

func (s *Service) deleteNotifications() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		user, ok := ctx.Value(userKey).(*com.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errors.New("Unable to get User from request context")
		}

		request, ok := ctx.Value(requestKey).(*obj.Notifications)
		if !ok {
			return ctx, http.StatusBadRequest, errUnknown
		}

		// Set notifications customer ID
		for _, notification := range request.Notifications {
			notification.CustomerID = user.CustomerID
			notification.UserID = user.ID
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
