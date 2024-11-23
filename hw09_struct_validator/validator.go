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
	if val.Kind() != reflect.Struct {
		return errors.New("input is not a struct")
	}

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
				}
			}
			continue
		}

		rules := strings.Split(tag, "|")
		for _, rule := range rules {
			if err := applyRule(field.Name, fieldValue, rule); err != nil {
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
		return fmt.Errorf("invalid rule format: %s", rule)
	}
	ruleName, ruleValue := ruleParts[0], ruleParts[1]

	switch fieldValue.Kind() {
	case reflect.String, reflect.Int:
		return validateValue(fieldName, fieldValue, ruleName, ruleValue)
	case reflect.Slice:
		for i := 0; i < fieldValue.Len(); i++ {
			elem := fieldValue.Index(i)
			if err := applyRule(fieldName, elem, rule); err != nil {
				return err
			}
		}
	default:
		panic("unhandled default case")
	}

	return nil
}

func validateValue(fieldName string, value reflect.Value, ruleName, ruleValue string) error {
	switch ruleName {
	case "len":
		if value.Kind() == reflect.String {
			expectedLen, err := strconv.Atoi(ruleValue)
			if err != nil {
				return fmt.Errorf("invalid length value for field %s: %s", fieldName, ruleValue)
			}
			if len(value.String()) != expectedLen {
				return fmt.Errorf("field %s length should be %d", fieldName, expectedLen)
			}
		} else {
			return fmt.Errorf("len validation is only applicable to string fields")
		}
	case "regexp":
		if value.Kind() == reflect.String {
			re, err := regexp.Compile(ruleValue)
			if err != nil {
				return fmt.Errorf("invalid regexp value for field %s: %s", fieldName, ruleValue)
			}
			if !re.MatchString(value.String()) {
				return fmt.Errorf("field %s does not match regexp %s", fieldName, ruleValue)
			}
		} else {
			return fmt.Errorf("regexp validation is only applicable to string fields")
		}
	case "in":
		options := strings.Split(ruleValue, ",")
		switch value.Kind() {
		case reflect.String:
			for _, option := range options {
				if value.String() == option {
					return nil
				}
			}
			return fmt.Errorf("field %s value is not in the set %v", fieldName, options)
		case reflect.Int:
			intOptions := make([]int, len(options))
			for i, option := range options {
				optionValue, err := strconv.Atoi(option)
				if err != nil {
					return fmt.Errorf("invalid in value for field %s: %s", fieldName, option)
				}
				intOptions[i] = optionValue
			}
			for _, optionValue := range intOptions {
				if int(value.Int()) == optionValue {
					return nil
				}
			}
			return fmt.Errorf("field %s value is not in the set %v", fieldName, intOptions)
		default:
			panic("unhandled default case")
		}
	case "min":
		if value.Kind() == reflect.Int {
			minValue, err := strconv.Atoi(ruleValue)
			if err != nil {
				return fmt.Errorf("invalid min value for field %s: %s", fieldName, ruleValue)
			}
			if int(value.Int()) < minValue {
				return fmt.Errorf("field %s should be at least %d", fieldName, minValue)
			}
		} else {
			return fmt.Errorf("min validation is only applicable to int fields")
		}
	case "max":
		if value.Kind() == reflect.Int {
			maxValue, err := strconv.Atoi(ruleValue)
			if err != nil {
				return fmt.Errorf("invalid max value for field %s: %s", fieldName, ruleValue)
			}
			if int(value.Int()) > maxValue {
				return fmt.Errorf("field %s should be at most %d", fieldName, maxValue)
			}
		} else {
			return fmt.Errorf("max validation is only applicable to int fields")
		}
	}
	return nil
}
