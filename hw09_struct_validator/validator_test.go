package hw09structvalidator

import (
	"encoding/json"
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
				Phones: []string{"12345"},
			},
			expectedErr: ValidationErrors{{Field: "Phones", Err: fmt.Errorf("field Phones length should be 11")}},
		},
		{
			name: "valid app version",
			in: App{
				Version: "1.0.0",
			},
			expectedErr: nil,
		},
		{
			name: "invalid app version length",
			in: App{
				Version: "1.0",
			},
			expectedErr: ValidationErrors{{Field: "Version", Err: fmt.Errorf("field Version length should be 5")}},
		},
		{
			name: "valid response code",
			in: Response{
				Code: 200,
				Body: "OK",
			},
			expectedErr: nil,
		},
		{
			name: "invalid response code",
			in: Response{
				Code: 403,
				Body: "Forbidden",
			},
			expectedErr: ValidationErrors{{Field: "Code", Err: fmt.Errorf("field Code value is not in the set [200 404 500]")}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.in)
			if err != nil {
				require.Equal(t, tt.expectedErr, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
