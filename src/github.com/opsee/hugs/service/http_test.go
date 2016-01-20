package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	//"golang.org/x/net/context"

	"os"

	"github.com/opsee/basic/com"
	"github.com/opsee/basic/tp"
	"github.com/opsee/hugs/store"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func GetUserAuthToken(user *com.User) string {
	userstring := fmt.Sprintf(`{"id": %d, "customer_id": "%s", "user_id": "%s", "email": "%s", "verified": %t, "admin": %t, "active": %t}`, user.ID, user.CustomerID, user.ID, user.Email, user.Verified, user.Admin, user.Active)
	token := base64.StdEncoding.EncodeToString([]byte(userstring))
	return fmt.Sprintf("Basic %s", token)
}

func fuckitTest() {
	user := &com.User{
		ID:         13,
		CustomerID: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
		Email:      "dan@opsee.com",
		Name:       "Dan",
		Verified:   true,
		Admin:      true,
		Active:     true,
	}
	logrus.Info(GetUserAuthToken(user))
}

type ServiceTest struct {
	Service       *Service
	Router        *tp.Router
	Notifications []*store.Notification
	User          *com.User
	UserToken     string
}

func NewServiceTest() *ServiceTest {
	user := &com.User{
		ID:         13,
		CustomerID: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
		Email:      "dan@opsee.com",
		Name:       "Dan",
		Verified:   true,
		Admin:      true,
		Active:     true,
	}
	userAuthToken := GetUserAuthToken(user)

	logrus.Info(userAuthToken)
	logrus.Info("Connecting to local test store")
	db, err := store.NewPostgres(os.Getenv("HUGS_POSTGRES_CONN"))
	if err != nil {
		panic(err)
	}
	logrus.Info(db)
	//logrus.Info("Clearing local test store of notifications")
	//err = db.DeleteNotificationsByUser(user)

	if err != nil {
		logrus.Warn("Warning: Couldn't clear local test store of notifications")
	}

	service, err := NewService()
	if err != nil {
		logrus.Fatal("Failed to create service: ", err)
	}

	serviceTest := &ServiceTest{
		Service:   service,
		Router:    service.NewRouter(),
		User:      user,
		UserToken: userAuthToken,
		Notifications: []*store.Notification{
			&store.Notification{
				ID:         0,
				CustomerID: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
				UserID:     13,
				CheckID:    "00000",
				Value:      "off",
				Type:       "slack",
			},
			&store.Notification{
				ID:         1,
				CustomerID: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
				UserID:     13,
				CheckID:    "00000",
				Value:      "you",
				Type:       "email",
			},
			&store.Notification{
				ID:         2,
				CustomerID: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
				UserID:     13,
				CheckID:    "00001",
				Value:      "fuck",
				Type:       "slack",
			},
		},
	}
	serviceTest.Service.router = serviceTest.Router

	logrus.Info("Adding initial notifications to store.")
	err = serviceTest.Service.db.PutNotifications(serviceTest.User, serviceTest.Notifications)
	if err != nil {
		logrus.WithFields(logrus.Fields{"Error": err.Error()}).Error("Couldn't add initial notifications to service store.")
	}
	return serviceTest
}

var Common = NewServiceTest()

func TestGetNotifications(t *testing.T) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/notifications", Common.Service.config.PublicHost), nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", Common.UserToken)

	rw := httptest.NewRecorder()

	Common.Service.router.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusOK, rw.Code)

	var resp CheckNotifications

	err = json.Unmarshal(rw.Body.Bytes(), &resp)
	if err != nil {
		t.Fatal(err)
	}

	logrus.Info(resp)

	assert.Equal(t, 3, len(resp.Notifications))
}

func TestPostNotifications(t *testing.T) {
	cn := &CheckNotifications{
		Notifications: []*store.Notification{
			&store.Notification{
				ID:         0,
				CustomerID: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
				UserID:     13,
				CheckID:    "00002",
				Value:      "off 2",
				Type:       "email",
			}},
	}

	cnBytes, err := json.Marshal(cn)
	if err != nil {
		t.FailNow()
	}
	rdr := bytes.NewBufferString(string(cnBytes))
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/notifications", Common.Service.config.PublicHost), rdr)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", Common.UserToken)

	rw := httptest.NewRecorder()

	Common.Service.router.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusOK, rw.Code)
}

func TestGetNotificationsByCheckID(t *testing.T) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/notifications/00002", Common.Service.config.PublicHost), nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", Common.UserToken)

	rw := httptest.NewRecorder()

	Common.Service.router.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusOK, rw.Code)

	var resp CheckNotifications

	err = json.Unmarshal(rw.Body.Bytes(), &resp)
	if err != nil {
		t.Fatal(err)
	}

	logrus.Info(resp)

	assert.Equal(t, 1, len(resp.Notifications))
}

func TestPutNotification(t *testing.T) {
	cn := &CheckNotifications{
		Notifications: []*store.Notification{
			&store.Notification{
				ID:         3,
				CustomerID: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
				UserID:     13,
				CheckID:    "666",
				Value:      "off 2",
				Type:       "email",
			}},
	}

	cnBytes, err := json.Marshal(cn)
	if err != nil {
		t.FailNow()
	}

	rdr := bytes.NewBufferString(string(cnBytes))
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/notifications/666", Common.Service.config.PublicHost), rdr)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", Common.UserToken)

	rw := httptest.NewRecorder()

	Common.Service.router.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusOK, rw.Code)
}

func TestGetNotificationsByCheckID666(t *testing.T) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/notifications/666", Common.Service.config.PublicHost), nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", Common.UserToken)

	rw := httptest.NewRecorder()

	Common.Service.router.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusOK, rw.Code)

	var resp CheckNotifications

	err = json.Unmarshal(rw.Body.Bytes(), &resp)
	if err != nil {
		t.Fatal(err)
	}

	logrus.Info(resp)

	assert.Equal(t, 1, len(resp.Notifications))
}

func TestDeleteNotifications(t *testing.T) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/notifications/00002", Common.Service.config.PublicHost), nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", Common.UserToken)

	rw := httptest.NewRecorder()

	Common.Service.router.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusOK, rw.Code)
}
