package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type ValidationError struct {
	Field string
	Err   error
}

type InternalError struct {
	Msg string
}

func (e *InternalError) Error() string {
	return e.Msg
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var sb strings.Builder
	for _, err := range v {
		sb.WriteString(fmt.Sprintf("%s: %s\n", err.Field, err.Err))
	}
	return sb.String()
}

func Validate(v interface{}) error {
	val := reflect.ValueOf(v)
	switch val.Kind() { //nolint:exhaustive
	case reflect.Struct:
		return validateStruct(val)
	case reflect.Slice:
		var validationErrors ValidationErrors
		for i := 0; i < val.Len(); i++ {
			if err := Validate(val.Index(i).Interface()); err != nil {
				var nestedErrors ValidationErrors
				if errors.As(err, &nestedErrors) {
					validationErrors = append(validationErrors, nestedErrors...)
				}
			}
		}
		if len(validationErrors) > 0 {
			return validationErrors
		}
		return nil
	default:
		return &InternalError{Msg: "input is not a struct or slice"}
	}
}

func validateStruct(val reflect.Value) error {
	var validationErrors ValidationErrors

	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		tag := field.Tag.Get("validate")
		if tag == "" {
			continue
		}

		fieldValue := val.Field(i)
		if tag == "nested" {
			if err := Validate(fieldValue.Interface()); err != nil {
				var nestedErrors ValidationErrors
				if errors.As(err, &nestedErrors) {
					for _, nestedErr := range nestedErrors {
						validationErrors = append(validationErrors, ValidationError{
							Field: fmt.Sprintf("%s.%s", field.Name, nestedErr.Field),
							Err:   nestedErr.Err,
						})
					}
				} else {
					validationErrors = append(validationErrors, ValidationError{Field: field.Name, Err: err})
				}
			}
			continue
		}

		rules := strings.Split(tag, "|")
		for _, rule := range rules {
			if err := applyRule(field.Name, fieldValue, rule); err != nil {
				// Check if the error is an InternalError
				var internalErr *InternalError
				if errors.As(err, &internalErr) {
					return internalErr
				}
				validationErrors = append(validationErrors, ValidationError{Field: field.Name, Err: err})
			}
		}
	}

	if len(validationErrors) > 0 {
		return validationErrors
	}

	return nil
}

func applyRule(fieldName string, fieldValue reflect.Value, rule string) error {
	ruleParts := strings.Split(rule, ":")
	if len(ruleParts) != 2 {
		return &InternalError{Msg: fmt.Sprintf("invalid rule format: %s", rule)}
	}
	ruleName, ruleValue := ruleParts[0], ruleParts[1]

	switch fieldValue.Kind() {
	case reflect.String, reflect.Bool:
		return validateStringOrBool(fieldName, fieldValue, ruleName, ruleValue)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		return validateDigit(fieldName, fieldValue, ruleName, ruleValue)
	case reflect.Slice:
		for i := 0; i < fieldValue.Len(); i++ {
			elem := fieldValue.Index(i)
			if err := applyRule(fieldName, elem, rule); err != nil {
				return err
			}
		}
		return nil
	case reflect.Invalid, reflect.Complex64, reflect.Complex128, reflect.Array, reflect.Chan,
		reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Struct,
		reflect.UnsafePointer:
		return &InternalError{Msg: fmt.Sprintf("unsupported field type: %s", fieldValue.Kind())}
	default:
		return &InternalError{Msg: fmt.Sprintf("unsupported field type: %s", fieldValue.Kind())}
	}
}

func validateStringOrBool(fieldName string, value reflect.Value, ruleName, ruleValue string) error {
	switch ruleName {
	case "len":
		return validateLength(fieldName, value, ruleValue)
	case "regexp":
		return validateRegexp(fieldName, value, ruleValue)
	case "in":
		return validateInStringOrBool(fieldName, value, ruleValue)
	default:
		return &InternalError{Msg: fmt.Sprintf("unknown validation rule: %s", ruleName)}
	}
}

func validateLength(fieldName string, value reflect.Value, ruleValue string) error {
	if value.Kind() != reflect.String {
		return &InternalError{Msg: "len validation is only applicable to string fields"}
	}
	expectedLen, err := strconv.Atoi(ruleValue)
	if err != nil {
		return &InternalError{Msg: fmt.Sprintf("invalid length value for field %s: %s", fieldName, ruleValue)}
	}
	if len(value.String()) != expectedLen {
		return fmt.Errorf("field %s length should be %d", fieldName, expectedLen)
	}
	return nil
}

