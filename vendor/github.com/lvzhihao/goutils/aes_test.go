package goutils

import (
	"testing"
)

func TestAesCBCEncrypt(t *testing.T) {
	key := []byte("1234567890123456")
	encryptMsg, _ := AesCBCEncrypt(key, "Hello World")
	msg, _ := AesCBCDecrypt(key, encryptMsg)
	t.Log(msg)
}
