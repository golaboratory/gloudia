package security

import (
	"bytes"
	"testing"
)

func TestEncryptString(t *testing.T) {
	data := "Hello, World!"
	encrypted, err := Encrypt(data)
	if err != nil {
		t.Errorf("EncryptString returned an error: %v", err)
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

func TestDecryptString(t *testing.T) {
	data := "Hello, World!"
	encrypted, err := Encrypt(data)
	if err != nil {
		t.Errorf("EncryptString returned an error: %v", err)
	}

	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Errorf("DecryptString returned an error: %v", err)
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

func TestEncryptDecryptString(t *testing.T) {
	data := "Hello, World!"
	encrypted, err := Encrypt(data)
	if err != nil {
		t.Errorf("EncryptString returned an error: %v", err)
	}

	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Errorf("DecryptString returned an error: %v", err)
	}

	unpadded := pkcs7Unpad(padded)
	if !bytes.Equal(unpadded, tt.data) {
		t.Errorf("pkcs7Unpad(pkcs7Pad()) = %v, want %v", unpadded, tt.data)
	}

}
