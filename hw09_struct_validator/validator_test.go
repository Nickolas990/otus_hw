package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	//nolint:depguard
	"github.com/stretchr/testify/require"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^[a-zA-Z0-9._%\\+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		in          interface{}
		expectedErr error
	}{
		{
			name: "valid user",
			in: User{
				ID:     "123456789012345678901234567890123456",
				Name:   "John",
				Age:    25,
				Email:  "john.doe@example.com",
				Role:   "admin",
				Phones: []string{"12345678901"},
			},
			expectedErr: nil,
		},
		{
			name: "invalid user ID length",
			in: User{
				ID:     "short_id",
				Name:   "John",
				Age:    25,
				Email:  "john.doe@example.com",
				Role:   "admin",
				Phones: []string{"12345678901"},
			},
			expectedErr: ValidationErrors{{Field: "ID", Err: fmt.Errorf("field ID length should be 36")}},
		},
		{
			name: "invalid user age",
			in: User{
				ID:     "123456789012345678901234567890123456",
				Name:   "John",
				Age:    17,
				Email:  "john.doe@example.com",
				Role:   "admin",
				Phones: []string{"12345678901"},
			},
			expectedErr: ValidationErrors{{Field: "Age", Err: fmt.Errorf("field Age should be at least 18")}},
		},
		{
			name: "invalid user email",
			in: User{
				ID:     "123456789012345678901234567890123456",
				Name:   "John",
				Age:    25,
				Email:  "invalid-email",
				Role:   "admin",
				Phones: []string{"12345678901"},
			},
			expectedErr: ValidationErrors{{
				Field: "Email",
				Err: fmt.Errorf(
					"field Email does not match regexp ^[a-zA-Z0-9._%%\\+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"),
			}},
		},
		{
			name: "invalid user role",
			in: User{
				ID:     "123456789012345678901234567890123456",
				Name:   "John",
				Age:    25,
				Email:  "john.doe@example.com",
				Role:   "user",
				Phones: []string{"12345678901"},
			},
			expectedErr: ValidationErrors{{Field: "Role", Err: fmt.Errorf("field Role value is not in the set [admin stuff]")}},
		},
		{
			name: "invalid user phone length",
			in: User{
				ID:     "123456789012345678901234567890123456",
				Name:   "John",
				Age:    25,
				Email:  "john.doe@example.com",
				Role:   "admin",
				Phones: []string{"1234567890"}, // Invalid phone length
			},
			expectedErr: ValidationErrors{{Field: "Phones", Err: fmt.Errorf("field Phones length should be 11")}},
		},
		{
			name: "non-struct input",
			in:   123,
			expectedErr: &InternalError{
				Msg: "input is not a struct",
			},
		},
		{
			name: "unsupported field type",
			in: struct {
				ComplexField complex128 `validate:"len:5"`
			}{
				ComplexField: 1 + 2i,
			},
			expectedErr: &InternalError{
				Msg: "unsupported field type: complex128",
			},
		},
		{
			name: "invalid rule format",
			in: struct {
				Field string `validate:"invalidrule"`
			}{
				Field: "value",
			},
			expectedErr: &InternalError{
				Msg: "invalid rule format: invalidrule",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.in)
			if tc.expectedErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				var validationErrs ValidationErrors
				if errors.As(tc.expectedErr, &validationErrs) {
					require.IsType(t, validationErrs, err)
					require.Equal(t, validationErrs, err)
				}
				var internalErr *InternalError
				if errors.As(tc.expectedErr, &internalErr) {
					require.IsType(t, internalErr, err)
					require.Equal(t, internalErr.Msg, err.Error())
				}
			}
		})
	}
}
