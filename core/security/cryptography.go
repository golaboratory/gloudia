package security

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
)

var (
	keyText  = []byte("645E739A7F9F162725C1533DC2C5E827")
	commonIV = []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}
)

// Encrypt は、与えられた文字列をAES暗号化します。
// 文字列を受け取り、暗号化された文字列とエラーを返します。
func Encrypt(dataString string) (string, error) {

	data := []byte(dataString)
	block, err := aes.NewCipher(keyText)
	if err != nil {
		return "", err
	}
	padded := pkcs7Pad(data)
	encrypted := make([]byte, len(padded))
	cbcEncrypt := cipher.NewCBCEncrypter(block, commonIV)
	cbcEncrypt.CryptBlocks(encrypted, padded)
	return string(encrypted), nil
}

// pkcs7Pad は、PKCS7パディングを使用してデータをパディングします。
// バイトスライスを受け取り、パディングされたバイトスライスを返します。
func pkcs7Pad(data []byte) []byte {
	length := aes.BlockSize - (len(data) % aes.BlockSize)
	trailing := bytes.Repeat([]byte{byte(length)}, length)
	return append(data, trailing...)
}

// Decrypt は、与えられた暗号化文字列をAES復号化します。
// 暗号化された文字列を受け取り、復号化された文字列とエラーを返します。
func Decrypt(dataString string) (string, error) {
	block, err := aes.NewCipher(keyText)
	if err != nil {
		return "", err
	}
	decrypted := make([]byte, len(dataString))
	cbcDecrypt := cipher.NewCBCDecrypter(block, commonIV)
	cbcDecrypt.CryptBlocks(decrypted, []byte(dataString))
	return string(pkcs7Unpad(decrypted)), nil
}

// pkcs7Unpad は、PKCS7パディングを削除します。
// パディングされたバイトスライスを受け取り、パディングが削除されたバイトスライスを返します。
func pkcs7Unpad(data []byte) []byte {
	dataLength := len(data)
	padLength := int(data[dataLength-1])
	return data[:dataLength-padLength]
}
