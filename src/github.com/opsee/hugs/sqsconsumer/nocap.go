package sqsconsumer

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/opsee/hugs/checker"
	"github.com/opsee/hugs/config"
	"github.com/opsee/hugs/obj"
	log "github.com/sirupsen/logrus"
)

func buildEvent(n *obj.Notification, result *checker.CheckResult) *obj.Event {
	log.WithFields(log.Fields{"notification": n}).Info("Building event.")

	event := &obj.Event{
		Result: result,
	}

	// If we can't finish the notificaption stuff, then we still want to return
	// an event with the CheckResult, because workers should be able to handle
	// not having the Notificaption data.
	if notifEndpoint := config.GetConfig().NotificaptionEndpoint; notifEndpoint != "" {
		resp, err := getNocapResponse(notifEndpoint, result)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("Error getting Notificaption data")
		} else {
			event.Nocap = resp
			log.WithFields(log.Fields{"nocap": resp}).Info("Got nocap response")
		}
	}

	return event
}

func getNocapResponse(nocapEndpoint string, result *checker.CheckResult) (*obj.NocapResponse, error) {
	checkBytes, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	body := bytes.NewBuffer(checkBytes)

	req, err := http.NewRequest(
		"POST",
		strings.Join([]string{
			nocapEndpoint,
			"screenshot",
		}, "/"),
		body)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	notifData := &obj.NocapResponse{}
	if err := json.Unmarshal(bodyBytes, notifData); err != nil {
		return nil, err
	}
	log.WithFields(log.Fields{"response": notifData}).Info("Got response from Notificaption")

	return notifData, nil
}
