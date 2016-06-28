package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/opsee/basic/schema"
	"github.com/opsee/hugs/obj"
	log "github.com/opsee/logrus"
	opsee_types "github.com/opsee/protobuf/opseeproto/types"
)

type WebHookSender struct {
	resultCache ResultCache
}

type FullCheckResponse struct {
	Target   *schema.Target   `protobuf:"bytes,1,opt,name=target" json:"target,omitempty"`
	Response *json.RawMessage `protobuf:"bytes,2,opt,name=response" json:"response,omitempty"`
	Error    string           `protobuf:"bytes,3,opt,name=error" json:"error,omitempty"`
	Passing  bool             `protobuf:"varint,4,opt,name=passing" json:"passing,omitempty"`
}

type FullCheckResult struct {
	CheckId    string                 `protobuf:"bytes,1,opt,name=check_id" json:"check_id,omitempty"`
	CustomerId string                 `protobuf:"bytes,2,opt,name=customer_id" json:"customer_id,omitempty"`
	Timestamp  *opsee_types.Timestamp `protobuf:"bytes,3,opt,name=timestamp" json:"timestamp,omitempty"`
	Passing    bool                   `protobuf:"varint,4,opt,name=passing" json:"passing,omitempty"`
	Responses  []*FullCheckResponse   `protobuf:"bytes,5,rep,name=responses" json:"responses,omitempty"`
	Target     *schema.Target         `protobuf:"bytes,6,opt,name=target" json:"target,omitempty"`
	CheckName  string                 `protobuf:"bytes,7,opt,name=check_name" json:"check_name,omitempty"`
	Version    int32                  `protobuf:"varint,8,opt,name=version" json:"version,omitempty"`
}

func (this *WebHookSender) NewFullCheckResult(checkResult *schema.CheckResult) (*FullCheckResult, error) {
	fullCheckResult := &FullCheckResult{
		CheckId:    checkResult.CheckId,
		CustomerId: checkResult.CustomerId,
		Timestamp:  checkResult.Timestamp,
		Passing:    checkResult.Passing,
		Target:     checkResult.Target,
		CheckName:  checkResult.CheckName,
		Version:    checkResult.Version,
	}

	results, err := this.resultCache.Results(checkResult.CheckId)
	if err != nil {
		return nil, err
	}

	fullResponses := []*FullCheckResponse{}
	for _, response := range results.Responses {
		fullResponse := &FullCheckResponse{
			Target:  response.Target,
			Error:   response.Error,
			Passing: response.Passing,
		}

		// Unmarshal protobuf
		any, err := opsee_types.UnmarshalAny(response.Response)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		// marshal json
		// TODO(dan) marshal other any types
		// this will get replaced when we replace any with oneof
		httpResponse, ok := any.(*schema.HttpResponse)
		if !ok {
			return nil, fmt.Errorf("Couldn't assert type on any.")
		}

		anyjson, err := json.Marshal(httpResponse)
		if err != nil {
			return nil, err
		}

		rawmsg := json.RawMessage(anyjson)
		fullResponse.Response = &rawmsg

		fullResponses = append(fullResponses, fullResponse)
	}

	fullCheckResult.Responses = fullResponses

	return fullCheckResult, nil
}

// Send notification to customer.  At this point we have done basic validation on notification and event
func (this *WebHookSender) Send(n *obj.Notification, e *obj.Event) error {

	fullResult, err := this.NewFullCheckResult(e.Result)
	if err != nil {
		return err
	}

	body, err := json.Marshal(fullResult)
	if err != nil {
		return err
	}

	buff := bytes.NewBufferString(string(body))

	resp, err := http.Post(n.Value, "application/json", buff)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 300 {
		return fmt.Errorf("Remote server returned status code %d: %s", resp.StatusCode, resp.Status)
	}

	return nil
}

func NewWebHookSender(resultCache ResultCache) (*WebHookSender, error) {
	return &WebHookSender{
		resultCache: resultCache,
	}, nil
}