func validateRegexp(fieldName string, value reflect.Value, ruleValue string) error {
	if value.Kind() != reflect.String {
		return &InternalError{Msg: "regexp validation is only applicable to string fields"}
	}
	re, err := regexp.Compile(ruleValue)
	if err != nil {
		return &InternalError{Msg: fmt.Sprintf("invalid regexp value for field %s: %s", fieldName, ruleValue)}
	}
	if !re.MatchString(value.String()) {
		return fmt.Errorf("field %s does not match regexp %s", fieldName, ruleValue)
	}
	return nil
}

func validateInStringOrBool(fieldName string, value reflect.Value, ruleValue string) error {
	options := strings.Split(ruleValue, ",")
	switch value.Kind() {
	case reflect.String:
		for _, option := range options {
			if value.String() == option {
				return nil
			}
		}
	case reflect.Bool:
		for _, option := range options {
			boolValue, err := strconv.ParseBool(option)
			if err != nil {
				return fmt.Errorf("invalid in value for field %s: %s", fieldName, option)
			}
			if value.Bool() == boolValue {
				return nil
			}
		}
	case reflect.Invalid, reflect.Complex64, reflect.Complex128, reflect.Array, reflect.Chan,
		reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64, reflect.Struct:
		return fmt.Errorf("in validation is only applicable to string and bool fields")
	default:
		panic("unhandled default case")
	}
	return fmt.Errorf("field %s value is not in the set %v", fieldName, options)
}

func validateDigit(fieldName string, value reflect.Value, ruleName, ruleValue string) error {
	switch ruleName {
	case "min":
		return validateMinDigit(fieldName, value, ruleValue)
	case "max":
		return validateMaxDigit(fieldName, value, ruleValue)
	case "in":
		return validateInDigit(fieldName, value, ruleValue)
	default:
		return fmt.Errorf("unknown validation rule: %s", ruleName)
	}
}

func validateMinDigit(fieldName string, value reflect.Value, ruleValue string) error {
	minValue, err := strconv.ParseFloat(ruleValue, 64)
	if err != nil {
		return fmt.Errorf("invalid min value for field %s: %s", fieldName, ruleValue)
	}
	if getNumericValue(value) < minValue {
		if minValue == float64(int(minValue)) {
			return fmt.Errorf("field %s should be at least %d", fieldName, int(minValue))
		}
		return fmt.Errorf("field %s should be at least %.6f", fieldName, minValue)
	}
	return nil
}

func validateMaxDigit(fieldName string, value reflect.Value, ruleValue string) error {
	maxValue, err := strconv.ParseFloat(ruleValue, 64)
	if err != nil {
		return fmt.Errorf("invalid max value for field %s: %s", fieldName, ruleValue)
	}
	if getNumericValue(value) > maxValue {
		if maxValue == float64(int(maxValue)) {
			return fmt.Errorf("field %s should be at most %d", fieldName, int(maxValue))
		}
		return fmt.Errorf("field %s should be at most %.6f", fieldName, maxValue)
	}
	return nil
}

func validateInDigit(fieldName string, value reflect.Value, ruleValue string) error {
	options := strings.Split(ruleValue, ",")
	floatOptions := make([]float64, len(options))
	for i, option := range options {
		optionValue, err := strconv.ParseFloat(option, 64)
		if err != nil {
			return fmt.Errorf("invalid in value for field %s: %s", fieldName, option)
		}
		floatOptions[i] = optionValue
	}
	for _, optionValue := range floatOptions {
		if getNumericValue(value) == optionValue {
			return nil
		}
	}
	return fmt.Errorf("field %s value is not in the set %v", fieldName, floatOptions)
}

func getNumericValue(value reflect.Value) float64 {
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(value.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(value.Uint())
	case reflect.Float32, reflect.Float64:
		return value.Float()
	case reflect.Bool, reflect.Complex64, reflect.Complex128, reflect.Array, reflect.Chan,
		reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice, reflect.String,
		reflect.Struct, reflect.UnsafePointer, reflect.Invalid:
		return 0
	default:
		return 0
	}
}
