package security

import (
	"bytes"
	"crypto/aes"
	"testing"
)

func TestCryptography_EncryptDecryptFlow(t *testing.T) {
	crypt := &Cryptography{
		KeyText:  keyText,
		commonIV: commonIV,
	}

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "SimpleString",
			input:   "Hello, World!",
			wantErr: false,
		},
		{
			name:    "EmptyString",
			input:   "",
			wantErr: false,
		},
		{
			name:    "JapaneseCharacters",
			input:   "こんにちは世界",
			wantErr: false,
		},
		{
			name:    "SpecialCharacters",
			input:   "!@#$%^&*()_+{}[]|\\:;\"'<>,.?/~`",
			wantErr: false,
		},
		{
			name:    "LongString",
			input:   string(bytes.Repeat([]byte("A"), 1024)),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := crypt.EncryptString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("EncryptString() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				decrypted, err := crypt.DecryptString(encrypted)
				if err != nil {
					t.Fatalf("DecryptString() error = %v", err)
				}

				if decrypted != tt.input {
					t.Errorf("DecryptString() = %v, want %v", decrypted, tt.input)
				}
			}
		})
	}
}

func TestPKCS7PaddingOperations(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "EmptyData",
			data: []byte{},
		},
		{
			name: "SingleByte",
			data: []byte{0x01},
		},
		{
			name: "BlockSizeData",
			data: bytes.Repeat([]byte{0x01}, 16),
		},
		{
			name: "MultipleBlockData",
			data: bytes.Repeat([]byte{0x01}, 32),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			padded := pkcs7Pad(tt.data)
			if len(padded)%aes.BlockSize != 0 {
				t.Errorf("pkcs7Pad() result length %v is not multiple of block size", len(padded))
			}

			unpadded := pkcs7Unpad(padded)
			if !bytes.Equal(unpadded, tt.data) {
				t.Errorf("pkcs7Unpad(pkcs7Pad()) = %v, want %v", unpadded, tt.data)
			}
		})
	}
}
