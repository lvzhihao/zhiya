package goutils

import (
	"testing"
	"time"

	"github.com/jinzhu/gorm"
)

type TestUser struct {
	gorm.Model
	Name     string
	Email    string
	BirthDay time.Time
}

func TestResource(t *testing.T) {
	res := NewResource(&TestUser{})
	t.Log(res)
}
