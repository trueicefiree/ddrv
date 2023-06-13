package api

import (
    "github.com/gofiber/fiber/v2"

    "github.com/forscht/ddrv/internal/ddrvfs"
)

const (
    StatusOk                  = fiber.StatusOK
    StatusRangeNotSatisfiable = fiber.StatusRequestedRangeNotSatisfiable
    StatusPartialContent      = fiber.StatusPartialContent
    StatusBadRequest          = fiber.StatusBadRequest
    StatusNotFound            = fiber.StatusNotFound
    StatusForbidden           = fiber.StatusForbidden
    StatusUnauthorized        = fiber.StatusUnauthorized
    StatusCreated             = fiber.StatusCreated
)

const (
    ErrExist               = "resource already exist"
    ErrNotFound            = "resource not found"
    ErrChangeRootDir       = "can not update/delete root dir"
    ErrBadRequest          = "bad request body"
    ErrUnauthorized        = "authorization failed"
    ErrBadUsernamePassword = "invalid username or password"
)

func GetDirHandler(fs ddrvfs.Fs) fiber.Handler {
    return func(c *fiber.Ctx) error {
        files, err := fs.GetChild(c.Params("id"))
        if err != nil {
            if err == ddrvfs.ErrNotExist {
                return fiber.NewError(StatusNotFound, err.Error())
            }
            return err
        }
        return c.JSON(Response{Message: "directory retrieved", Data: files})
    }
}

func CreateDirHandler(fs ddrvfs.Fs) fiber.Handler {
    return func(c *fiber.Ctx) error {
        file := new(ddrvfs.File)

        if err := c.BodyParser(file); err != nil {
            return fiber.NewError(StatusBadRequest, ErrBadRequest)
        }

        if err := validate.Struct(file); err != nil {
            return fiber.NewError(StatusBadRequest, err.Error())
        }
        file, err := fs.Create(file.Name, string(file.Parent), true)
        if err != nil {
            if err == ddrvfs.ErrExist || err == ddrvfs.ErrInvalidParent {
                return fiber.NewError(StatusBadRequest, err.Error())
            }
            return err
        }
        return c.Status(StatusCreated).
            JSON(Response{Message: "directory created", Data: file})
    }
}

func UpdateDirHandler(fs ddrvfs.Fs) fiber.Handler {
    return func(c *fiber.Ctx) error {
        id := c.Params("id")

        f := new(ddrvfs.File)

        if err := c.BodyParser(f); err != nil {
            return fiber.NewError(StatusBadRequest, ErrBadRequest)
        }

        if err := validate.Struct(f); err != nil {
            return fiber.NewError(StatusBadRequest, err.Error())
        }

        file, err := fs.Update(id, f.Name, string(f.Parent), true)
        if err != nil {
            if err == ddrvfs.ErrNotExist {
                return fiber.NewError(StatusNotFound, err.Error())
            }
            if err == ddrvfs.ErrExist {
                return fiber.NewError(StatusBadRequest, err.Error())
            }
            return err
        }

        return c.JSON(Response{Message: "directory updated", Data: file})
    }
}

func DelDirHandler(fs ddrvfs.Fs) fiber.Handler {
    return func(c *fiber.Ctx) error {
        id := c.Params("id")

        if err := fs.Delete(id); err != nil {
            if err == ddrvfs.ErrPermission {
                return fiber.NewError(StatusForbidden, err.Error())
            }
            if err == ddrvfs.ErrNotExist {
                return fiber.NewError(StatusNotFound, err.Error())
            }
            return err
        }
        return c.JSON(Response{Message: "directory deleted"})
    }
}
