package security

import (
	"bytes"
	"crypto/aes"
	"testing"
)

func TestCryptography_EncryptString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"normal text", "Hello, World!", false},
		{"empty string", "", false},
		{"japanese text", "こんにちは世界", false},
		{"long text", "This is a very long text that needs padding", false},
	}

	crypt := &Cryptography{
		KeyText:  keyText,
		commonIV: commonIV,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := crypt.EncryptString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncryptString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if encrypted == "" {
				t.Error("EncryptString() returned empty string")
			}
		})
	}
}

func TestCryptography_DecryptString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"normal text", "Hello, World!", false},
		{"empty string", "", false},
		{"japanese text", "こんにちは世界", false},
		{"long text", "This is a very long text that needs padding", false},
	}

	crypt := &Cryptography{
		KeyText:  keyText,
		commonIV: commonIV,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := crypt.EncryptString(tt.input)
			if err != nil {
				t.Fatalf("EncryptString() failed: %v", err)
			}

			decrypted, err := crypt.DecryptString(encrypted)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecryptString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if decrypted != tt.input {
				t.Errorf("DecryptString() = %v, want %v", decrypted, tt.input)
			}
		})
	}
}

func TestPKCS7Padding(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"one block", []byte("1234567890123456")},
		{"partial block", []byte("12345")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			padded := pkcs7Pad(tt.data)
			if len(padded)%aes.BlockSize != 0 {
				t.Errorf("pkcs7Pad() result length %d is not multiple of block size", len(padded))
			}

			unpadded := pkcs7Unpad(padded)
			if !bytes.Equal(unpadded, tt.data) {
				t.Errorf("pkcs7Unpad(pkcs7Pad()) = %v, want %v", unpadded, tt.data)
			}
		})
	}
}
