package repository

import (
	"fmt"

	"auth-service/models"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

// ValidateUserCreate validates a user for creation
func ValidateUserCreate(u *models.User) error {
	err := validation.ValidateStruct(u,
		validation.Field(&u.Name, validation.Required, validation.Length(1, 100)),
		validation.Field(&u.Email, validation.Required, validation.Length(1, 100), is.Email),
		validation.Field(&u.Password, validation.Required, validation.Length(6, 255)),
	)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	return nil
}
