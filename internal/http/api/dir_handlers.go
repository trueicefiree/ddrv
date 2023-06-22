package api

import (
	"github.com/gofiber/fiber/v2"

	dataprovider2 "github.com/forscht/ddrv/internal/dataprovider"
)

func GetDirHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		files, err := dataprovider2.GetChild(c.Params("id"))
		if err != nil {
			if err == dataprovider2.ErrNotExist {
				return fiber.NewError(StatusNotFound, err.Error())
			}
			return err
		}
		return c.JSON(Response{Message: "directory retrieved", Data: files})
	}
}

func CreateDirHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		file := new(dataprovider2.File)

		if err := c.BodyParser(file); err != nil {
			return fiber.NewError(StatusBadRequest, ErrBadRequest)
		}

		if err := validate.Struct(file); err != nil {
			return fiber.NewError(StatusBadRequest, err.Error())
		}

		file, err := dataprovider2.Create(file.Name, string(file.Parent), true)
		if err != nil {
			if err == dataprovider2.ErrExist || err == dataprovider2.ErrInvalidParent {
				return fiber.NewError(StatusBadRequest, err.Error())
			}
			return err
		}
		return c.Status(StatusCreated).
			JSON(Response{Message: "directory created", Data: file})
	}
}

func UpdateDirHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		dir := new(dataprovider2.File)

		if err := c.BodyParser(dir); err != nil {
			return fiber.NewError(StatusBadRequest, ErrBadRequest)
		}

		if err := validate.Struct(dir); err != nil {
			return fiber.NewError(StatusBadRequest, err.Error())
		}

		dir, err := dataprovider2.Update(id, "", dir)
		if err != nil {
			if err == dataprovider2.ErrNotExist {
				return fiber.NewError(StatusNotFound, err.Error())
			}
			if err == dataprovider2.ErrExist {
				return fiber.NewError(StatusBadRequest, err.Error())
			}
			return err
		}

		return c.JSON(Response{Message: "directory updated", Data: dir})
	}
}

func DelDirHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		if err := dataprovider2.Delete(id, ""); err != nil {
			if err == dataprovider2.ErrPermission {
				return fiber.NewError(StatusForbidden, err.Error())
			}
			if err == dataprovider2.ErrNotExist {
				return fiber.NewError(StatusNotFound, err.Error())
			}
			return err
		}
		return c.JSON(Response{Message: "directory deleted"})
	}
}
