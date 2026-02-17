package validation

import (
	"errors"
	"net/mail"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	ErrInvalidEmail    = errors.New("invalid email")
	ErrInvalidName     = errors.New("linvalid name")
	ErrInvalidPassword = errors.New("invalid password")
)

func ValidateEmail(email string) error {
	e := strings.ToLower(strings.TrimSpace(email))
	if e == "" {
		return ErrInvalidEmail
	}
	if _, err := mail.ParseAddress(e); err != nil {
		return ErrInvalidEmail
	}
	return nil
}

func ValidatePassword(pw string) error {
	if pw == "" {
		return ErrInvalidPassword
	}

	rc := utf8.RuneCountInString(pw)
	if rc < 8 || rc > 64 {
		return ErrInvalidPassword
	}

	for _, r := range pw {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return ErrInvalidPassword
		}
	}
	return nil
}

func ValidateName(name string) error {
	n := strings.TrimSpace(name)
	if n == "" {
		return ErrInvalidName
	}
	if utf8.RuneCountInString(n) > 255 {
		return ErrInvalidName
	}
	return nil
}

func ValidateRegisterRequest(name, email, password string) error {
	if err := ValidateName(strings.TrimSpace(name)); err != nil {
		return err
	}

	if err := ValidatePassword(password); err != nil {
		return err
	}

	if err := ValidateEmail(strings.TrimSpace(email)); err != nil {
		return err
	}
	return nil
}
