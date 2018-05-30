package prism

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

func Sign(r *http.Request, secret string) string {
	query := r.URL.Query()
	signdata := []string{
		secret,
		r.Method,
		escape(r.URL.Path),
		escape(headers_str(&r.Header)),
		escape(sorted_str(&query, "&", "=", "sign")),
		escape(sorted_str(&r.PostForm, "&", "=", "sign")),
		secret,
	}
	signstr := strings.Join(signdata, "&")

	h := md5.New()
	io.WriteString(h, signstr)

	return fmt.Sprintf("%032X", h.Sum(nil))
}

func escape(s string) string {
	return strings.Replace(url.QueryEscape(s), `+`, `%20`, -1)
}

func headers_str(sets *http.Header) string {
	mk := make([]string, len(*sets))
	i := 0
	for k, _ := range *sets {
		if k == "Authorization" || strings.Contains(k, "X-Api-") {
			mk[i] = k
			i++
		}
	}
	sort.Strings(mk)
	s := []string{}
	for _, k := range mk {
		for _, v := range (*sets)[k] {
			s = append(s, k+"="+v)
		}
	}
	return strings.Join(s, "&")
}

//sorted_str
func sorted_str(sets *url.Values, sep1 string, sep2 string, skip string) string {
	mk := make([]string, len(*sets))
	i := 0
	for k, _ := range *sets {
		mk[i] = k
		i++
	}
	sort.Strings(mk)

	s := []string{}

	for _, k := range mk {
		if k != skip {
			for _, v := range (*sets)[k] {
				s = append(s, k+sep2+v)
			}
		}
	}
	return strings.Join(s, sep1)
}

