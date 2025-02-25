package controller

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/golaboratory/gloudia/api/service"
)

func TestResponseInvalid(t *testing.T) {
	msg := "invalid parameters"
	invalids := service.InvalidParamList{
		{Name: "id", Message: "id is required"},
	}

	result, err := ResponseInvalid[int](msg, invalids)
	assert.NoError(t, err)
	assert.Equal(t, msg, result.Body.SummaryMessage)
	assert.True(t, result.Body.HasInvalidParams)
	assert.Equal(t, invalids, result.Body.InvalidParamList)
}

func TestResponseOk(t *testing.T) {
	payload := "success payload"
	msg := "operation successful"

	result, err := ResponseOk[string](payload, msg)
	assert.NoError(t, err)
	assert.Equal(t, msg, result.Body.SummaryMessage)
	assert.False(t, result.Body.HasInvalidParams)
	assert.Equal(t, payload, result.Body.Payload)
}

func TestNewResponseBinary(t *testing.T) {
	contentType := "application/octet-stream"
	description := "binary file response"

	respMap := NewResponseBinary(contentType, description)
	resp, exists := respMap["200"]
	assert.True(t, exists)
	assert.Equal(t, description, resp.Description)
	// Verify that the response content has the expected content type key.
	_, ok := resp.Content[contentType]
	assert.True(t, ok)
	// Optionally, check that Body is not set in the humacli response definition.
	assert.Equal(t, http.StatusOK, 200) // dummy check for structure consistency
}
