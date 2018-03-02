package goutils

import (
	"crypto/md5"
	"testing"
)

func TestAesCBCEncrypt(t *testing.T) {
	key := []byte("8e74f8e7bca79eb1150281741223870aba412ec2841718038ad3988566487")
	md5Key := md5.Sum([]byte(key))
	encryptMsg, _ := AesCBCEncrypt(md5Key[:16], `{"info":"北京到上海的飞机","key":"70aba412ec2841718038ad3988566487","userid":"testUtil"}`)
	msg, _ := AesCBCDecrypt(md5Key[:16], encryptMsg)
	t.Log(msg)
}
