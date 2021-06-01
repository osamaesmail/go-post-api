package validation

import (
	"github.com/go-playground/validator"
	"github.com/osamaesmail/go-post-api/internal/constant"
)

func Struct(s interface{}) error {
	err := validator.New().Struct(s)
	if err != nil {
		errs := err.(validator.ValidationErrors)
		for _, e := range errs {
			return constant.NewErrFieldValidation(e)
		}
	}
	return nil
}
