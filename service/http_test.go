package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	//"golang.org/x/net/context"

	"github.com/opsee/basic/schema"
	"github.com/opsee/basic/tp"
	"github.com/opsee/hugs/config"
	"github.com/opsee/hugs/obj"
	"github.com/opsee/hugs/store"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func webhooktest(rw http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error(err)
	} else {
		log.Info("Test webhook endpoint got: ", string(body))
	}
}

// emulates slack api endpoints, provides endpoint for /services/webhook/test
func startTestServer() {
	oaResponse := &obj.SlackOAuthResponse{
		AccessToken: "test",
		Scope:       "test",
		TeamName:    "test",
		TeamId:      "test",
		IncomingWebhook: &obj.SlackIncomingWebhook{
			URL:              "test",
			Channel:          "test",
			ConfigurationURL: "test",
		},
		Bot: &obj.SlackBotCreds{
			BotUserId:      "test",
			BotAccessToken: "test",
		},
	}

	oaResponseData, err := json.Marshal(oaResponse)
	if err != nil {
		log.Fatal(err)
	}
	// emulate slack oauth endpoint
	http.HandleFunc("/api/oauth.access", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, string(oaResponseData), r.URL.Path)
	})
	http.HandleFunc("/hook", webhooktest)

	log.Fatal(http.ListenAndServe(":7766", nil))
}

func GetUserAuthToken(user *schema.User) string {
	userstring := fmt.Sprintf(`{"id": %d, "customer_id": "%s", "user_id": "%s", "email": "%s", "verified": %t, "admin": %t, "active": %t}`, user.Id, user.CustomerId, user.Id, user.Email, user.Verified, user.Admin, user.Active)
	token := base64.StdEncoding.EncodeToString([]byte(userstring))
	return fmt.Sprintf("Basic %s", token)
}

