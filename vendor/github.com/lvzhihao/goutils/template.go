package goutils

import (
	"html/template"
	"strconv"
	"strings"
)

var (
	TemplateHelpers = template.FuncMap{
		"IntSub": func(v1 interface{}, v2 interface{}) int64 {
			return IntArithmetic(v1, v2, "-")
		},
		"IntAdd": func(v1 interface{}, v2 interface{}) int64 {
			return IntArithmetic(v1, v2, "+")
		},
		"ToLower": func(s interface{}) string {
			return strings.ToLower(ToString(s))
		},
		"Price": func(v int64) interface{} {
			str := strconv.FormatInt(v, 10)
			len := len(str)
			return str[:len-2] + "." + str[len-2:]
		},
		"PriceInt": func(v int64) interface{} {
			str := strconv.FormatInt(v, 10)
			len := len(str)
			return str[:len-2]
		},
		"unescape": func(s string) interface{} {
			return template.HTML(s)
		},
	}
)

func NewTemplate(name, patter string) *template.Template {
	return template.Must(template.New(name).Funcs(TemplateHelpers).ParseGlob(patter))
}

/*
 * template helpers
 */
func IntArithmetic(v1, v2 interface{}, t string) (result int64) {
	switch t {
	case "+":
		result = ToInt(v1) + ToInt(v2)
		break
	case "-":
		result = ToInt(v1) - ToInt(v2)
		break
	}
	return
}
