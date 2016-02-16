package checker

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnyMarshalJSON(t *testing.T) {
	check := &HttpCheck{
		Name:     "check",
		Path:     "/",
		Protocol: "http",
		Port:     1,
		Verb:     "GET",
		Headers:  []*Header{},
	}

	checkBytes, err := json.Marshal(check)
	if err != nil {
		assert.NoError(t, err)
		t.FailNow()
	}

	assert.NotEmpty(t, checkBytes)

	any, err := MarshalAny(check)
	if err != nil {
		assert.NoError(t, err)
		t.FailNow()
	}

	assert.IsType(t, &Any{}, any)

	anyBytes, err := json.Marshal(any)
	if err != nil {
		assert.NoError(t, err)
	}
	assert.NotEmpty(t, anyBytes)

	assert.Equal(t, checkBytes, anyBytes)

	origCheck := &HttpCheck{}

	if err := json.Unmarshal(anyBytes, origCheck); err != nil {
		assert.NoError(t, err)
	}
	assert.IsType(t, &HttpCheck{}, origCheck)

	assert.Equal(t, check.Name, origCheck.Name)
	assert.Equal(t, check.Path, origCheck.Path)
	assert.Equal(t, check.Protocol, origCheck.Protocol)
	assert.Equal(t, check.Port, origCheck.Port)
	assert.Equal(t, check.Verb, origCheck.Verb)
	assert.Nil(t, origCheck.Headers)
	assert.Empty(t, origCheck.Body)
}
