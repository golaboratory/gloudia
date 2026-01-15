package api

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestResponse is a struct for testing UnifiedResponseBody embedding.
type TestResponse struct {
	UnifiedResponseBody
	Payload map[string]string `json:"payload,omitempty"`
}

func TestSetSuccess(t *testing.T) {
	payload := map[string]string{"foo": "bar"}
	msg := "Operation successful"

	resp := &TestResponse{
		Payload: payload,
	}
	resp.SetSuccess(msg)

	assert.NotNil(t, resp)
	assert.False(t, resp.IsInvalid)
	assert.Equal(t, msg, resp.SummaryMessage)
	assert.Equal(t, payload, resp.Payload)
	assert.Nil(t, resp.Error)
	assert.Nil(t, resp.InvalidList)
}

func TestSetInvalid(t *testing.T) {
	msg := "Validation failed"
	details := InvalidItem{
		"field1": "required",
	}

	resp := &TestResponse{}
	resp.SetInvalid(msg, details)

	assert.NotNil(t, resp)
	assert.True(t, resp.IsInvalid)
	assert.Equal(t, msg, resp.SummaryMessage)
	assert.Equal(t, details, resp.InvalidList)
	assert.Nil(t, resp.Error)
}

func TestSetError(t *testing.T) {
	err := errors.New("business logic error")

	resp := &TestResponse{}
	resp.SetError(err)

	assert.NotNil(t, resp)
	assert.True(t, resp.IsInvalid)
	assert.Equal(t, err.Error(), resp.SummaryMessage)
	assert.Equal(t, err, resp.Error)
	assert.Nil(t, resp.InvalidList)
}
