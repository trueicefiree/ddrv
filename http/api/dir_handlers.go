package api

import (
    "github.com/gofiber/fiber/v2"

    "github.com/forscht/ddrv/dataprovider"
)

func GetDirHandler() fiber.Handler {
    return func(c *fiber.Ctx) error {
        files, err := dataprovider.GetChild(c.Params("id"))
        if err != nil {
            if err == dataprovider.ErrNotExist {
                return fiber.NewError(StatusNotFound, err.Error())
            }
            return err
        }
        return c.JSON(Response{Message: "directory retrieved", Data: files})
    }
}

func CreateDirHandler() fiber.Handler {
    return func(c *fiber.Ctx) error {
        file := new(dataprovider.File)

        if err := c.BodyParser(file); err != nil {
            return fiber.NewError(StatusBadRequest, ErrBadRequest)
        }

        if err := validate.Struct(file); err != nil {
            return fiber.NewError(StatusBadRequest, err.Error())
        }
        file, err := dataprovider.Create(file.Name, string(file.Parent), true)
        if err != nil {
            if err == dataprovider.ErrExist || err == dataprovider.ErrInvalidParent {
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

        dir := new(dataprovider.File)

        if err := c.BodyParser(dir); err != nil {
            return fiber.NewError(StatusBadRequest, ErrBadRequest)
        }

        if err := validate.Struct(dir); err != nil {
            return fiber.NewError(StatusBadRequest, err.Error())
        }

        dir, err := dataprovider.Update(id, "", dir)
        if err != nil {
            if err == dataprovider.ErrNotExist {
                return fiber.NewError(StatusNotFound, err.Error())
            }
            if err == dataprovider.ErrExist {
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

        if err := dataprovider.Delete(id, ""); err != nil {
            if err == dataprovider.ErrPermission {
                return fiber.NewError(StatusForbidden, err.Error())
            }
            if err == dataprovider.ErrNotExist {
                return fiber.NewError(StatusNotFound, err.Error())
            }
            return err
        }
        return c.JSON(Response{Message: "directory deleted"})
    }
}
