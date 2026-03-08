package ports

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func validateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}
	return nil
}

func validatePassword(password string) error {
	if password == "" {
		return errors.New("password is required")
	}
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	return nil
}

func validateName(name string) error {
	if name == "" {
		return errors.New("name is required")
	}
	if len(strings.TrimSpace(name)) == 0 {
		return errors.New("name cannot be whitespace only")
	}
	return nil
}

func validateTitle(title string) error {
	if title == "" {
		return errors.New("title is required")
	}
	if len(strings.TrimSpace(title)) == 0 {
		return errors.New("title cannot be whitespace only")
	}
	if len(title) > 500 {
		return errors.New("title must be less than 500 characters")
	}
	return nil
}

func validateContent(content string) error {
	if content == "" {
		return errors.New("content is required")
	}
	if len(strings.TrimSpace(content)) == 0 {
		return errors.New("content cannot be whitespace only")
	}
	return nil
}

func validationError(ctx context.Context, err error, w http.ResponseWriter, r *http.Request) {
	badRequest(ctx, err, w, r)
}
