package api

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSuccessResponse(t *testing.T) {
	payload := map[string]string{"foo": "bar"}
	msg := "Operation successful"

	resp := NewSuccessResponse(payload, msg)

	assert.NotNil(t, resp)
	assert.False(t, resp.Body.IsInvalid)
	assert.Equal(t, msg, resp.Body.SummaryMessage)
	assert.Equal(t, payload, resp.Body.Payload)
	assert.Nil(t, resp.Body.Error)
	assert.Nil(t, resp.Body.InvalidList)
}

func TestNewInvalidResponse(t *testing.T) {
	msg := "Validation failed"
	details := InvalidItem{
		"field1": "required",
	}

	// T can be anything, e.g. struct{} or map
	resp := NewInvalidResponse[any](msg, details)

	assert.NotNil(t, resp)
	assert.True(t, resp.Body.IsInvalid)
	assert.Equal(t, msg, resp.Body.SummaryMessage)
	assert.Equal(t, details, resp.Body.InvalidList)
	assert.Nil(t, resp.Body.Error)
}

func TestNewErrorResponse(t *testing.T) {
	err := errors.New("business logic error")

	resp := NewErrorResponse[any](err)

	assert.NotNil(t, resp)
	assert.True(t, resp.Body.IsInvalid)
	assert.Equal(t, err.Error(), resp.Body.SummaryMessage)
	assert.Equal(t, err, resp.Body.Error)
	assert.Nil(t, resp.Body.InvalidList)
}
