package sqsconsumer

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/opsee/basic/com"
	"github.com/opsee/hugs/checker"
	"github.com/opsee/hugs/config"
	"github.com/opsee/hugs/obj"
	log "github.com/sirupsen/logrus"
)

func buildEvent(n *obj.Notification, result *checker.CheckResult) *obj.Event {
	event := &obj.Event{
		Result: result,
	}

	// If we can't finish the notificaption stuff, then we still want to return
	// an event with the CheckResult, because workers should be able to handle
	// not having the Notificaption data.
	if notifEndpoint := config.GetConfig().NotificaptionEndpoint; notifEndpoint != "" {
		resp, err := getNocapResponse(n, result)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("Error getting Notificaption data")
		} else {
			event.Nocap = resp
		}
	}

	return event
}

func getNocapResponse(n *obj.Notification, result *checker.CheckResult) (*obj.NocapResponse, error) {
	if notifEndpoint := config.GetConfig().NotificaptionEndpoint; notifEndpoint != "" {
		user := &com.User{
			ID:         n.UserID,
			CustomerID: n.CustomerID,
			Verified:   true,
			Active:     true,
		}
		uBytes, err := json.Marshal(user)
		if err != nil {
			return nil, err
		}
		base64User := base64.StdEncoding.EncodeToString(uBytes)

		checkPath := strings.Join([]string{
			config.GetConfig().BartnetEndpoint,
			"checks",
			result.CheckId,
		}, "/")

		// shared http client comes from notifier.go
		req, err := http.NewRequest("GET", checkPath, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Add("Authorization", fmt.Sprintf("Basic %s", base64User))

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		checkBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		check := &checker.Check{}
		if err := json.Unmarshal(checkBytes, check); err != nil {
			return nil, err
		}

		body := bytes.NewBuffer(checkBytes)

		req, err = http.NewRequest(
			"POST",
			strings.Join([]string{
				notifEndpoint,
				"screenshot",
			}, "/"),
			body)
		if err != nil {
			return nil, err
		}

		resp, err = httpClient.Do(req)
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

		return notifData, nil
	}
	return nil, errors.New("No notificaption endpoint configured")
}
