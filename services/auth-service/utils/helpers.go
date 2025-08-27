package utils

import (
	"fmt"
	"regexp"
	"time"
)

// ValidatePassword checks if the password meets the complexity requirements:
func ValidatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString
	hasUppercase := regexp.MustCompile(`[A-Z]`).MatchString
	return hasDigit(password) && hasUppercase(password)
}

func ValidDate(value any) error {
	switch v := value.(type) {
	case time.Time:
		if v.IsZero() {
			return fmt.Errorf("date_of_birth is required and must be a valid date")
		}
	case *time.Time:
		if v == nil || v.IsZero() {
			return fmt.Errorf("date_of_birth is required and must be a valid date")
		}
	default:
		return fmt.Errorf("date_of_birth must be a valid date")
	}
	return nil
}

// ValidateAlphaNumericSpace checks if the input contains only letters, numbers, and spaces.
func ValidateAlphaNumericSpace(input string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9 ]+$`)
	return re.MatchString(input)
}
