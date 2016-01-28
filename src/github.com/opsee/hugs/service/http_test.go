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
	"github.com/opsee/hugs/apiutils"
	"github.com/opsee/hugs/config"
	"github.com/opsee/hugs/obj"
	"github.com/opsee/hugs/store"
	log "github.com/sirupsen/logrus"
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
	log.Info(GetUserAuthToken(user))
}

type ServiceTest struct {
	Service       *Service
	Router        *tp.Router
	Notifications []*obj.Notification
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

	log.Info(userAuthToken)
	log.Info("Connecting to local test obj")
	db, err := store.NewPostgres(os.Getenv("HUGS_POSTGRES_CONN"))
	if err != nil {
		panic(err)
	}
	log.Info(db)
	//log.Info("Clearing local test obj of notifications")
	//err = db.DeleteNotificationsByUser(user)

	if err != nil {
		log.Warn("Warning: Couldn't clear local test obj of notifications")
	}

	service, err := NewService()
	if err != nil {
		log.Fatal("Failed to create service: ", err)
	}

	serviceTest := &ServiceTest{
		Service:   service,
		Router:    service.NewRouter(),
		User:      user,
		UserToken: userAuthToken,
		Notifications: []*obj.Notification{
			&obj.Notification{
				ID:         0,
				CustomerID: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
				UserID:     13,
				CheckID:    "00000",
				Value:      "C0ATUFZ7X", // this a channel
				Type:       "slack_bot",
			},
			&obj.Notification{
				ID:         1,
				CustomerID: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
				UserID:     13,
				CheckID:    "00000",
				Value:      "dan@opsee.com",
				Type:       "email",
			},
			&obj.Notification{
				ID:         2,
				CustomerID: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
				UserID:     13,
				CheckID:    "00000",
				Value:      "someslackhook.com",
				Type:       "slack_hook",
			},
		},
	}

	serviceTest.Service.router = serviceTest.Router
	log.Info("Starting slack api emulator...")
	go apiutils.StartSlackAPIEmulator()

	log.Info("Adding initial notifications to obj...")
	err = serviceTest.Service.db.PutNotifications(serviceTest.User, serviceTest.Notifications)
	if err != nil {
		log.WithFields(log.Fields{"Error": err.Error()}).Error("Couldn't add initial notifications to service obj.")
	}

	log.Info("Adding initial slack oauth shit to obj...")
	slackOAuthResponse := &obj.SlackOAuthResponse{
		AccessToken: config.GetConfig().SlackTestToken,
		Scope:       "bot",
		TeamName:    "opsee",
		TeamID:      "opsee",
		IncomingWebhook: &obj.SlackIncomingWebhook{
			URL:              "test",
			Channel:          "test",
			ConfigurationURL: "test",
		},
		Bot: &obj.SlackBotCreds{
			BotUserID:      "test",
			BotAccessToken: "test",
		},
	}

	err = serviceTest.Service.db.PutSlackOAuthResponse(serviceTest.User, slackOAuthResponse)
	if err != nil {
		log.Fatal(err)
	}

	return serviceTest
}

var Common = NewServiceTest()

func TestGetNotifications(t *testing.T) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/notifs", Common.Service.config.PublicHost), nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", Common.UserToken)

	rw := httptest.NewRecorder()

	Common.Service.router.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusOK, rw.Code)

	var resp obj.Notifications

	err = json.Unmarshal(rw.Body.Bytes(), &resp)
	if err != nil {
		t.Fatal(err)
	}

	log.Info(resp)
	if len(resp.Notifications) == 0 {
		t.FailNow()
	}
}

func TestPostNotifications(t *testing.T) {
	cn := &obj.Notifications{
		Notifications: []*obj.Notification{
			&obj.Notification{
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
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/notifs", Common.Service.config.PublicHost), rdr)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", Common.UserToken)

	rw := httptest.NewRecorder()

	Common.Service.router.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusOK, rw.Code)
}

func TestGetNotificationsByCheckID(t *testing.T) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/notifs/00002", Common.Service.config.PublicHost), nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", Common.UserToken)

	rw := httptest.NewRecorder()

	Common.Service.router.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusOK, rw.Code)

	var resp obj.Notifications

	err = json.Unmarshal(rw.Body.Bytes(), &resp)
	if err != nil {
		t.Fatal(err)
	}

	log.Info(resp)

	assert.Equal(t, 1, len(resp.Notifications))
}

func TestPutNotification(t *testing.T) {
	cn := &obj.Notifications{
		Notifications: []*obj.Notification{
			&obj.Notification{
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
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/notifs/666", Common.Service.config.PublicHost), rdr)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", Common.UserToken)

	rw := httptest.NewRecorder()

	Common.Service.router.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusOK, rw.Code)
}

func TestGetNotificationsByCheckID666(t *testing.T) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/notifs/666", Common.Service.config.PublicHost), nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", Common.UserToken)

	rw := httptest.NewRecorder()

	Common.Service.router.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusOK, rw.Code)

	var resp obj.Notifications

	err = json.Unmarshal(rw.Body.Bytes(), &resp)
	if err != nil {
		t.Fatal(err)
	}

	log.Info(resp)
	if len(resp.Notifications) == 0 {
		t.FailNow()
	}
}

func TestDeleteNotifications(t *testing.T) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/notifs/00002", Common.Service.config.PublicHost), nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", Common.UserToken)

	rw := httptest.NewRecorder()

	Common.Service.router.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusOK, rw.Code)
}

func TestGetSlackChannels(t *testing.T) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/services/slack/channels", Common.Service.config.PublicHost), nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", Common.UserToken)

	rw := httptest.NewRecorder()
	Common.Service.router.ServeHTTP(rw, req)
	log.WithFields(log.Fields{"TestGetSlackChannels": "Got channel list."}).Info(rw.Body)

	assert.Equal(t, http.StatusOK, rw.Code)
}

func TestGetSlackToken(t *testing.T) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/services/slack", Common.Service.config.PublicHost), nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", Common.UserToken)

	rw := httptest.NewRecorder()

	Common.Service.router.ServeHTTP(rw, req)

	var resp obj.SlackOAuthResponse

	err = json.Unmarshal(rw.Body.Bytes(), &resp)
	if err != nil {
		t.Fatal(err)
	}

	log.WithFields(log.Fields{"TestGetSlackToken": "Got slack token."}).Info(resp)
	assert.Equal(t, http.StatusOK, rw.Code)
}

// Note that this should fail because code will be invalid.
func TestPostSlackCode(t *testing.T) {
	oar := &obj.SlackOAuthRequest{
		Code:        "test",
		RedirectURI: "test",
	}

	oarbits, err := json.Marshal(oar)
	if err != nil {
		t.FailNow()
	}

	rdr := bytes.NewBufferString(string(oarbits))
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/services/slack", Common.Service.config.PublicHost), rdr)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", Common.UserToken)

	rw := httptest.NewRecorder()

	Common.Service.router.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusInternalServerError, rw.Code)
}
