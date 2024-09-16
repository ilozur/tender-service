package response

import (
	"errors"
	"fmt"

	validator "github.com/go-playground/validator/v10"
)

type Response struct {
	Reason string `json:"reason"`
}

func Error(msg string) Response {
	return Response{
		Reason: msg,
	}
}

func ValidationError(errs validator.ValidationErrors) string {
	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			return fmt.Sprintf("field %s is required", err.Field())
		default:
			return fmt.Sprintf("field %s is not valid", err.Field())
		}
	}
	return ""
}

var (
	ErrUserNotExists   = errors.New("username not exists")
	ErrIncorrectValue  = errors.New("incorrect value")
	ErrInternalError   = errors.New("internal error")
	ErrTenderNotExists = errors.New("tender not exists")
	ErrBidNotExists    = errors.New("bid not exists")

	ErrNoRights = errors.New("no rights for this operation")
)
