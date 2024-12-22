package security

import (
	"testing"
)

func TestEncryptString(t *testing.T) {
	data := "Hello, World!"
	encrypted, err := EncryptString(data)
	if err != nil {
		t.Errorf("EncryptString returned an error: %v", err)
	}

	if encrypted == "" {
		t.Errorf("EncryptString returned an empty string")
	}
}

func TestDecryptString(t *testing.T) {
	data := "Hello, World!"
	encrypted, err := EncryptString(data)
	if err != nil {
		t.Errorf("EncryptString returned an error: %v", err)
	}

	decrypted, err := DecryptString(encrypted)
	if err != nil {
		t.Errorf("DecryptString returned an error: %v", err)
	}

	if decrypted != data {
		t.Errorf("DecryptString = %v; want %v", decrypted, data)
	}
}

func TestEncryptDecryptString(t *testing.T) {
	data := "Hello, World!"
	encrypted, err := EncryptString(data)
	if err != nil {
		t.Errorf("EncryptString returned an error: %v", err)
	}

	decrypted, err := DecryptString(encrypted)
	if err != nil {
		t.Errorf("DecryptString returned an error: %v", err)
	}

	if decrypted != data {
		t.Errorf("DecryptString = %v; want %v", decrypted, data)
	}
}
