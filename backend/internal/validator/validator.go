package validator

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/moedahpos/backend/internal/dto"
)

// Validator wraps go-playground/validator with helper methods.
type Validator struct {
	v *validator.Validate
}

// New creates a configured Validator instance.
func New() *Validator {
	v := validator.New()

	// Use JSON field names in error messages instead of Go struct field names.
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{v: v}
}

// ValidateStruct validates a struct and returns a slice of FieldError, or nil.
func (vl *Validator) ValidateStruct(s interface{}) []dto.FieldError {
	if err := vl.v.Struct(s); err != nil {
		var errs []dto.FieldError
		for _, e := range err.(validator.ValidationErrors) {
			errs = append(errs, dto.FieldError{
				Field:   e.Field(),
				Message: humanizeTag(e),
			})
		}
		return errs
	}
	return nil
}

func humanizeTag(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return e.Field() + " is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return e.Field() + " must be at least " + e.Param() + " characters"
	case "max":
		return e.Field() + " must be at most " + e.Param() + " characters"
	default:
		return e.Field() + " is invalid"
	}
}
