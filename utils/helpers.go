package utils

import (
	"encoding/base64"
	"reflect"
	"strings"

	"github.com/Moranilt/config-keeper/custom_errors"
	"github.com/Moranilt/http-utils/tiny_errors"
)

func MakePointer[T any](value T) *T {
	return &value
}

func ClearName(name string) (string, tiny_errors.ErrorHandler) {
	name = strings.ReplaceAll(name, "/", "")
	name = strings.ReplaceAll(name, "\\", "")
	name = strings.ReplaceAll(name, "'", "")
	name = strings.ReplaceAll(name, "\"", "")
	if name == "" {
		return "", tiny_errors.New(custom_errors.ERR_CODE_NotValid, tiny_errors.Message("not valid name"))
	}
	return name, nil
}

type RequiredField struct {
	Name  string
	Value any
}

func ValidateRequiredFields(data []RequiredField) []tiny_errors.ErrorOption {
	var options []tiny_errors.ErrorOption
	for _, field := range data {
		if isEmptyValue(field.Value) {
			options = append(options, tiny_errors.Detail(field.Name, "required"))
		}
	}
	return options
}

func isEmptyValue(v any) bool {
	switch value := v.(type) {
	case nil:
		return true
	case string:
		return value == ""
	case int:
		return value == 0
	case float64:
		return value == 0.0
	default:
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Ptr, reflect.Interface:
			return rv.IsNil()
		case reflect.Slice, reflect.Map:
			return rv.Len() == 0
		}
	}
	return false
}

func StringToBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}
