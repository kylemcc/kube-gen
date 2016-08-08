package kubegen

import (
	"strings"
	"text/template"
)

var Funcs = template.FuncMap{
	"hasPrefix":   strings.HasPrefix,
	"hasSuffix":   strings.HasSuffix,
	"replace":     strings.Replace,
	"split":       strings.Split,
	"splitN":      strings.SplitN,
	"strContains": strings.Contains,
	"trim":        strings.Trim,
	"trimLeft":    strings.TrimLeft,
	"trimRight":   strings.TrimRight,
	"trimPrefix":  strings.TrimPrefix,
	"trimSuffix":  strings.TrimSuffix,
}
