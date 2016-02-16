package checker

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
)

var (
	registry = make(map[string]reflect.Type)
)

func init() {
	registry["HttpCheck"] = reflect.TypeOf(HttpCheck{})
	registry["HttpResponse"] = reflect.TypeOf(HttpResponse{})
}

func UnmarshalAny(any *Any) (interface{}, error) {
	class := any.TypeUrl
	bytes := any.Value

	instance := reflect.New(registry[class]).Interface()
	err := proto.Unmarshal(bytes, instance.(proto.Message))
	if err != nil {
		log.WithError(err).Error("Couldn't unmarshal any: ", any.String())
		return nil, err
	}
	log.WithFields(log.Fields{"service": "checker", "event": "unmarshal successful"}).Debug("unmarshaled Any to: ", instance)

	return instance, nil
}

func MarshalAny(i interface{}) (*Any, error) {
	msg, ok := i.(proto.Message)
	if !ok {
		err := fmt.Errorf("Unable to convert to proto.Message: %v", i)
		log.WithFields(log.Fields{"service": "checker", "event": "marshalling error"}).Error(err.Error())
		return nil, err
	}
	bytes, err := proto.Marshal(msg)

	if err != nil {
		log.WithFields(log.Fields{"service": "checker", "event": "marshalling error"}).Error(err.Error())
		return nil, err
	}

	return &Any{
		TypeUrl: reflect.ValueOf(i).Elem().Type().Name(),
		Value:   bytes,
	}, nil
}

func (a *Any) MarshalJSON() ([]byte, error) {
	obj, err := UnmarshalAny(a)
	if err != nil {
		return []byte{}, err
	}

	bytes, err := json.Marshal(obj)
	if err != nil {
		return []byte{}, err
	}

	return bytes, nil
}

func (r *CheckResult) filterResponses(passing bool) []*CheckResponse {
	responses := []*CheckResponse{}
	if r != nil {
		for _, resp := range r.Responses {
			if resp.Passing == passing {
				responses = append(responses, resp)
			}
		}
	}
	return responses
}

func (r *CheckResult) countResponses(passing bool) int {
	count := 0
	if r != nil {
		for _, resp := range r.Responses {
			if resp.Passing == passing {
				count += 1
			}
		}
	}
	return count
}

func (r *CheckResult) PassingResponses() []*CheckResponse {
	return r.filterResponses(true)
}

func (r *CheckResult) FailingResponses() []*CheckResponse {
	return r.filterResponses(false)
}

func (r *CheckResult) FailingCount() int {
	return r.countResponses(false)
}

func (r *CheckResult) PassingCount() int {
	return r.countResponses(true)
}
