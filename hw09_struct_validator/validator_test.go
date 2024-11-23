package hw09structvalidator

import (
	"encoding/json"
	"fmt"
	"testing"

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
			name: "invalid user",
			in: User{
				ID:     "123",
				Name:   "John",
				Age:    17,
				Email:  "not-an-email",
				Role:   "user",
				Phones: []string{"123"},
			},
			expectedErr: ValidationErrors{
				{Field: "ID", Err: fmt.Errorf("field ID length should be 36")},
				{Field: "Age", Err: fmt.Errorf("field Age should be at least 18")},
				{Field: "Email", Err: fmt.Errorf("field Email does not match regexp ^[a-zA-Z0-9._%%\\+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$")},
				{Field: "Role", Err: fmt.Errorf("field Role value is not in the set [admin stuff]")},
				{Field: "Phones", Err: fmt.Errorf("field Phones length should be 11")},
			},
		},
		{
			name: "valid app",
			in: App{
				Version: "12345",
			},
			expectedErr: nil,
		},
		{
			name: "invalid app",
			in: App{
				Version: "1234",
			},
			expectedErr: ValidationErrors{
				{Field: "Version", Err: fmt.Errorf("field Version length should be 5")},
			},
		},
		{
			name: "valid response",
			in: Response{
				Code: 200,
				Body: "OK",
			},
			expectedErr: nil,
		},
		{
			name: "invalid response",
			in: Response{
				Code: 201,
				Body: "Created",
			},
			expectedErr: ValidationErrors{
				{Field: "Code", Err: fmt.Errorf("field Code value is not in the set [200 404 500]")},
			},
		},
		{
			name: "nested user",
			in: struct {
				Nested User `validate:"nested"`
			}{
				Nested: User{
					ID:     "123",
					Age:    10,
					Email:  "invalid-email",
					Role:   "invalid-role",
					Phones: []string{"123"},
				},
			},
			expectedErr: ValidationErrors{
				{Field: "Nested.ID", Err: fmt.Errorf("field ID length should be 36")},
				{Field: "Nested.Age", Err: fmt.Errorf("field Age should be at least 18")},
				{Field: "Nested.Email", Err: fmt.Errorf("field Email does not match regexp ^[a-zA-Z0-9._%%\\+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$")},
				{Field: "Nested.Role", Err: fmt.Errorf("field Role value is not in the set [admin stuff]")},
				{Field: "Nested.Phones", Err: fmt.Errorf("field Phones length should be 11")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.in)
			if tt.expectedErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.IsType(t, ValidationErrors{}, err)
				validationErrors := err.(ValidationErrors)
				require.Equal(t, len(tt.expectedErr.(ValidationErrors)), len(validationErrors))
				for _, expectedErr := range tt.expectedErr.(ValidationErrors) {
					found := false
					for _, actualErr := range validationErrors {
						if expectedErr.Field == actualErr.Field && expectedErr.Err.Error() == actualErr.Err.Error() {
							found = true
							break
						}
					}
					require.True(t, found, "expected error %v not found in actual errors %v", expectedErr, validationErrors)
				}
			}
		})
	}
}
