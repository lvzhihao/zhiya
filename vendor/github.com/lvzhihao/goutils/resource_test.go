package goutils

import (
	"testing"
	"time"
)

type TestUser struct {
	Name     string
	Email    string
	BirthDay time.Time
}

func TestResource(t *testing.T) {
	res := NewResource(&TestUser{})
	t.Log(res)
}
