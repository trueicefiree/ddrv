package http

import (
	"errors"
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var fileNameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

var validate = validator.New()

func ValidateDir(dir *File) error {
	if err := validate.Struct(dir); err != nil {
		return err
	}
	if !fileNameRegex.MatchString(dir.Name) {
		return errors.New(ErrBadName)
	}
	return nil
}

func validateParam(paramNames ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		for _, paramName := range paramNames {
			paramValue := c.Params(paramName)
			if err := validate.Var(paramValue, "uuid"); err != nil {
				return fiber.NewError(StatusBadRequest, ErrBadId)
			}
		}
		return c.Next()
	}
}
