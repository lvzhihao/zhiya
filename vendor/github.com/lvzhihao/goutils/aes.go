package goutils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"strings"
)

func AddBase64Padding(value string) string {
	m := len(value) % 4
	if m != 0 {
		value += strings.Repeat("=", 4-m)
	}

	return value
}

func RemoveBase64Padding(value string) string {
	return strings.Replace(value, "=", "", -1)
}

func PKCS5PaddingPad(src []byte) []byte {
	padding := aes.BlockSize - len(src)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func PKCS5PaddingUnpad(src []byte) ([]byte, error) {
	length := len(src)
	unpadding := int(src[length-1])

	if unpadding > length {
		return nil, errors.New("unpad error. This could happen when incorrect encryption key is used")
	}

	return src[:(length - unpadding)], nil
}

func AesCBCEncrypt(key []byte, text string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	msg := PKCS5PaddingPad([]byte(text))
	ciphertext := make([]byte, aes.BlockSize+len(msg))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	cbc := cipher.NewCBCEncrypter(block, iv)
	cbc.CryptBlocks(ciphertext[aes.BlockSize:], msg)
	finalMsg := RemoveBase64Padding(base64.URLEncoding.EncodeToString(ciphertext))
	return finalMsg, nil
}

func AesCBCDecrypt(key []byte, text string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	decodedMsg, err := base64.URLEncoding.DecodeString(AddBase64Padding(text))
	if err != nil {
		return "", err
	}

	if (len(decodedMsg) % aes.BlockSize) != 0 {
		return "", errors.New("blocksize must be multipe of decoded message length")
	}

	iv := decodedMsg[:aes.BlockSize]
	msg := decodedMsg[aes.BlockSize:]

	cbc := cipher.NewCBCDecrypter(block, iv)
	cbc.CryptBlocks(msg, msg)

	unpadMsg, err := PKCS5PaddingUnpad(msg)
	if err != nil {
		return "", err
	}

	return string(unpadMsg), nil
}