func fuckitTest() {
	user := &schema.User{
		Id:         13,
		CustomerId: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
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
	User          *schema.User
	UserToken     string
}

func NewServiceTest() *ServiceTest {
	user := &schema.User{
		Id:         13,
		CustomerId: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
		Email:      "dan@opsee.com",
		Name:       "Dan",
		Verified:   true,
		Admin:      true,
		Active:     true,
	}
	userAuthToken := GetUserAuthToken(user)

	log.Info(userAuthToken)
	log.Info("Connecting to local test obj")
	db, err := store.NewPostgres()
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
				Id:         0,
				CustomerId: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
				UserId:     13,
				CheckId:    "00000",
				Value:      "C0ATUFZ7X", // this a channel
				Type:       "slack_bot",
			},
			&obj.Notification{
				Id:         1,
				CustomerId: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
				UserId:     13,
				CheckId:    "00000",
				Value:      "dan@opsee.com",
				Type:       "email",
			},
			&obj.Notification{
				Id:         2,
				CustomerId: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
				UserId:     13,
				CheckId:    "00000",
				Value:      "someslackhook.com",
				Type:       "webhook",
			},
		},
	}

	serviceTest.Service.router = serviceTest.Router
	log.Info("Starting slack api emulator...")
	go startTestServer()

	log.Info("Adding initial notifications to obj...")
	err = serviceTest.Service.db.PutNotifications(user, serviceTest.Notifications)
	if err != nil {
		log.WithFields(log.Fields{"Error": err.Error()}).Error("Couldn't add initial notifications to service obj.")
	}

	log.Info("Adding initial slack oauth shit to obj...")
	slackOAuthResponse := &obj.SlackOAuthResponse{
		AccessToken: config.GetConfig().SlackTestToken,
		Scope:       "bot",
		TeamName:    "opsee",
		TeamId:      "opsee",
		TeamDomain:  "opsee",
		IncomingWebhook: &obj.SlackIncomingWebhook{
			URL:              "test",
			Channel:          "test",
			ConfigurationURL: "test",
		},
		Bot: &obj.SlackBotCreds{
			BotUserId:      "test",
			BotAccessToken: config.GetConfig().SlackTestToken,
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
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/notifications", Common.Service.config.PublicHost), nil)
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

// test inserting/updating notifications for multiple checks
func TestPostNotificationsMultiCheck(t *testing.T) {
	cn := []*obj.Notifications{
		&obj.Notifications{
			CheckId: "TestMultiEdit0",
			Notifications: []*obj.Notification{
				&obj.Notification{
					Value: "off 2",
					Type:  "email",
				},
				&obj.Notification{
					CheckId: "TestMultiEdit0",
					Value:   "off 2",
					Type:    "email",
				},
			},
		},
		&obj.Notifications{
			CheckId: "TestMultiEdit1",
			Notifications: []*obj.Notification{
				&obj.Notification{
					Value: "off 2",
					Type:  "email",
				},
				&obj.Notification{
					Value: "off 2",
					Type:  "email",
				},
			},
		},
	}

	cnBytes, err := json.Marshal(cn)
	if err != nil {
		t.Fatal(err)
	}

	rdr := bytes.NewBufferString(string(cnBytes))
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/notifications-multicheck", Common.Service.config.PublicHost), rdr)
	if err != nil {
		log.WithError(err).Error("%v", err)
		t.Fatal(err)
	}
	req.Header.Set("Authorization", Common.UserToken)

	rw := httptest.NewRecorder()

	Common.Service.router.ServeHTTP(rw, req)
	bytes := rw.Body.Bytes()

	notificationsObjs := []*obj.Notification{}
	err = json.Unmarshal(bytes, &notificationsObjs)
	if err != nil {
		t.Fatal(err)
	}

	if len(notificationsObjs) == 0 {
		t.FailNow()
	}

	assert.Equal(t, http.StatusOK, rw.Code)
}

func TestPostNotifications(t *testing.T) {
	cn := &obj.Notifications{
		Notifications: []*obj.Notification{
			&obj.Notification{
				Id:         99,
				CustomerId: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
				UserId:     13,
				CheckId:    "00002",
				Value:      "off 2",
				Type:       "email",
			},
			&obj.Notification{
				Id:         98,
				CustomerId: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
				UserId:     13,
				CheckId:    "00002",
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

	var resp obj.Notifications

	err = json.Unmarshal(rw.Body.Bytes(), &resp)
	if err != nil {
		t.Fatal(err)
	}

	log.Info(resp)
	if len(resp.Notifications) == 0 {
		t.FailNow()
	}

	assert.Equal(t, len(cn.Notifications), len(resp.Notifications))
	assert.Equal(t, http.StatusCreated, rw.Code)
}

func TestPutNotification(t *testing.T) {
	cn := &obj.Notifications{
		Notifications: []*obj.Notification{
			&obj.Notification{
				Id:         9999,
				CustomerId: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
				UserId:     13,
				CheckId:    "666",
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
	assert.Equal(t, http.StatusCreated, rw.Code)
}

func TestGetNotificationsByCheckId(t *testing.T) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/notifications/666", Common.Service.config.PublicHost), nil)
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

func TestDeleteNotification(t *testing.T) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/notifications/00002", Common.Service.config.PublicHost), nil)
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
	assert.Equal(t, http.StatusBadRequest, rw.Code)
}

// Test posting a message to slack
/*
func TestPostSlackTest(t *testing.T) {
	cn := &obj.Notifications{
		Notifications: []*obj.Notification{
			&obj.Notification{
				Id:         0,
				CustomerId: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
				UserId:     13,
				CheckId:    "00002",
				Value:      "C0ADACATT",
				Type:       "slack_bot",
			}},
	}

	notifs, err := json.Marshal(cn)
	if err != nil {
		t.FailNow()
	}

	rdr := bytes.NewBufferString(string(notifs))
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/services/slack/test", Common.Service.config.PublicHost), rdr)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", Common.UserToken)

	rw := httptest.NewRecorder()

	Common.Service.router.ServeHTTP(rw, req)
	log.Info(string(rw.Body.Bytes()))
	assert.Equal(t, http.StatusOK, rw.Code)
}

func TestPostEmailTest(t *testing.T) {
	cn := &obj.Notifications{
		Notifications: []*obj.Notification{
			&obj.Notification{
				Id:         0,
				CustomerId: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
				UserId:     13,
				CheckId:    "00002",
				Value:      "dan@opsee.com",
				Type:       "email",
			}},
	}

	notifs, err := json.Marshal(cn)
	if err != nil {
		t.FailNow()
	}

	rdr := bytes.NewBufferString(string(notifs))
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/services/email/test", Common.Service.config.PublicHost), rdr)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", Common.UserToken)

	rw := httptest.NewRecorder()

	Common.Service.router.ServeHTTP(rw, req)
	log.Info(string(rw.Body.Bytes()))
	assert.Equal(t, http.StatusOK, rw.Code)
}
*/
func TestPostWebHookTest(t *testing.T) {
	cn := &obj.Notifications{
		Notifications: []*obj.Notification{
			&obj.Notification{
				Id:         1,
				CustomerId: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
				UserId:     13,
				CheckId:    "00002",
				Value:      "http://localhost:7766/hook",
				Type:       "webhook",
			}},
	}

	notifs, err := json.Marshal(cn)
	if err != nil {
		t.FailNow()
	}

	rdr := bytes.NewBufferString(string(notifs))
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/services/webhook/test", Common.Service.config.PublicHost), rdr)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", Common.UserToken)

	rw := httptest.NewRecorder()

	Common.Service.router.ServeHTTP(rw, req)
	log.Info(string(rw.Body.Bytes()))
	assert.Equal(t, http.StatusOK, rw.Code)
}
