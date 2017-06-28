package goutils

import (
	"encoding/json"
	"testing"
)

func TestTag(t *testing.T) {
	user := NewResource(&TestUser{})
	user.Tag(&Tag{Name: "ID", Type: "ai", Label: "自增ID"})
	user.Tag(&Tag{Name: "CreateAt", Type: "time", Label: "创建时间"})
	user.Tag(&Tag{Name: "Name", Type: "string", Label: "用户名", Memo: ""})
	user.Tag(&Tag{Name: "Email", Type: "string", Label: "邮箱地址", Memo: ""})
	user.Tag(&Tag{Name: "BirthDay", Type: "date", Label: "生日", Memo: ""})
	t.Log(user)
}

func TestTagExt(t *testing.T) {
	user := NewResource(&TestUser{})
	user.Tag(&Tag{Name: "ID", Type: "ai", Label: "自增ID"})
	user.Tag(&Tag{Name: "CreateAt", Type: "time", Label: "创建时间"})
	user.Tag(&Tag{Name: "Name", Type: "string", Label: "用户名", Memo: ""})
	user.Tag(&Tag{Name: "Email", Type: "string", Label: "邮箱地址", Memo: ""})
	user.Tag(&Tag{Name: "BirthDay", Type: "date", Label: "生日", Memo: ""})
	obj := user.NewStruct()
	t.Log(obj)
	b, _ := json.Marshal(obj)
	t.Log(string(b))
	objs := user.NewSlice()
	t.Log(objs)
	bs, _ := json.Marshal(objs)
	t.Log(string(bs))
}
