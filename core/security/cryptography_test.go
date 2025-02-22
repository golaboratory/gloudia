package security

import (
	"bytes"
	"testing"
)

func TestCryptography_EncryptDecryptFlow(t *testing.T) {
	crypt := &Cryptography{}

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

func TestCryptography_EncryptDecrypt(t *testing.T) {
	crypt := &Cryptography{}

	tests := []struct {
		name  string
		input string
	}{
		{name: "SimpleString", input: "Hello, World!"},
		{name: "EmptyString", input: ""},
		{name: "JapaneseCharacters", input: "こんにちは世界"},
		{name: "SpecialCharacters", input: "!@#$%^&*()_+{}[]|\\:;\"'<>,.?/~`"},
		{name: "LongString", input: string(make([]byte, 1024))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := crypt.EncryptString(tt.input)
			if err != nil {
				t.Fatalf("EncryptString() error = %v", err)
			}

			decrypted, err := crypt.DecryptString(encrypted)
			if err != nil {
				t.Fatalf("DecryptString() error = %v", err)
			}

			if decrypted != tt.input {
				t.Errorf("DecryptString() = %q, want %q", decrypted, tt.input)
			}
		})
	}
}
