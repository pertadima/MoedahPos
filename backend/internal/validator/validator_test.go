package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

func TestValidator_ValidateStruct(t *testing.T) {
	v := New()

	t.Run("Valid", func(t *testing.T) {
		s := TestStruct{Name: "John", Email: "john@example.com"}
		errs := v.ValidateStruct(s)
		assert.Nil(t, errs)
	})

	t.Run("Invalid Required", func(t *testing.T) {
		s := TestStruct{Email: "john@example.com"}
		errs := v.ValidateStruct(s)
		assert.NotNil(t, errs)
		assert.Len(t, errs, 1)
		assert.Equal(t, "name", errs[0].Field)
		assert.Equal(t, "name is required", errs[0].Message)
	})

	t.Run("Invalid Email", func(t *testing.T) {
		s := TestStruct{Name: "John", Email: "invalid"}
		errs := v.ValidateStruct(s)
		assert.NotNil(t, errs)
		assert.Len(t, errs, 1)
		assert.Equal(t, "email", errs[0].Field)
		assert.Equal(t, "must be a valid email address", errs[0].Message)
	})

	t.Run("Invalid Max", func(t *testing.T) {
		type MaxStruct struct {
			Code string `json:"code" validate:"max=3"`
		}
		s := MaxStruct{Code: "LONG"}
		errs := v.ValidateStruct(s)
		assert.NotNil(t, errs)
		assert.Equal(t, "code must be at most 3 characters", errs[0].Message)
	})
}
