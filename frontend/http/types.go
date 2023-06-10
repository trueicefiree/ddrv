package http

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	ErrBadId         = "bad resource id"
	ErrExist         = "resource already exist"
	ErrBadName       = "invalid resource name"
	ErrNotFound      = "resource not found"
	ErrChangeRootDir = "can not update/delete root dir"
	ErrBadRequest    = "bad request body"
)

const (
	StatusOk                  = fiber.StatusOK
	StatusRangeNotSatisfiable = fiber.StatusRequestedRangeNotSatisfiable
	StatusPartialContent      = fiber.StatusPartialContent
	StatusBadRequest          = fiber.StatusBadRequest
	StatusNotFound            = fiber.StatusNotFound
	StatusForbidden           = fiber.StatusForbidden
)

type File struct {
	ID     string     `json:"id"`
	Name   string     `json:"name,omitempty" validate:"required"`
	Dir    bool       `json:"dir"`
	Size   int        `json:"size,omitempty"`
	Parent NullString `json:"parent,omitempty" validate:"required,uuid"`
	MTime  time.Time  `json:"mtime"`
}

type Node struct {
	URL  string
	Size int
}

type Response struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type NullString string

func (ns *NullString) Scan(value interface{}) error {
	if value == nil {
		*ns = ""
	} else if v, ok := value.([]byte); ok {
		*ns = NullString(v)
	} else if v, ok := value.(string); ok {
		*ns = NullString(v)
	} else {
		return fmt.Errorf("cannot convert %v of type %T to NullString", value, value)
	}
	return nil
}

func (ns *NullString) Value() (driver.Value, error) {
	if *ns == "" {
		return nil, nil
	}
	return string(*ns), nil
}
