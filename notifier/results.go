package notifier

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/opsee/basic/schema"
	"github.com/opsee/pracovnik/results"
)

type ResultCache interface {
	Results(string) (*ResultCacheItem, error)
}

type resultCache struct {
	store results.Store
}

type ResultCacheItem struct {
	Targets          []*schema.Target
	Responses        []*schema.CheckResponse
	FailingResponses []*schema.CheckResponse
	PassingResponses []*schema.CheckResponse
}

func NewResultCache() *resultCache {
	return &resultCache{&results.DynamoStore{dynamodb.New(session.New(&aws.Config{Region: aws.String("us-west-2")}))}}
}

func (rs *resultCache) Results(checkId string) (*ResultCacheItem, error) {
	results, err := rs.store.GetResultsByCheckId(checkId)
	if err != nil {
		return nil, err
	}

	ci := &ResultCacheItem{}

	for _, r := range results {
		if r.Responses == nil || len(r.Responses) == 0 {
			// like what tha hell
			continue
		}

		ci.Responses = append(ci.Responses, r.Responses...)
		ci.FailingResponses = append(ci.FailingResponses, r.FailingResponses()...)
		ci.PassingResponses = append(ci.PassingResponses, r.PassingResponses()...)
		ci.Targets = append(ci.Targets, r.Targets()...)
	}

	if len(ci.Targets) == 0 || len(ci.Responses) == 0 {
		return nil, fmt.Errorf("no responses or results found in dynamo")
	}

	return ci, nil
}
