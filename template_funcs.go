package kubegen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	kapi "k8s.io/kubernetes/pkg/api"
)

var Funcs = template.FuncMap{
	"closest":       arrayClosest,
	"coalesce":      coalesce,
	"combine":       combine,
	"dir":           dirList,
	"exists":        exists,
	"first":         first,
	"groupBy":       groupBy,
	"groupByKeys":   groupByKeys,
	"groupByMulti":  groupByMulti,
	"hasPrefix":     strings.HasPrefix,
	"hasSuffix":     strings.HasSuffix,
	"hasField":      hasField,
	"intersect":     intersect,
	"isPodReady":    isPodReady,
	"json":          marshalJson,
	"pathJoin":      filepath.Join,
	"keys":          keys,
	"last":          last,
	"dict":          dict,
	"mapContains":   mapContains,
	"parseBool":     strconv.ParseBool,
	"parseJson":     unmarshalJson,
	"replace":       strings.Replace,
	"shell":         execShell,
	"split":         strings.Split,
	"splitN":        strings.SplitN,
	"strContains":   strings.Contains,
	"trim":          strings.TrimSpace,
	"trimPrefix":    strings.TrimPrefix,
	"trimSuffix":    strings.TrimSuffix,
	"values":        values,
	"when":          when,
	"where":         where,
	"whereExist":    whereExist,
	"whereNotExist": whereNotExist,
	"whereAny":      whereAny,
	"whereAll":      whereAll,
}

// combine multiple slices into a single slice
func combine(slices ...interface{}) ([]interface{}, error) {
	var cnt int
	for _, s := range slices {
		val := reflect.ValueOf(s)
		if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
			return nil, fmt.Errorf("combine can only be called with slice types. received: %v", val.Kind())
		}
		cnt += val.Len()
	}
	ret := make([]interface{}, 0, cnt)
	for _, s := range slices {
		val := reflect.ValueOf(s)
		for i := 0; i < val.Len(); i++ {
			ret = append(ret, val.Index(i).Interface())
		}
	}
	return ret, nil
}

// returns bool indicating whether the provided value contains the specified field
func hasField(input interface{}, field string) bool {
	return deepGet(input, field) != nil
}

func values(input interface{}) (interface{}, error) {
	if input == nil {
		return nil, nil
	}

	val := reflect.ValueOf(input)
	if val.Kind() != reflect.Map {
		return nil, fmt.Errorf("Cannot call values on a non-map value: %v", input)
	}

	keys := val.MapKeys()
	vals := make([]interface{}, val.Len())
	for i := range keys {
		vals[i] = val.MapIndex(keys[i]).Interface()
	}

	return vals, nil
}

func marshalJson(input interface{}) (string, error) {
	if b, err := json.Marshal(input); err != nil {
		return "", err
	} else {
		return string(bytes.TrimRight(b, "\n")), nil
	}
}

func unmarshalJson(input string) (interface{}, error) {
	var v interface{}
	if err := json.Unmarshal([]byte(input), &v); err != nil {
		return nil, err
	}
	return v, nil
}

type ShellResult struct {
	Success bool
	Stdout  string
	Stderr  string
}

func execShell(cs string) *ShellResult {
	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)
	cmd := exec.Command("/bin/sh", "-c", cs)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	res := &ShellResult{
		Success: err == nil,
		Stdout:  stdout.String(),
		Stderr:  stderr.String(),
	}
	return res
}

func isPodReady(pod kapi.Pod) bool {
	return kapi.IsPodReady(&pod)
}
