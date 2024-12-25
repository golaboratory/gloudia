package io

import (
	"encoding/base64"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestFile_ToBase64(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "testfile")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write some data to the temporary file
	_, err = tmpFile.WriteString("test data")
	assert.NoError(t, err)
	tmpFile.Close()

	file := File{Path: tmpFile.Name()}
	encoded, err := file.ToBase64()
	assert.NoError(t, err)

	expected := base64.StdEncoding.EncodeToString([]byte("test data"))
	assert.Equal(t, expected, encoded)
}

func TestBase64_ToFile(t *testing.T) {
	// Base64 encoded data
	data := base64.StdEncoding.EncodeToString([]byte("test data"))
	b64 := Base64{Data: data}

	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "testfile")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	err = b64.ToFile(tmpFile.Name())
	assert.NoError(t, err)

	// Read the data back from the file
	content, err := os.ReadFile(tmpFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, "test data", string(content))
}
