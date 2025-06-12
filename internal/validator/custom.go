package validators

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
)

// ValidateTrimmedMin validates minimum length after trimming whitespace
func ValidateTrimmedMin(fl validator.FieldLevel) bool {
	value := fl.Field().String()

	trimmedValue := strings.TrimSpace(value)

	// Get the parameter (minimum length)
	param := fl.Param()
	min, err := strconv.Atoi(param)
	if err != nil {
		return false
	}

	length := utf8.RuneCountInString(trimmedValue)

	// Validate minimum length
	return length >= min
}

func RegisterCustomValidators(v *validator.Validate) {
	// Register the trimmed_min validator
	v.RegisterValidation("trimmed_min", ValidateTrimmedMin)
}
