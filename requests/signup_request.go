package request

import "github.com/go-playground/validator/v10"

type SignUpRequest struct {
	Name              string `json:"name" validate:"required"`
	Username          string `json:"username" validate:"required,min=3,max=20"`
	Phone             string `json:"phone" validate:"required,len=10"`
	Bio               string `json:"bio" validate:"required"`
	ImageProfile      *string `json:"image_profile"`
	Email             string `json:"email" validate:"required,email"`
	Role              string `json:"role" validate:"required,oneof=Accountant Admin Manager"`
	Password          string `json:"password" validate:"required,len=8"`
}
// Humanize errors returned by the validator.
func CustomeErrorMessage(errs validator.ValidationErrors) []string {
	var errMessages []string
	for _, e := range errs {
		// Customize the message to include the field and a user-friendly error.
		switch e.Tag() {
		case "required":
			errMessages = append(errMessages, e.Field()+" is required")
		case "min":
			errMessages = append(errMessages, e.Field()+" must be at least "+e.Param()+" characters long")
		case "max":
			errMessages = append(errMessages, e.Field()+" must be at most "+e.Param()+" characters long")
		case "email":
			errMessages = append(errMessages, e.Field()+" must be a valid email address")
		case "len":
			errMessages = append(errMessages, e.Field()+" must be "+e.Param()+" characters long")
		default:
			errMessages = append(errMessages, e.Field()+" is invalid")
		}
	}
	return errMessages
}