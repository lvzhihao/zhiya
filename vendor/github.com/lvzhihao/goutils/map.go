package goutils

import (
	"encoding/json"
)

type Map struct {
	Value interface{}
}

func NewMap(value interface{}) *Map {
	m := &Map{}
	m.Load(value)
	return m
}

func (c *Map) Load(value interface{}) {
	c.Value = value
}

func (c *Map) get(key interface{}) (ret interface{}, ok bool) {
	switch c.Value.(type) {
	case map[string]interface{}:
		ret, ok = c.Value.(map[string]interface{})[ToString(key)]
	default:
		//todo
		ret = nil
		ok = false
	}
	return
}

func (c *Map) Get(key interface{}) (ret interface{}, ok bool) {
	ret, ok = c.get(key)
	return
}

func (c *Map) GetString(key interface{}) (string, bool) {
	ret, ok := c.get(key)
	if ok {
		return ToString(ret), true
	} else {
		return "", false
	}
}

func (c *Map) String(key interface{}, def ...interface{}) string {
	ret, ok := c.GetString(key)
	if !ok && len(def) > 0 {
		return ToString(def[0])
	} else {
		return ret
	}
}

func (c *Map) GetStringSlice(key interface{}) ([]string, bool) {
	ret, ok := c.get(key)
	data := make([]string, 0)
	if ok {
		err := json.Unmarshal([]byte(ToString(ret)), &data)
		if err == nil {
			return data, true
		}
	}
	return data, false
}

func (c *Map) GetSlice(key interface{}) ([]interface{}, bool) {
	ret, ok := c.get(key)
	data := make([]interface{}, 0)
	if ok {
		err := json.Unmarshal([]byte(ToString(ret)), &data)
		if err == nil {
			return data, true
		}
	}
	return data, false
}

func (c *Map) GetInt64(key interface{}) (int64, bool) {
	ret, ok := c.get(key)
	if ok {
		return ToInt64(ret), true
	} else {
		return 0, false
	}
}

func (c *Map) GetInt32(key interface{}) (int32, bool) {
	ret, ok := c.get(key)
	if ok {
		return ToInt32(ret), true
	} else {
		return 0, false
	}
}

func (c *Map) Int32(key interface{}, def ...interface{}) int32 {
	ret, ok := c.GetInt32(key)
	if !ok && len(def) > 0 {
		return ToInt32(def[0])
	} else {
		return ret
	}
}

func (c *Map) GetFloat(key interface{}) (float64, bool) {
	ret, ok := c.get(key)
	if ok {
		return ToFloat(ret), true
	} else {
		return 0, false
	}
}

func (c *Map) GetBool(key interface{}) (bool, bool) {
	ret, ok := c.get(key)
	if ok {
		return ToBool(ret), true
	} else {
		return false, false
	}
}
