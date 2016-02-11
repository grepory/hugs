package notifier

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/opsee/hugs/obj"
	log "github.com/sirupsen/logrus"
)

func webhooktest(rw http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error(err)
	} else {
		log.Info("Test webhook endpoint got: ", string(body))
	}
}

func setupWebhookTestServer() {
	http.HandleFunc("/hook", webhooktest)
	log.Fatal(http.ListenAndServe(":8888", nil))
}

func TestWebHookNotifier(t *testing.T) {
	go setupWebhookTestServer()

	notif := &obj.Notification{
		CustomerID: "5963d7bc-6ba2-11e5-8603-6ba085b2f5b5",
		UserID:     13,
		CheckID:    "test",
		Value:      "http://localhost:8888/hook",
		Type:       "web_hook",
	}
	event := obj.GenerateTestEvent()

	webhookSender, err := NewWebHookSender()
	if err != nil {
		log.Error(err)
		t.FailNow()
	}

	err = webhookSender.Send(notif, event)
	if err != nil {
		log.Error(err)
		t.FailNow()
	}
}
