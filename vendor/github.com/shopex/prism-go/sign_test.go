package prism

import (
	"testing"
	"net/url"
	"net/http"
)


func Test_sorted_str(t *testing.T) {
	var a = make(url.Values)
	a.Add("sign", "xxx")
	a.Add("A", "a")
	a.Add("B", "b")

	if sorted_str(&a,"&", "=", "sign" ) == "A=a&B=b" {
		t.Log("test pass")
	}

}


func Test_headers_str(t *testing.T) {
	var h = make(http.Header)
	h.Add("X-Api-A", "a")


	if headers_str(&h) == "X-Api-A=a" {
		t.Log("test pass")
	}

	var h1 = make(http.Header)
	h1.Add("X-Api-A", "a")
	h1.Add("X-Api-B", "b")

	if headers_str(&h) == "X-Api-A=a&X-Api-B=b" {
		t.Log("test pass")
	}
}


func Test_escape(t *testing.T) {
	t.Log(escape("http://www.baidu.com/"))
}
