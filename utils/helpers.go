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
	options := []tiny_errors.ErrorOption{}
	for _, field := range data {
		switch field.Value.(type) {
		case nil:
			options = append(options, tiny_errors.Detail(field.Name, "required"))
		case string:
			if field.Value == "" {
				options = append(options, tiny_errors.Detail(field.Name, "required"))
			}
		case int:
			if field.Value == 0 {
				options = append(options, tiny_errors.Detail(field.Name, "required"))
			}
		case float64:
			if field.Value == 0.0 {
				options = append(options, tiny_errors.Detail(field.Name, "required"))
			}
		default:
			anyType := reflect.ValueOf(field.Value)
			switch anyType.Kind() {
			case reflect.Ptr:
				if anyType.IsNil() {
					options = append(options, tiny_errors.Detail(field.Name, "required"))
				}
			default:
				continue
			}
		}
	}

	if len(options) > 0 {
		return options
	}
	return nil
}

func StringToBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}
